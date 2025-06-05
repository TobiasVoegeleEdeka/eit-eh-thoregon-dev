package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"
)

// Konfiguration - muss noch in env gepackt werden
const (
	// Private IP-Adresse der Postfix-VM
	// Diese sollte für die Kommunikation innerhalb Ihres VNet verwendet werden.
	postfixHost = "10.50.1.6"
	postfixPort = "25" // Standard-SMTP-Port für Relay/interne Einlieferung ohne Auth

	// Die Absenderadresse. Diese sollte zu einer Domain gehören,
	// die der Postfix-Server verarbeiten und korrekt umschreiben kann
	// (z.B. über sender_canonical_maps oder generic_maps in Postfix).
	// anpassen an FQDN oder  'sending_domain' .
	defaultSender = "api-service@postfix-mail-vm.francecentral.cloudapp.azure.com"
)

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

	// E-Mail-Nachricht im RFC 822 Format erstellen
	// Wichtig: \r\n als Zeilenumbrüche verwenden!
	message := []byte(fmt.Sprintf("To: %s\r\n"+
		"From: %s\r\n"+
		"Subject: %s\r\n"+
		"Content-Type: text/plain; charset=UTF-8\r\n"+ // Explizit Content-Type setzen
		"\r\n"+
		"%s\r\n", req.To, defaultSender, req.Subject, req.Body))

	// SMTP-Server-Adresse
	smtpAddr := fmt.Sprintf("%s:%s", postfixHost, postfixPort)

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

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Standardport, falls nicht über Umgebungsvariable gesetzt
	}

	log.Printf("Starting API server on port %s...", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Could not start server: %s\n", err)
	}
}
