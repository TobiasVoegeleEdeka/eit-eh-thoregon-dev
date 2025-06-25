package domain

import (
	"time"

	"github.com/google/uuid"
)

type EmailRequest struct {
	From    string `json:"from,omitempty"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

type EmailResponse struct {
	Status    string `json:"status"`
	MessageID string `json:"messageId"`
}

type Email struct {
	From    string            `json:"from"`
	To      string            `json:"to"`
	Subject string            `json:"subject"`
	Body    string            `json:"body"`
	Headers map[string]string `json:"headers"`
}

func NewEmail(from, to, subject, body string) *Email {
	return &Email{
		From:    from,
		To:      to,
		Subject: subject,
		Body:    body,
		Headers: map[string]string{
			"Date":         time.Now().Format(time.RFC1123Z),
			"Message-ID":   "<" + uuid.New().String() + "@mailservice.local>",
			"MIME-Version": "1.0",
			"Content-Type": "text/plain; charset=utf-8",
		},
	}
}

func (e *Email) String() string {
	msg := ""
	msg += "From: " + e.From + "\r\n"
	msg += "To: " + e.To + "\r\n"
	msg += "Subject: " + e.Subject + "\r\n"

	// Füge alle weiteren Header hinzu
	for k, v := range e.Headers {
		msg += k + ": " + v + "\r\n"
	}

	msg += "\r\n" + e.Body

	return msg
}
