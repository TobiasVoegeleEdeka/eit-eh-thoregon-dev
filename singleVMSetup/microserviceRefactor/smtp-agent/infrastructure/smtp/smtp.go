package smtp

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"

	"smtp-worker/domain"
	"smtp-worker/infrastructure/config"
	"smtp-worker/infrastructure/logging"
)

type Client struct {
	config     *config.SMTPConfig
	logger     logging.Logger
	connLogger logging.Logger
	timeout    time.Duration
}

type DeliveryResult struct {
	Success      bool
	SMTPCode     string
	Message      string
	BounceReason string
}

func NewClient(cfg *config.SMTPConfig, logger logging.Logger, connLogger logging.Logger) *Client {
	return &Client{
		config:     cfg,
		logger:     logger,
		connLogger: connLogger,
		timeout:    30 * time.Second,
	}
}

func (c *Client) Send(email *domain.Email) (*DeliveryResult, error) {
	result := &DeliveryResult{}
	startTime := time.Now()
	c.logger.Printf("Starte SMTP-Übergabe für E-Mail an %s", email.To)
	defer func() {
		c.logger.Printf("SMTP-Übergabe beendet (Dauer: %v)", time.Since(startTime))
	}()

	conn, err := net.DialTimeout("tcp", net.JoinHostPort(c.config.ConnectIP, c.config.Port), c.timeout)
	if err != nil {
		return nil, fmt.Errorf("TCP-Verbindung zu Postfix fehlgeschlagen: %w", err)
	}
	defer conn.Close()

	wrappedConn := &loggedConn{
		Conn:   conn,
		logger: c.connLogger,
	}

	client, err := smtp.NewClient(wrappedConn, c.config.TargetFQDN)
	if err != nil {
		return nil, fmt.Errorf("erstellung des SMTP-Clients fehlgeschlagen: %w", err)
	}
	defer func() {
		if err := client.Quit(); err != nil {
			c.logger.Printf("Fehler beim Senden von QUIT: %v", err)
		}
	}()

	if ok, _ := client.Extension("STARTTLS"); ok {
		if err := client.StartTLS(&tls.Config{
			ServerName:         c.config.TargetFQDN,
			InsecureSkipVerify: true,
		}); err != nil {
			return nil, fmt.Errorf("STARTTLS fehlgeschlagen: %w", err)
		}
	}

	if err := client.Mail(email.From); err != nil {
		return c.handleSMTPError(result, "MAIL FROM fehlgeschlagen: %w", err)
	}

	if err := client.Rcpt(email.To); err != nil {
		return c.handleSMTPError(result, "RCPT TO fehlgeschlagen: %w", err)
	}

	wc, err := client.Data()
	if err != nil {
		return c.handleSMTPError(result, "DATA-Kommando fehlgeschlagen: %w", err)
	}

	message := email.String()

	c.logger.Printf("Sende E-Mail-Inhalt (Message-ID: %s)...", email.Headers["Message-ID"])

	if _, err := fmt.Fprint(wc, message); err != nil {
		wc.Close()
		return c.handleSMTPError(result, "schreiben der Nachricht fehlgeschlagen: %w", err)
	}

	if err := wc.Close(); err != nil {
		return c.handleSMTPError(result, "schließen des Datenstroms fehlgeschlagen: %w", err)
	}

	result.Success = true
	result.SMTPCode = "250"
	result.Message = "Nachricht erfolgreich an Postfix zur Zustellung übergeben"
	c.logger.Printf("E-Mail erfolgreich an Postfix übergeben.")
	return result, nil
}

func (c *Client) handleSMTPError(result *DeliveryResult, format string, err error) (*DeliveryResult, error) {
	if errStr := err.Error(); strings.HasPrefix(errStr, "5") || strings.HasPrefix(errStr, "4") {
		parts := strings.SplitN(errStr, " ", 3)
		if len(parts) >= 2 {
			result.SMTPCode = parts[0] + " " + parts[1]
		}
		result.Message = errStr
	} else {
		result.Message = err.Error()
	}

	c.logger.Printf("SMTP-Fehler: %s (Code: %s)", err, result.SMTPCode)
	return result, fmt.Errorf(format, err)
}

type loggedConn struct {
	net.Conn
	logger logging.Logger
}

func (lc *loggedConn) Read(b []byte) (n int, err error) {
	n, err = lc.Conn.Read(b)
	if n > 0 {

		lc.logger.Printf("SMTP-IN: %q", strings.TrimSpace(string(b[:n])))
	}
	return
}

func (lc *loggedConn) Write(b []byte) (n int, err error) {
	n, err = lc.Conn.Write(b)
	if n > 0 {

		lc.logger.Printf("SMTP-OUT: %q", strings.TrimSpace(string(b[:n])))
	}
	return
}
