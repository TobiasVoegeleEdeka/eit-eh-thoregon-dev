package main

import (
	"bufio"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

type BounceInfo struct {
	Timestamp time.Time `json:"timestamp"`
	Email     string    `json:"email"`
	Reason    string    `json:"reason"`
	DSN       string    `json:"dsn"`
}

var (
	bounces     []BounceInfo
	mu          sync.Mutex
	bounceRegex *regexp.Regexp
)

func init() {
	bounceRegex = regexp.MustCompile(`to=<(?P<email>[^>]+)>,.*?dsn=(?P<dsn>[^,]+),.*?status=bounced \((?P<reason>.*)\)`)
}

func main() {
	go watchPostfixLogs()

	http.HandleFunc("/bounces", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(bounces)
	})

	log.Println("Bounce service listening on :8081")
	http.ListenAndServe(":8081", nil)
}

func parseBounceLine(line string) *BounceInfo {
	matches := bounceRegex.FindStringSubmatch(line)
	if matches == nil {
		return nil
	}

	result := make(map[string]string)
	for i, name := range bounceRegex.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = matches[i]
		}
	}

	return &BounceInfo{
		Timestamp: time.Now().UTC(),
		Email:     result["email"],
		Reason:    result["reason"],
		DSN:       result["dsn"],
	}
}

func watchPostfixLogs() {
	var lastKnownSize int64 = 0

	info, err := os.Stat("/data/mail.log")
	if err == nil {
		lastKnownSize = info.Size()
	}

	for {
		time.Sleep(5 * time.Second)
		stat, err := os.Stat("/data/mail.log")
		if err != nil {
			log.Printf("WARN: Could not stat log file: %v", err)
			continue
		}
		if stat.Size() > lastKnownSize {
			file, err := os.Open("/data/mail.log")
			if err != nil {
				log.Printf("WARN: Could not open log file for reading: %v", err)
				continue
			}
			file.Seek(lastKnownSize, 0)
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := scanner.Text()
				if strings.Contains(line, "status=bounced") {
					if bounceInfo := parseBounceLine(line); bounceInfo != nil {
						mu.Lock()
						bounces = append(bounces, *bounceInfo)
						mu.Unlock()
						log.Printf("New structured bounce detected for email: %s", bounceInfo.Email)
					}
				}
			}
			file.Close()
			lastKnownSize = stat.Size()
		}
	}
}
