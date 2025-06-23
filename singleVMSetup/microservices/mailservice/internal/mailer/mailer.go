package mailer

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

	smtpResult, err := m.smtpClient.Send(req.To, req.Subject, req.Body)
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

// Vereinfachte Version ohne erweitertes Result
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
