package mailer

import (
	"fmt"
	"mailservice/internal/domain"
	"mailservice/internal/infrastructure/smtp"
)

type Mailer struct {
	smtpClient *smtp.Client
	fromEmail  string // Store default sender
}

func NewMailer(client *smtp.Client, fromEmail string) *Mailer {
	return &Mailer{
		smtpClient: client,
		fromEmail:  fromEmail,
	}
}

type SendResult struct {
	Success      bool
	SMTPCode     string
	Message      string
	BounceReason string
	Error        error
}

func (m *Mailer) SendEmail(req *domain.EmailRequest) (*SendResult, error) {
	result := &SendResult{}

	if err := validateRequest(req); err != nil {
		result.Error = fmt.Errorf("validation failed: %w", err)
		return result, result.Error
	}

	// Das Email-Objekt wird wie zuvor erstellt
	email := domain.NewEmail(m.fromEmail, req.To, req.Subject, req.Body)

	// Geändert: Die Send-Methode wird jetzt mit dem gesamten Email-Objekt aufgerufen.
	smtpResult, err := m.smtpClient.Send(email)
	if err != nil {
		result.Success = false
		result.Error = fmt.Errorf("smtp delivery failed: %w", err)

		if smtpResult != nil {
			result.SMTPCode = smtpResult.SMTPCode
			result.Message = smtpResult.Message
			result.BounceReason = smtpResult.BounceReason
		}

		return result, result.Error
	}

	result.Success = true
	result.SMTPCode = smtpResult.SMTPCode
	result.Message = smtpResult.Message
	return result, nil
}

// Simplified version without extended result
func (m *Mailer) SendEmailSimple(req *domain.EmailRequest) error {
	_, err := m.SendEmail(req)
	return err
}

func validateRequest(req *domain.EmailRequest) error {
	if req.To == "" || req.Subject == "" || req.Body == "" {
		return fmt.Errorf("missing required fields: to=%s, subject=%s, body=%s",
			req.To, req.Subject, req.Body)
	}
	return nil
}
