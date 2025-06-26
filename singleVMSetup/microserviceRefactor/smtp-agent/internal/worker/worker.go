package worker

import (
	"encoding/json"
	"sync"
	"time"

	"smtp-worker/domain"
	"smtp-worker/infrastructure/logging"
	"smtp-worker/infrastructure/smtp"

	"github.com/nats-io/nats.go"
)

const (
	JobsStream    = "EMAIL_JOPcS"
	JobsSubject   = "email.jobs.new"
	PullBatchSize = 10
	PullMaxWait   = 10 * time.Second
	RetryDelay    = 5 * time.Second
	WorkerCount   = 5
)

type Worker struct {
	subscription *nats.Subscription
	smtpClient   *smtp.Client
	logger       logging.Logger
	jobChan      chan *nats.Msg
	wg           sync.WaitGroup
	shutdownChan chan struct{}
}

func New(sub *nats.Subscription, smtpClient *smtp.Client, logger logging.Logger) *Worker {
	return &Worker{
		subscription: sub,
		smtpClient:   smtpClient,
		logger:       logger,
		jobChan:      make(chan *nats.Msg, 100),
		shutdownChan: make(chan struct{}),
	}
}

func (w *Worker) Run() error {

	for i := 0; i < WorkerCount; i++ {
		w.wg.Add(1)
		go w.worker()
	}

	go w.fetchMessages()

	<-w.shutdownChan
	close(w.jobChan)
	w.wg.Wait()
	return nil
}

func (w *Worker) Shutdown() {
	close(w.shutdownChan)
}

func (w *Worker) fetchMessages() {
	for {
		select {
		case <-w.shutdownChan:
			return
		default:
			msgs, err := w.subscription.Fetch(PullBatchSize, nats.MaxWait(PullMaxWait))
			if err != nil {
				if err == nats.ErrTimeout {
					continue
				}
				w.logger.Printf("Fehler beim Abrufen von Nachrichten aus NATS: %v", err)
				time.Sleep(RetryDelay)
				continue
			}

			w.logger.Printf("%d neue E-Mail-Aufträge aus der Queue geholt", len(msgs))

			for _, msg := range msgs {
				select {
				case w.jobChan <- msg:
				case <-w.shutdownChan:
					return
				}
			}
		}
	}
}

func (w *Worker) worker() {
	defer w.wg.Done()

	for msg := range w.jobChan {
		w.processMessage(msg)
	}
}

func (w *Worker) processMessage(msg *nats.Msg) {
	var email domain.Email

	if err := json.Unmarshal(msg.Data, &email); err != nil {
		w.logger.Printf("FEHLER: Ungültige Nachricht: %v. Nachricht wird verworfen.", err)
		msg.Ack()
		return
	}

	msgID := email.Headers["Message-ID"]
	w.logger.Printf("Verarbeite E-Mail (Message-ID: %s)", msgID)

	_, err := w.smtpClient.Send(&email)
	if err != nil {
		w.logger.Printf("FEHLER beim Senden (Message-ID: %s): %v", msgID, err)
		return
	}

	if err := msg.Ack(); err != nil {
		w.logger.Printf("FEHLER: ACK für (Message-ID: %s): %v", msgID, err)
	} else {
		w.logger.Printf("E-Mail (Message-ID: %s) erfolgreich verarbeitet", msgID)
	}
}
