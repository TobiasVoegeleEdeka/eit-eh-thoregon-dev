package http

import (
	"encoding/json"
	"mail-gateway/domain"
	"mail-gateway/infrastructure/logging"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

type Handler struct {
	js            nats.JetStreamContext
	logger        logging.Logger
	defaultSender string
}

func NewHandler(js nats.JetStreamContext, logger logging.Logger, defaultSender string) *Handler {
	return &Handler{
		js:            js,
		logger:        logger,
		defaultSender: defaultSender,
	}
}

func (h *Handler) SendEmailHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Ungültige Anfragemethode", http.StatusMethodNotAllowed)
		return
	}

	var req domain.EmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Ungültiger Request-Body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if req.To == "" || req.Subject == "" || req.Body == "" {
		http.Error(w, "Fehlende Pflichtfelder: to, subject, body", http.StatusBadRequest)
		return
	}

	// --- Kernlogik: E-Mail-Objekt erstellen und an NATS übergeben ---

	fullEmail := &domain.Email{
		From:    req.From,
		To:      req.To,
		Subject: req.Subject,
		Body:    req.Body,
		Headers: map[string]string{
			"Date":         time.Now().Format(time.RFC1123Z),
			"Message-ID":   "<" + uuid.New().String() + "@mailservice.local>",
			"MIME-Version": "1.0",
			"Content-Type": "text/plain; charset=utf-8",
		},
	}

	if fullEmail.From == "" {
		fullEmail.From = h.defaultSender
	}

	h.logger.Printf("Neuer E-Mail-Auftrag erhalten. Message-ID: %s", fullEmail.Headers["Message-ID"])

	emailJSON, err := json.Marshal(fullEmail)
	if err != nil {
		h.logger.Printf("FEHLER: Serialisierung fehlgeschlagen: %v", err)
		http.Error(w, "Interner Serverfehler", http.StatusInternalServerError)
		return
	}

	ack, err := h.js.Publish("email.jobs.new", emailJSON)
	if err != nil {
		h.logger.Printf("FEHLER: Veröffentlichung an NATS fehlgeschlagen: %v", err)
		http.Error(w, "Fehler beim Einreihen des Auftrags", http.StatusServiceUnavailable)
		return
	}

	h.logger.Printf("Auftrag erfolgreich in NATS Stream '%s' (Sequenz: %d) veröffentlicht.", ack.Stream, ack.Sequence)

	responsePayload := map[string]string{
		"status":    "queued",
		"messageId": fullEmail.Headers["Message-ID"],
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(responsePayload)
}
