package config

import (
	"os"
)

type ServerConfig struct {
	ListenPort string
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
