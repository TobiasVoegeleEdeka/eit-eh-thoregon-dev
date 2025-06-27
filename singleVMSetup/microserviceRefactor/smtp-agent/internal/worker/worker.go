package worker

import (
	"encoding/json"
	"sync"
	"time"

	"smtp-worker/domain"
	"smtp-worker/infrastructure/config"
	"smtp-worker/infrastructure/logging"

	"github.com/nats-io/nats.go"
)

const (
	JobsStream  = "EMAIL_JOPS"
	JobsSubject = "email.jobs.new"
)

// NatsMsgAdapter adapts *nats.Msg to our NatsMsger interface
type NatsMsgAdapter struct {
	*nats.Msg
}

func (n NatsMsgAdapter) Data() []byte {
	return n.Msg.Data
}

type Worker struct {
	subscription *nats.Subscription
	smtpClient   SMTPClienter
	logger       logging.Logger
	config       *config.WorkerConfig
	jobChan      chan *nats.Msg
	wg           sync.WaitGroup
	shutdownChan chan struct{}
}

func New(sub *nats.Subscription, smtpClient SMTPClienter, logger logging.Logger, cfg *config.WorkerConfig) *Worker {
	return &Worker{
		subscription: sub,
		smtpClient:   smtpClient,
		logger:       logger,
		config:       cfg,
		jobChan:      make(chan *nats.Msg, cfg.PullBatchSize),
		shutdownChan: make(chan struct{}),
	}
}

func (w *Worker) Run() error {
	for i := 0; i < w.config.WorkerCount; i++ {
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
			msgs, err := w.subscription.Fetch(w.config.PullBatchSize, nats.MaxWait(w.config.PullMaxWait))
			if err != nil {
				if err == nats.ErrTimeout {
					continue
				}
				w.logger.Printf("Fehler beim Abrufen von Nachrichten aus NATS: %v", err)
				time.Sleep(w.config.RetryDelay)
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
		w.processMessage(NatsMsgAdapter{msg})
	}
}

func (w *Worker) processMessage(msg NatsMsger) {
	var email domain.Email

	if err := json.Unmarshal(msg.Data(), &email); err != nil {
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
