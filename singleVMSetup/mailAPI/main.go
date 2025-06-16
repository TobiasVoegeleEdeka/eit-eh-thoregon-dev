package main

import (
	"crypto/tls" // Für benutzerdefinierte TLS-Konfiguration
	"encoding/json"
	"fmt"
	"log"
	"net" // Für net.Dial
	"net/http"
	"net/smtp"
	"os"
)

// Konfigurationswerte, die jetzt aus Umgebungsvariablen gelesen werden
var (
	// Der FQDN, der für TLS SNI und den EHLO-Befehl verwendet wird.
	// Dieser sollte über die Umgebungsvariable POSTFIX_FQDN gesetzt werden
	postfixTargetFQDN string

	// Die IP-Adresse, zu der die TCP-Verbindung tatsächlich aufgebaut wird
	// Dieser sollte über die Umgebungsvariable POSTFIX_CONNECT_IP gesetzt werden
	postfixConnectIP string

	postfixPort   string
	defaultSender string
	listenPort    string
)

func init() {
	// Lese Konfiguration aus Umgebungsvariablen oder setze Standardwerte
	postfixTargetFQDN = getEnv("POSTFIX_FQDN", "mail-service-vm.francecentral.cloudapp.azure.com")
	postfixConnectIP = getEnv("POSTFIX_CONNECT_IP", "10.50.1.7") // Standard: Private IP der Postfix-VM
	postfixPort = getEnv("POSTFIX_PORT", "25")
	defaultSender = getEnv("DEFAULT_SENDER", "api-service@mail-service-vm.francecentral.cloudapp.azure.com") // an FQDN anpassen
	listenPort = getEnv("LISTEN_PORT", "8080")
}

// Hilfsfunktion, um Umgebungsvariablen mit einem Standardwert zu lesen
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		if value != "" {
			return value
		}
	}
	log.Printf("Umgebungsvariable %s nicht gesetzt oder leer, verwende Standardwert: %s", key, fallback)
	return fallback
}

// EmailRequest Struktur für den JSON-Payload
type EmailRequest struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

// sendMailViaPostfix übernimmt die detaillierte SMTP-Logik
func sendMailViaPostfix(to, subject, body string) error {
	from := defaultSender
	recipients := []string{to}

	// Adresse für die TCP-Verbindung (IP-basiert für internes Routing)
	smtpConnectAddr := fmt.Sprintf("%s:%s", postfixConnectIP, postfixPort)

	// Verbindung aufbauen ZUERST OHNE TLS
	conn, err := net.Dial("tcp", smtpConnectAddr)
	if err != nil {
		return fmt.Errorf("failed to dial TCP to %s: %w", smtpConnectAddr, err)
	}
	// `defer conn.Close()` wird vom smtp.NewClient übernommen bzw. muss nach Fehlern von NewClient explizit erfolgen

	// SMTP-Client über die bestehende Verbindung erstellen
	// Wichtig: Hier den FQDN für EHLO etc. verwenden, den der Server erwartet
	c, err := smtp.NewClient(conn, postfixTargetFQDN)
	if err != nil {
		conn.Close() // Schließen, wenn NewClient fehlschlägt
		return fmt.Errorf("failed to create SMTP client with target host %s: %w", postfixTargetFQDN, err)
	}
	// `defer c.Quit()` stellt sicher, dass QUIT am Ende gesendet wird (oder bei Panic)

	// TLS-Konfiguration für STARTTLS
	tlsConfig := &tls.Config{
		ServerName:         postfixTargetFQDN, // Dieser Name wird für SNI verwendet und (wenn InsecureSkipVerify=false) für die Zertifikatsvalidierung
		InsecureSkipVerify: true,              // Zertifikatsprüfung clientseitig überspringen
	}

	// STARTTLS initiieren, falls vom Server angeboten
	if ok, _ := c.Extension("STARTTLS"); ok {
		if err = c.StartTLS(tlsConfig); err != nil {
			c.Close() // Verbindung schließen bei TLS-Fehler
			return fmt.Errorf("failed to start TLS with %s (target FQDN %s, connect IP %s): %w", smtpConnectAddr, postfixTargetFQDN, postfixConnectIP, err)
		}
	} else {
		log.Println("Warning: STARTTLS not offered by server. Sending unencrypted is not recommended.")

		// return fmt.Errorf("STARTTLS not offered by server %s", smtpConnectAddr)
	}

	// Sofern Postfix auf Port 25 für interne Relays Authentifizierung erfordert,
	// müsste hier c.Auth(auth) aufgerufen werden

	// Absender setzen
	if err = c.Mail(from); err != nil {
		c.Quit()
		return fmt.Errorf("failed to set mail from (%s): %w", from, err)
	}

	// Empfänger setzen
	for _, rcpt := range recipients {
		if err = c.Rcpt(rcpt); err != nil {
			c.Quit()
			return fmt.Errorf("failed to set rcpt to %s: %w", rcpt, err)
		}
	}

	// E-Mail-Daten senden
	wc, err := c.Data()
	if err != nil {
		c.Quit()
		return fmt.Errorf("failed to get data writer: %w", err)
	}

	message := fmt.Sprintf("To: %s\r\n"+
		"From: %s\r\n"+
		"Subject: %s\r\n"+
		"Content-Type: text/plain; charset=UTF-8\r\n"+
		"\r\n"+
		"%s\r\n", to, from, subject, body)

	_, err = fmt.Fprint(wc, message)
	if err != nil {
		wc.Close() // Versuche, den Writer zu schließen
		c.Quit()
		return fmt.Errorf("failed to write message: %w", err)
	}

	err = wc.Close()
	if err != nil {
		c.Quit()
		return fmt.Errorf("failed to close data writer: %w", err)
	}

	// Verbindung sauber beenden
	return c.Quit()
}

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

	err := sendMailViaPostfix(req.To, req.Subject, req.Body)
	if err != nil {
		log.Printf("Error sending email to %s: %v", req.To, err)
		http.Error(w, fmt.Sprintf("Failed to send email. Check server logs. Internal error: %v", err), http.StatusInternalServerError)
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
