package smtp

import (
	"fmt"
	"mailservice/internal/config"
	"net"
	"net/smtp"
)

type LoggedConn struct {
	net.Conn
	LogFunc func(string)
}

func (lc *LoggedConn) Read(b []byte) (n int, err error) {
	n, err = lc.Conn.Read(b)
	if n > 0 && lc.LogFunc != nil {
		lc.LogFunc(fmt.Sprintf("SMTP-IN: %s", string(b[:n])))
	}
	return
}

func (lc *LoggedConn) Write(b []byte) (n int, err error) {
	n, err = lc.Conn.Write(b)
	if n > 0 && lc.LogFunc != nil {
		lc.LogFunc(fmt.Sprintf("SMTP-OUT: %s", string(b[:n])))
	}
	return
}

type Client struct {
	config *config.SMTPConfig
	logger Logger
}

type Logger interface {
	Printf(format string, v ...interface{})
}

func NewClient(cfg *config.SMTPConfig, logger Logger) *Client {
	return &Client{config: cfg, logger: logger}
}

func (c *Client) Send(to, subject, body string) error {
	conn, err := net.Dial("tcp", net.JoinHostPort(c.config.ConnectIP, c.config.Port))
	if err != nil {
		return fmt.Errorf("dial failed: %w", err)
	}
	defer conn.Close()

	loggedConn := &LoggedConn{
		Conn:    conn,
		LogFunc: func(s string) { c.logger.Printf(s) },
	}

	client, err := smtp.NewClient(loggedConn, c.config.TargetFQDN)
	if err != nil {
		return fmt.Errorf("SMTP client creation failed: %w", err)
	}
	defer client.Quit()

	// TLS und Mail-Versand Logik hier...
	// (Wie in Ihrem ursprünglichen Code)

	return nil
}
