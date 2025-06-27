package config

import (
	"os"
	"strconv"
	"time"
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

type WorkerConfig struct {
	WorkerCount   int
	PullBatchSize int
	PullMaxWait   time.Duration
	RetryDelay    time.Duration
}

func LoadSMTPAndWorkerConfig() (*SMTPConfig, *WorkerConfig) {
	smtpCfg := &SMTPConfig{
		ConnectIP:     getEnv("POSTFIX_CONNECT_IP", "host.docker.internal"),
		Port:          getEnv("POSTFIX_PORT", "25"),
		TargetFQDN:    getEnv("POSTFIX_TARGET_FQDN", "localhost"),
		DefaultSender: getEnv("DEFAULT_SENDER_EMAIL", "admin@mail-service-vm.francecentral.cloudapp.azure.com"),
	}

	workerCfg := &WorkerConfig{
		WorkerCount:   getEnvAsInt("WORKER_COUNT", 5),
		PullBatchSize: getEnvAsInt("PULL_BATCH_SIZE", 10),
		PullMaxWait:   time.Duration(getEnvAsInt("PULL_MAX_WAIT_SECONDS", 10)) * time.Second,
		RetryDelay:    time.Duration(getEnvAsInt("RETRY_DELAY_SECONDS", 5)) * time.Second,
	}

	return smtpCfg, workerCfg
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

func getEnvAsInt(key string, fallback int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return fallback
}
