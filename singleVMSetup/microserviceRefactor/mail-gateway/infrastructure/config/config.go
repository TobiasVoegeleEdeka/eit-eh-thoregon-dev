package config

import (
	"os"
)

type ServerConfig struct {
	ListenPort         string
	DefaultSenderEmail string
}

func LoadConfig() *ServerConfig {
	serverCfg := &ServerConfig{
		ListenPort:         getEnv("LISTEN_PORT", "8080"),
		DefaultSenderEmail: getEnv("DEFAULT_SENDER_EMAIL", "hostmaster@mail.edeka-inforservice.duckdns.org"),
	}

	return serverCfg
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
