package domain

import (
	"time"

	"github.com/google/uuid"
)

type EmailRequest struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

type EmailResponse struct {
	Status string `json:"status"`
}

type Email struct {
	From    string
	To      string
	Subject string
	Body    string
	Headers map[string]string
}

func NewEmail(from, to, subject, body string) *Email {
	return &Email{
		From:    from,
		To:      to,
		Subject: subject,
		Body:    body,
		Headers: map[string]string{
			"Date":         time.Now().Format(time.RFC1123Z),
			"Message-ID":   "<" + uuid.New().String() + ">",
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

	for k, v := range e.Headers {
		msg += k + ": " + v + "\r\n"
	}

	msg += "\r\n" + e.Body

	return msg
}
