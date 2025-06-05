package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"
)

// Konfigurationswerte, die jetzt aus Umgebungsvariablen gelesen werden
var (
	postfixHost   string
	postfixPort   string
	defaultSender string
	listenPort    string
)

func init() {
	// Lese Konfiguration aus Umgebungsvariablen oder setze Standardwerte
	postfixHost = getEnv("POSTFIX_HOST", "10.50.1.6")                                                        // Standard: Private IP Ihrer Postfix-VM
	postfixPort = getEnv("POSTFIX_PORT", "25")                                                               // Standard: SMTP-Relay-Port
	defaultSender = getEnv("DEFAULT_SENDER", "api-service@postfix-mail-vm.francecentral.cloudapp.azure.com") // Passen Sie den Standardwert an Ihren FQDN an
	listenPort = getEnv("LISTEN_PORT", "8080")                                                               // Port, auf dem die API lauscht
}

// Hilfsfunktion, um Umgebungsvariablen mit einem Standardwert zu lesen
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	log.Printf("Umgebungsvariable %s nicht gesetzt, verwende Standardwert: %s", key, fallback)
	return fallback
}

// EmailRequest Struktur für den JSON-Payload
type EmailRequest struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

// sendEmailHandler verarbeitet die API-Anfragen zum E-Mail-Versand
func sendEmailHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var req EmailRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "Invalid request payload: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if req.To == "" || req.Subject == "" || req.Body == "" {
		http.Error(w, "Missing fields: 'to', 'subject', and 'body' are required", http.StatusBadRequest)
		return
	}

	message := []byte(fmt.Sprintf("To: %s\r\n"+
		"From: %s\r\n"+
		"Subject: %s\r\n"+
		"Content-Type: text/plain; charset=UTF-8\r\n"+
		"\r\n"+
		"%s\r\n", req.To, defaultSender, req.Subject, req.Body))

	smtpAddr := fmt.Sprintf("%s:%s", postfixHost, postfixPort)

	// Beachten Sie: Für die TLS-Problematik mit dem IP-SAN Fehler
	// hatten wir eine komplexere SMTP-Client-Implementierung.
	// Dieser Code verwendet weiterhin das einfache smtp.SendMail.
	// Wenn der TLS-Fehler wieder auftritt, muss die erweiterte Implementierung
	// mit custom tls.Config und InsecureSkipVerify hier wieder rein.
	err := smtp.SendMail(smtpAddr, nil, defaultSender, []string{req.To}, message)
	if err != nil {
		log.Printf("Error sending email to %s: %v", req.To, err)
		http.Error(w, fmt.Sprintf("Failed to send email: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("Email successfully sent to %s with subject: %s", req.To, req.Subject)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Email accepted for delivery"})
}

func main() {
	http.HandleFunc("/send-email", sendEmailHandler)

	log.Printf("Starting API server on port %s...", listenPort)
	if err := http.ListenAndServe(":"+listenPort, nil); err != nil {
		log.Fatalf("Could not start server: %s\n", err)
	}
}
