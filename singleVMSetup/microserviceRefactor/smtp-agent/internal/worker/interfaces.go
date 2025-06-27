package worker

import (
	"smtp-worker/domain"
	"smtp-worker/infrastructure/smtp"

	"github.com/nats-io/nats.go"
)

type SMTPClienter interface {
	Send(email *domain.Email) (*smtp.DeliveryResult, error)
}

type NatsMsger interface {
	Ack(opts ...nats.AckOpt) error
	Data() []byte
}
