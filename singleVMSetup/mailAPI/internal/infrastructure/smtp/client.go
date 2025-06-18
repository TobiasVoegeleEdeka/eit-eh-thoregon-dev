package smtp

import (
	"crypto/tls"
	"fmt"
	"mailservice/internal/config"
	"mailservice/internal/infrastructure/logging"
	"net"
	"net/smtp"
	"time"
)

type Client struct {
	config     *config.SMTPConfig
	logger     logging.Logger
	connLogger logging.Logger
	timeout    time.Duration
}

func NewClient(cfg *config.SMTPConfig, logger logging.Logger, connLogger logging.Logger) *Client {
	return &Client{
		config:     cfg,
		logger:     logger,
		connLogger: connLogger,
		timeout:    30 * time.Second,
	}
}

// SetTimeout sets custom timeout for SMTP operations
func (c *Client) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
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

func (c *Client) Send(to, subject, body string) error {
	startTime := time.Now()
	c.logger.Printf("Starting email delivery to %s", to)
	defer func() {
		c.logger.Printf("Email delivery completed (duration: %v)", time.Since(startTime))
	}()

	// Establish connection
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(c.config.ConnectIP, c.config.Port), c.timeout)
	if err != nil {
		return fmt.Errorf("TCP connection failed: %w", err)
	}
	defer conn.Close()

	// Wrap connection for logging
	wrappedConn := &loggedConn{
		Conn:   conn,
		logger: c.connLogger,
	}

	// Create SMTP client
	client, err := smtp.NewClient(wrappedConn, c.config.TargetFQDN)
	if err != nil {
		return fmt.Errorf("SMTP client creation failed: %w", err)
	}
	defer func() {
		if err := client.Quit(); err != nil {
			c.logger.Printf("QUIT error: %v", err)
		}
	}()

	// TLS Configuration
	tlsConfig := &tls.Config{
		ServerName:         c.config.TargetFQDN,
		InsecureSkipVerify: true, // For testing only, use proper certs in production
	}

	// STARTTLS if available
	if ok, _ := client.Extension("STARTTLS"); ok {
		c.logger.Printf("Initiating STARTTLS...")
		if err := client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("STARTTLS failed: %w", err)
		}
		c.logger.Printf("STARTTLS established successfully")
	} else {
		c.logger.Printf("WARNING: Server does not support STARTTLS - insecure connection!")
	}

	// Set sender
	c.logger.Printf("Setting sender: %s", c.config.DefaultSender)
	if err := client.Mail(c.config.DefaultSender); err != nil {
		return fmt.Errorf("MAIL FROM failed: %w", err)
	}

	// Set recipient
	c.logger.Printf("Setting recipient: %s", to)
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("RCPT TO failed: %w", err)
	}

	// Send email data
	c.logger.Printf("Starting data transfer")
	wc, err := client.Data()
	if err != nil {
		return fmt.Errorf("DATA command failed: %w", err)
	}

	// Construct email message
	message := fmt.Sprintf("To: %s\r\nFrom: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s\r\n",
		to, c.config.DefaultSender, subject, body)

	// Write message
	c.logger.Printf("Sending email content (%d bytes)", len(message))
	if _, err := fmt.Fprint(wc, message); err != nil {
		wc.Close()
		return fmt.Errorf("message write failed: %w", err)
	}

	// Close writer
	if err := wc.Close(); err != nil {
		return fmt.Errorf("message close failed: %w", err)
	}

	c.logger.Printf("Email successfully delivered to SMTP server")
	return nil
}
