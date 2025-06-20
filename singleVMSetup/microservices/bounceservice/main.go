package main

import (
	"bufio"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time" // Stellen Sie sicher, dass "time" importiert ist
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
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(bounces)
	})

	log.Println("Bounce service listening on :8081")
	http.ListenAndServe(":8081", nil)
}

// DIES IST DIE NEUE, ROBUSTE VERSION DER FUNKTION
func watchPostfixLogs() {
	var lastKnownSize int64 = 0 // Speichert die zuletzt gesehene Dateigröße

	// Zuerst die Startgröße der Datei ermitteln, damit wir nur neue Zeilen lesen
	info, err := os.Stat("/data/mail.log")
	if err == nil {
		lastKnownSize = info.Size()
	}

	// Eine Endlosschleife, die alle 5 Sekunden prüft
	for {
		time.Sleep(5 * time.Second)

		stat, err := os.Stat("/data/mail.log")
		if err != nil {
			log.Printf("WARN: Could not stat log file: %v", err)
			continue // Mache nach der Pause weiter
		}

		// Nur wenn die Datei gewachsen ist, lesen wir sie
		if stat.Size() > lastKnownSize {
			log.Printf("File has grown from %d to %d bytes. Reading new lines.", lastKnownSize, stat.Size())

			file, err := os.Open("/data/mail.log")
			if err != nil {
				log.Printf("WARN: Could not open log file for reading: %v", err)
				continue
			}

			// Springe zur letzten bekannten Position und lese von dort
			file.Seek(lastKnownSize, 0)
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
			file.Close()

			// Aktualisiere die Größe für den nächsten Durchlauf
			lastKnownSize = stat.Size()
		}
	}
}
