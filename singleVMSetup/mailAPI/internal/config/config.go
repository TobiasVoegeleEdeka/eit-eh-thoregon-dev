package config

import (
	"log"
	"os"
)

type SMTPConfig struct {
	TargetFQDN    string
	ConnectIP     string
	Port          string
	DefaultSender string
}

type ServerConfig struct {
	ListenPort string
}

func LoadConfig() (*SMTPConfig, *ServerConfig) {
	smtpCfg := &SMTPConfig{
		TargetFQDN:    getEnv("POSTFIX_FQDN", "mail-service-vm.francecentral.cloudapp.azure.com"),
		ConnectIP:     getEnv("POSTFIX_CONNECT_IP", "10.50.1.7"),
		Port:          getEnv("POSTFIX_PORT", "25"),
		DefaultSender: getEnv("DEFAULT_SENDER", "mail-service@mail-service-vm.francecentral.cloudapp.azure.com"),
	}

	serverCfg := &ServerConfig{
		ListenPort: getEnv("LISTEN_PORT", "8080"),
	}

	return smtpCfg, serverCfg
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	log.Printf("Using default for %s: %s", key, fallback)
	return fallback
}
