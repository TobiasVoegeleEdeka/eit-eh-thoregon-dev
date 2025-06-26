package main

import (
	"log"
	customhttp "mail-gateway/delivery/http"
	"mail-gateway/infrastructure/config"
	"mail-gateway/infrastructure/logging"
	"net/http"
	"os"

	"github.com/nats-io/nats.go"
)

func main() {
	// --- Konfiguration und Setup ---
	serverCfg := config.LoadConfig()
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = nats.DefaultURL // Fallback auf den Standard-NATS-Port
	}

	// Logger initialisieren
	mainLogger := logging.NewSMTPLogger()

	mainLogger.Printf("Go-API Annahme-Service wird gestartet...")

	// --- NATS-Verbindung herstellen ---
	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatalf("Fehler beim Verbinden mit NATS: %v", err)
	}
	// Wichtig: Die Verbindung wird am Ende des Programms sauber geschlossen.
	defer nc.Close()
	mainLogger.Printf("Erfolgreich mit NATS verbunden unter: %s", nc.ConnectedUrl())

	// --- JetStream Context erstellen ---
	js, err := nc.JetStream()
	if err != nil {
		log.Fatalf("Fehler beim Erstellen des JetStream Context: %v", err)
	}

	// Stream erstellen (idempotent, d.h. erzeugt ihn nur, wenn er nicht existiert)
	_, err = js.AddStream(&nats.StreamConfig{
		Name:      "EMAIL_JOPS",               // Name des Streams
		Subjects:  []string{"email.jobs.new"}, // Das Subject, das in diesem Stream gespeichert wird
		Storage:   nats.FileStorage,           // Persistenz auf der Festplatte
		Retention: nats.WorkQueuePolicy,       // Nachricht wird nach Bestätigung durch einen Worker gelöscht
	})
	if err != nil {
		log.Fatalf("Fehler beim Erstellen des Streams: %v", err)
	}
	mainLogger.Printf("JetStream Stream 'EMAIL_JOBS' ist bereit.")

	// --- HTTP-Server starten ---
	// HTTP-Handler erstellen und NATS JetStream Context übergeben
	handler := customhttp.NewHandler(js, mainLogger)

	// Routen definieren
	http.HandleFunc("/send-email", handler.SendEmailHandler)

	mainLogger.Printf("Server lauscht auf Port :%s", serverCfg.ListenPort)
	if err := http.ListenAndServe(":"+serverCfg.ListenPort, nil); err != nil {
		log.Fatalf("Fehler beim Starten des HTTP-Servers: %v", err)
	}
}
