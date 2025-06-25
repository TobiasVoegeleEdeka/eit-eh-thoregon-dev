package worker

import (
	"encoding/json"
	"time"

	"smtp-worker/domain"
	"smtp-worker/infrastructure/logging"
	"smtp-worker/infrastructure/smtp"

	"github.com/nats-io/nats.go"
)

const (
	JobsStream    = "EMAIL_JOBS"
	JobsSubject   = "email.jobs.new"
	PullBatchSize = 10
	PullMaxWait   = 10 * time.Second
	RetryDelay    = 5 * time.Second
)

type Worker struct {
	subscription *nats.Subscription
	smtpClient   *smtp.Client
	logger       logging.Logger
}

func New(sub *nats.Subscription, smtpClient *smtp.Client, logger logging.Logger) *Worker {
	return &Worker{
		subscription: sub,
		smtpClient:   smtpClient,
		logger:       logger,
	}
}

func (w *Worker) Run() error {
	for {
		msgs, err := w.subscription.Fetch(PullBatchSize, nats.MaxWait(PullMaxWait))
		if err != nil {
			if err == nats.ErrTimeout {
				continue
			}
			w.logger.Printf("Fehler beim Abrufen von Nachrichten aus NATS: %v", err)
			time.Sleep(RetryDelay)
			continue
		}

		w.logger.Printf("%d neue E-Mail-Aufträge aus der Queue geholt. Verarbeite...", len(msgs))

		for _, msg := range msgs {
			w.handleMessage(msg)
		}
	}
}

func (w *Worker) handleMessage(msg *nats.Msg) {
	var email domain.Email

	if err := json.Unmarshal(msg.Data, &email); err != nil {
		w.logger.Printf("FEHLER: Ungültige Nachricht erhalten, kann nicht deserialisiert werden: %v. Nachricht wird verworfen.", err)

		msg.Ack()
		return
	}

	w.logger.Printf("Versuche E-Mail (Message-ID: %s) an Postfix zu übergeben...", email.Headers["Message-ID"])

	_, err := w.smtpClient.Send(&email)
	if err != nil {
		w.logger.Printf("FEHLER beim Senden (Message-ID: %s): %v. Zustellung wird später erneut versucht.", email.Headers["Message-ID"], err)

		return
	}

	if err := msg.Ack(); err != nil {
		w.logger.Printf("FEHLER: Nachricht (Message-ID: %s) konnte nicht bei NATS bestätigt werden: %v", email.Headers["Message-ID"], err)
	} else {
		w.logger.Printf("E-Mail (Message-ID: %s) erfolgreich an Postfix übergeben und aus der Queue entfernt.", email.Headers["Message-ID"])
	}
}
