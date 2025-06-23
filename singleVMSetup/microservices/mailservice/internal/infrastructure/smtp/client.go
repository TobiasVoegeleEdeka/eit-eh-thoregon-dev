package smtp

import (
	"crypto/tls"
	"fmt"
	"mailservice/internal/config"
	"mailservice/internal/domain" // <-- Hinzugefügter Import
	"mailservice/internal/infrastructure/logging"
	"net"
	"net/smtp"
	"strings"
	"time"
)

type Client struct {
	config     *config.SMTPConfig
	logger     logging.Logger
	connLogger logging.Logger
	timeout    time.Duration
}

// DeliveryResult fuer detaillierte Statusinformationen
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

// Geändert: Die Methode akzeptiert jetzt das domain.Email Objekt
func (c *Client) Send(email *domain.Email) (*DeliveryResult, error) {
	result := &DeliveryResult{}
	startTime := time.Now()
	// Geändert: Logger verwendet den Empfänger aus dem Email-Objekt
	c.logger.Printf("Starting email delivery to %s", email.To)
	defer func() {
		c.logger.Printf("Email delivery completed (duration: %v)", time.Since(startTime))
	}()

	// 1. Verbindung herstellen
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(c.config.ConnectIP, c.config.Port), c.timeout)
	if err != nil {
		return nil, fmt.Errorf("TCP connection failed: %w", err)
	}
	defer conn.Close()

	wrappedConn := &loggedConn{
		Conn:   conn,
		logger: c.connLogger,
	}

	// 2. SMTP-Client erstellen
	client, err := smtp.NewClient(wrappedConn, c.config.TargetFQDN)
	if err != nil {
		return nil, fmt.Errorf("SMTP client creation failed: %w", err)
	}
	defer func() {
		if err := client.Quit(); err != nil {
			c.logger.Printf("QUIT error: %v", err)
		}
	}()

	// 3. TLS handhaben
	if ok, _ := client.Extension("STARTTLS"); ok {
		c.logger.Printf("Initiating STARTTLS...")
		if err := client.StartTLS(&tls.Config{
			ServerName:         c.config.TargetFQDN,
			InsecureSkipVerify: true,
		}); err != nil {
			return nil, fmt.Errorf("STARTTLS failed: %w", err)
		}
	}

	// 4. Absender setzen (Geändert: wird aus dem Email-Objekt genommen)
	if err := client.Mail(email.From); err != nil {
		return c.handleSMTPError(client, result, "MAIL FROM failed: %w", err)
	}

	// 5. Empfänger setzen (Geändert: wird aus dem Email-Objekt genommen)
	if err := client.Rcpt(email.To); err != nil {
		return c.handleSMTPError(client, result, "RCPT TO failed: %w", err)
	}

	// 6. Datenübertragung
	wc, err := client.Data()
	if err != nil {
		return c.handleSMTPError(client, result, "DATA command failed: %w", err)
	}

	// Geändert: Die vollständige Nachricht wird von der email.String() Methode erstellt.
	// Das manuelle Zusammensetzen der Nachricht entfällt.
	message := email.String()

	if _, err := fmt.Fprint(wc, message); err != nil {
		wc.Close()
		return c.handleSMTPError(client, result, "message write failed: %w", err)
	}

	if err := wc.Close(); err != nil {
		return c.handleSMTPError(client, result, "message close failed: %w", err)
	}

	// Erfolgreiche Zustellung
	result.Success = true
	result.SMTPCode = "250"
	result.Message = "Message accepted for delivery"
	c.logger.Printf("Email successfully delivered to SMTP server")
	return result, nil
}

// handleSMTPError verarbeitet SMTP-Fehler und extrahiert Statusinformationen
func (c *Client) handleSMTPError(client *smtp.Client, result *DeliveryResult, format string, err error) (*DeliveryResult, error) {
	// Extrahiere SMTP Statuscode aus der Fehlermeldung
	if errStr := err.Error(); strings.Contains(errStr, "code=") {
		parts := strings.Split(errStr, "code=")
		if len(parts) > 1 {
			result.SMTPCode = strings.Split(parts[1], " ")[0]
		}
	}

	// Bounce-Grund identifizieren
	result.BounceReason = c.extractBounceReason(err.Error())
	result.Message = err.Error()
	c.logger.Printf("SMTP error: %s (Code: %s)", err, result.SMTPCode)

	return result, fmt.Errorf(format, err)
}

func (c *Client) extractBounceReason(response string) string {
	switch {
	case strings.Contains(response, "550"):
		return "Mailbox not found or access denied"
	case strings.Contains(response, "552"):
		return "Mailbox full"
	case strings.Contains(response, "554"):
		return "Transaction failed"
	case strings.Contains(response, "status=bounced"):
		if parts := strings.Split(response, "status=bounced"); len(parts) > 1 {
			return strings.Trim(parts[1], " ()")
		}
	}
	return "Unknown bounce reason"
}

type loggedConn struct {
	net.Conn
	logger logging.Logger
}

func (lc *loggedConn) Read(b []byte) (n int, err error) {
	n, err = lc.Conn.Read(b)
	if n > 0 {
		lc.logger.Printf("SMTP-IN: %q", string(b[:n]))
	}
	return
}

func (lc *loggedConn) Write(b []byte) (n int, err error) {
	n, err = lc.Conn.Write(b)
	if n > 0 {
		lc.logger.Printf("SMTP-OUT: %q", string(b[:n]))
	}
	return
}
