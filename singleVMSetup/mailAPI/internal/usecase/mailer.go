package usecase

import (
	"fmt"
	"mailservice/internal/domain"
	"mailservice/internal/infrastructure/smtp"
)

type Mailer struct {
	smtpClient *smtp.Client
}

func NewMailer(client *smtp.Client) *Mailer {
	return &Mailer{smtpClient: client}
}

func (m *Mailer) SendEmail(req *domain.EmailRequest) error {
	if err := validateRequest(req); err != nil {
		return err
	}

	return m.smtpClient.Send(req.To, req.Subject, req.Body)
}

func validateRequest(req *domain.EmailRequest) error {
	if req.To == "" || req.Subject == "" || req.Body == "" {
		return fmt.Errorf("missing required fields")
	}
	return nil
}
