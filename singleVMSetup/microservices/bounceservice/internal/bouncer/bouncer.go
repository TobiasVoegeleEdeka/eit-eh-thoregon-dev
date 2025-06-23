package bouncer

import (
	"bounceservice/internal/parser"
	"bounceservice/types"
	"bufio"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

type Bouncer struct {
	Bounces  []types.Bounce
	mu       sync.Mutex
	LogFile  string
	OnBounce func(bounce types.Bounce)
}

func New(logFile string) *Bouncer {
	return &Bouncer{
		LogFile: logFile,
		Bounces: make([]types.Bounce, 0),
	}
}

func (b *Bouncer) Watch() {
	var lastKnownSize int64 = 0

	if info, err := os.Stat(b.LogFile); err == nil {
		lastKnownSize = info.Size()
	}

	for {
		time.Sleep(5 * time.Second)

		stat, err := os.Stat(b.LogFile)
		if err != nil {
			log.Printf("WARN: Could not stat log file: %v", err)
			continue
		}

		if stat.Size() > lastKnownSize {
			log.Printf("File has grown from %d to %d bytes. Reading new lines.", lastKnownSize, stat.Size())
			b.processNewLines(lastKnownSize, stat.Size())
			lastKnownSize = stat.Size()
		} else if stat.Size() < lastKnownSize {

			log.Printf("Log file shrunk from %d to %d bytes (likely rotated), resetting position", lastKnownSize, stat.Size())
			lastKnownSize = 0
		}
	}
}

func (b *Bouncer) processNewLines(startPos, endPos int64) {
	file, err := os.Open(b.LogFile)
	if err != nil {
		log.Printf("WARN: Could not open log file for reading: %v", err)
		return
	}
	defer file.Close()

	_, err = file.Seek(startPos, 0)
	if err != nil {
		log.Printf("WARN: Could not seek in log file: %v", err)
		return
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "status=bounced") {
			bounce := parser.ParseBounceLine(line)
			b.mu.Lock()
			b.Bounces = append(b.Bounces, bounce)
			b.mu.Unlock()

			if b.OnBounce != nil {
				b.OnBounce(bounce)
			}

			log.Printf("New bounce detected: %s -> %s: %s", bounce.From, bounce.To, bounce.Reason)
		}
	}
}
