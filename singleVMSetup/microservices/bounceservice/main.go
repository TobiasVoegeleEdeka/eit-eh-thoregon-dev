package main

import (
	"bufio"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	// Starte HTTP-Server für Healthchecks
	go func() {
		http.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		http.ListenAndServe(":8081", nil)
	}()

	// Log-Überwachung
	logFile := os.Getenv("POSTFIX_LOG_PATH")
	if logFile == "" {
		logFile = "/var/log/mail.log"
	}

	file, err := os.Open(logFile)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "status=bounced") {
			log.Printf("BOUNCE DETECTED: %s", line)

		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Log scanner error: %v", err)
	}
}
