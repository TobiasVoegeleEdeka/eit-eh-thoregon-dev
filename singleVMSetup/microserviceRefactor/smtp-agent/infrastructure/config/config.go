package config

import (
	"os"
)

type SMTPConfig struct {
	ConnectIP     string
	Port          string
	TargetFQDN    string
	DefaultSender string
}

type ServerConfig struct {
	ListenPort string
}

func LoadSMTPConfigOnly() (*SMTPConfig, error) {
	smtpCfg := &SMTPConfig{
		ConnectIP:     getEnv("POSTFIX_CONNECT_IP", "host.docker.internal"),
		Port:          getEnv("POSTFIX_PORT", "25"),
		TargetFQDN:    getEnv("POSTFIX_TARGET_FQDN", "localhost"),
		DefaultSender: getEnv("DEFAULT_SENDER_EMAIL", "noreply@example.com"),
	}
	return smtpCfg, nil
}

func LoadConfig() *ServerConfig {
	serverCfg := &ServerConfig{
		ListenPort: getEnv("LISTEN_PORT", "8080"),
	}
	return serverCfg
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
