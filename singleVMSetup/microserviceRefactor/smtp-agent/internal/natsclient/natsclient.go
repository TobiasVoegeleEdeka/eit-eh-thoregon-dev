package natsclient

import (
	"fmt"
	"time"

	"smtp-worker/infrastructure/logging"

	"github.com/nats-io/nats.go"
)

func NewPullSubscriber(natsURL, streamName, subject, durableName string, logger logging.Logger) (*nats.Subscription, error) {
	if natsURL == "" {
		natsURL = nats.DefaultURL
	}

	nc, err := nats.Connect(natsURL, nats.MaxReconnects(-1), nats.ReconnectWait(2*time.Second))
	if err != nil {
		return nil, fmt.Errorf("fehler beim Verbinden mit NATS: %w", err)
	}
	logger.Printf("Erfolgreich mit NATS verbunden unter: %s", nc.ConnectedUrl())

	js, err := nc.JetStream()
	if err != nil {

		return nil, fmt.Errorf("fehler beim Erstellen des JetStream Context: %w", err)
	}

	stream, err := js.StreamInfo(streamName)
	if err != nil {

		if err == nats.ErrStreamNotFound {
			logger.Printf("Stream '%s' nicht gefunden, wird erstellt...", streamName)
			_, err = js.AddStream(&nats.StreamConfig{
				Name:      streamName,
				Subjects:  []string{subject},
				Storage:   nats.FileStorage,
				Retention: nats.WorkQueuePolicy,
			})
			if err != nil {
				return nil, fmt.Errorf("fehler beim Erstellen des Streams '%s': %w", streamName, err)
			}
		} else {

			return nil, fmt.Errorf("fehler beim Abrufen der Stream-Info für '%s': %w", streamName, err)
		}
	}
	if stream != nil {
		logger.Printf("JetStream Stream '%s' ist bereit.", streamName)
	}

	sub, err := js.PullSubscribe(subject, durableName)
	if err != nil {
		return nil, fmt.Errorf("fehler beim Erstellen der Pull Subscription für '%s': %w", subject, err)
	}

	return sub, nil
}
