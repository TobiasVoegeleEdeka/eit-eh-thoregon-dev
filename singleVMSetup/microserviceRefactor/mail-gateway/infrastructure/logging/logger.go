package logging

import (
	"log"
	"os"
)

type Logger interface {
	Printf(format string, v ...interface{})
}

type smtpLogger struct {
	logger *log.Logger
}

func NewSMTPLogger() Logger {
	return &smtpLogger{
		logger: log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime),
	}
}

func (l *smtpLogger) Printf(format string, v ...interface{}) {
	l.logger.Printf(format, v...)
}

type connLogger struct {
	logger *log.Logger
}

func NewConnLogger() Logger {
	return &connLogger{
		logger: log.New(os.Stdout, "CONN: ", log.Ldate|log.Ltime|log.Lmicroseconds),
	}
}

func (l *connLogger) Printf(format string, v ...interface{}) {
	l.logger.Printf(format, v...)
}
