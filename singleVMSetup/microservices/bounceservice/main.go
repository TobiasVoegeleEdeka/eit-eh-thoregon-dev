package main

import (
	"bufio"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
)

var (
	bounces []string
	mu      sync.Mutex
)

func main() {
	// Starte Log-Überwachung im Hintergrund
	go watchPostfixLogs()

	// API-Endpoint
	http.HandleFunc("/bounces", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		json.NewEncoder(w).Encode(bounces)
	})

	log.Println("Bounce service listening on :8081")
	http.ListenAndServe(":8081", nil)
}

func watchPostfixLogs() {
	file, err := os.Open("/var/log/mail.log")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Log-Datei vom Ende lesen
	file.Seek(0, 2)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "status=bounced") {
			mu.Lock()
			bounces = append(bounces, line)
			mu.Unlock()
			log.Println("New bounce detected:", line)
		}
	}
}
