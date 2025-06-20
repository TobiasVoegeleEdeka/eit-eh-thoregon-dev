package logging

import (
	"log"
	"os"
)

type Logger interface {
	Printf(format string, v ...interface{})
}

type SMTPLogger struct {
	*log.Logger
}

func NewSMTPLogger() *SMTPLogger {
	return &SMTPLogger{
		Logger: log.New(os.Stdout, "[SMTP] ", log.LstdFlags|log.Lmicroseconds),
	}
}

type ConnLogger struct {
	*log.Logger
}

func NewConnLogger() *ConnLogger {
	return &ConnLogger{
		Logger: log.New(os.Stdout, "[SMTP-CONN] ", log.LstdFlags|log.Lmicroseconds),
	}
}
