package main

import (
	"bufio"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type Bounce struct {
	Timestamp string `json:"timestamp"`
	QueueID   string `json:"queue_id"`
	From      string `json:"from"`
	To        string `json:"to"`
	Status    string `json:"status"`
	Reason    string `json:"reason"`
	Raw       string `json:"raw,omitempty"`
}

var (
	bounces []Bounce
	mu      sync.Mutex
)

func main() {
	go watchPostfixLogs()

	http.HandleFunc("/bounces", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(bounces)
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	log.Println("Bounce service listening on :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}

func watchPostfixLogs() {
	var lastKnownSize int64 = 0
	logFile := "/data/mail.log"

	// Initialize with current size if file exists
	if info, err := os.Stat(logFile); err == nil {
		lastKnownSize = info.Size()
	}

	for {
		time.Sleep(5 * time.Second)

		stat, err := os.Stat(logFile)
		if err != nil {
			log.Printf("WARN: Could not stat log file: %v", err)
			continue
		}

		if stat.Size() > lastKnownSize {
			log.Printf("File has grown from %d to %d bytes. Reading new lines.", lastKnownSize, stat.Size())

			file, err := os.Open(logFile)
			if err != nil {
				log.Printf("WARN: Could not open log file for reading: %v", err)
				continue
			}

			_, err = file.Seek(lastKnownSize, 0)
			if err != nil {
				log.Printf("WARN: Could not seek in log file: %v", err)
				file.Close()
				continue
			}

			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := scanner.Text()
				if strings.Contains(line, "status=bounced") {
					bounce := parseBounceLine(line)
					mu.Lock()
					bounces = append(bounces, bounce)
					mu.Unlock()
					log.Printf("New bounce detected: %s -> %s: %s", bounce.From, bounce.To, bounce.Reason)
				}
			}
			file.Close()
			lastKnownSize = stat.Size()
		} else if stat.Size() < lastKnownSize {
			// Log file was rotated or truncated
			log.Printf("Log file shrunk from %d to %d bytes (likely rotated), resetting position", lastKnownSize, stat.Size())
			lastKnownSize = 0
		}
	}
}

func parseBounceLine(line string) Bounce {
	bounce := Bounce{
		Raw: line,
	}

	// Extract timestamp (assuming standard Postfix format)
	if parts := strings.SplitN(line, " ", 3); len(parts) >= 2 {
		bounce.Timestamp = parts[0] + " " + parts[1]
	}

	// Extract common Postfix fields
	bounce.QueueID = extractField(line, "queueid=", " ")
	bounce.From = extractField(line, "from=<", ">")
	bounce.To = extractField(line, "to=<", ">")
	bounce.Status = extractField(line, "status=", " ")
	bounce.Reason = extractBounceReason(line)

	// Clean up extracted fields
	if bounce.Reason == "" {
		bounce.Reason = "Unknown reason"
	}
	if bounce.From == "" {
		bounce.From = extractField(line, "from=", " ")
	}
	if bounce.To == "" {
		bounce.To = extractField(line, "to=", " ")
	}

	return bounce
}

func extractField(line, startDelim, endDelim string) string {
	startIdx := strings.Index(line, startDelim)
	if startIdx == -1 {
		return ""
	}

	startIdx += len(startDelim)
	endIdx := strings.Index(line[startIdx:], endDelim)
	if endIdx == -1 {
		return line[startIdx:]
	}

	return line[startIdx : startIdx+endIdx]
}

func extractBounceReason(line string) string {

	if reason := extractField(line, "reason=", " "); reason != "" {
		return reason
	}

	switch {
	case strings.Contains(line, "User unknown"):
		return "Recipient address does not exist"
	case strings.Contains(line, "Mailbox full"):
		return "Recipient mailbox is full"
	case strings.Contains(line, "Relay access denied"):
		return "Relay access denied"
	case strings.Contains(line, "Blocked"):
		return "Blocked by recipient server"
	case strings.Contains(line, "host not found"):
		return "Recipient domain not found"
	case strings.Contains(line, "Message too large"):
		return "Message size exceeds limit"
	case strings.Contains(line, "spam"):
		return "Marked as spam"
	case strings.Contains(line, "rejected"):
		return "Rejected by recipient server"
	}

	if parts := strings.Split(line, "status=bounced"); len(parts) > 1 {
		reason := strings.TrimSpace(parts[1])
		if len(reason) > 100 {
			reason = reason[:100] + "..."
		}
		return reason
	}

	return ""
}
