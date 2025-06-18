package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/smtp"
	"os"
	"time"
)

var (
	postfixTargetFQDN string
	postfixConnectIP  string
	postfixPort       string
	defaultSender     string
	listenPort        string
	logger            *log.Logger
)

func init() {
	// Logger mit Timestamps initialisieren
	logger = log.New(os.Stdout, "[MAIL-API] ", log.LstdFlags|log.Lmicroseconds)

	// Konfiguration laden
	postfixTargetFQDN = getEnv("POSTFIX_FQDN", "mail-service-vm.francecentral.cloudapp.azure.com")
	postfixConnectIP = getEnv("POSTFIX_CONNECT_IP", "localhost")
	postfixPort = getEnv("POSTFIX_PORT", "25")
	defaultSender = getEnv("DEFAULT_SENDER", "mail-service@mail-service-vm.francecentral.cloudapp.azure.com")
	listenPort = getEnv("LISTEN_PORT", "8080")

	logger.Printf("Konfiguration geladen: FQDN=%s, IP=%s, Port=%s", postfixTargetFQDN, postfixConnectIP, postfixPort)
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		if value != "" {
			return value
		}
	}
	logger.Printf("Umgebungsvariable %s nicht gesetzt, verwende Standardwert: %s", key, fallback)
	return fallback
}

type EmailRequest struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

func logSMTPInteraction(step string, action string, response string) {
	logger.Printf("SMTP-INTERACTION | %-15s | %-30s | %s", step, action, response)
}

func sendMailViaPostfix(to, subject, body string) error {
	startTime := time.Now()
	logger.Printf("Starte Mailversand an %s", to)
	defer func() {
		logger.Printf("Mailversand abgeschlossen (Dauer: %v)", time.Since(startTime))
	}()

	from := defaultSender
	recipients := []string{to}
	smtpConnectAddr := fmt.Sprintf("%s:%s", postfixConnectIP, postfixPort)

	logger.Printf("Stelle TCP-Verbindung her zu %s", smtpConnectAddr)
	conn, err := net.Dial("tcp", smtpConnectAddr)
	if err != nil {
		logger.Printf("TCP-Verbindungsfehler: %v", err)
		return fmt.Errorf("failed to dial TCP: %w", err)
	}

	// Wrap connection for logging
	loggingConn := &loggedConn{Conn: conn, logger: logger}
	defer func() {
		if conn != nil {
			conn.Close()
		}
	}()

	logger.Printf("Erstelle SMTP-Client mit FQDN: %s", postfixTargetFQDN)
	c, err := smtp.NewClient(loggingConn, postfixTargetFQDN)
	if err != nil {
		logger.Printf("SMTP-Client-Erstellungsfehler: %v", err)
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer func() {
		if err := c.Quit(); err != nil {
			logger.Printf("QUIT-Fehler: %v", err)
		}
	}()

	// TLS-Konfiguration
	tlsConfig := &tls.Config{
		ServerName:         postfixTargetFQDN,
		InsecureSkipVerify: true,
	}

	if ok, _ := c.Extension("STARTTLS"); ok {
		logger.Printf("STARTTLS wird initiiert...")
		if err := c.StartTLS(tlsConfig); err != nil {
			logger.Printf("STARTTLS-Fehler: %v", err)
			return fmt.Errorf("STARTTLS failed: %w", err)
		}
		logger.Printf("STARTTLS erfolgreich eingerichtet")
	} else {
		logger.Printf("WARNUNG: Server bietet kein STARTTLS an - unsichere Verbindung!")
	}

	logger.Printf("Setze Absender: %s", from)
	if err := c.Mail(from); err != nil {
		logger.Printf("MAIL FROM-Fehler: %v", err)
		return fmt.Errorf("MAIL FROM failed: %w", err)
	}

	for _, rcpt := range recipients {
		logger.Printf("Setze Empfänger: %s", rcpt)
		if err := c.Rcpt(rcpt); err != nil {
			logger.Printf("RCPT TO-Fehler für %s: %v", rcpt, err)
			return fmt.Errorf("RCPT TO failed: %w", err)
		}
	}

	logger.Printf("Beginne mit Datentransfer")
	wc, err := c.Data()
	if err != nil {
		logger.Printf("DATA-Fehler: %v", err)
		return fmt.Errorf("DATA command failed: %w", err)
	}

	message := fmt.Sprintf("To: %s\r\nFrom: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s\r\n",
		to, from, subject, body)

	logger.Printf("Sende E-Mail-Inhalt (%d bytes)", len(message))
	if _, err := fmt.Fprint(wc, message); err != nil {
		logger.Printf("Schreibfehler beim Datentransfer: %v", err)
		wc.Close()
		return fmt.Errorf("message write failed: %w", err)
	}

	if err := wc.Close(); err != nil {
		logger.Printf("Fehler beim Beenden des Datentransfers: %v", err)
		return fmt.Errorf("message close failed: %w", err)
	}

	logger.Printf("E-Mail erfolgreich an Postfix übergeben")
	return nil
}

// loggedConn protokolliert alle gelesenen und geschriebenen Daten
type loggedConn struct {
	net.Conn
	logger *log.Logger
	buf    []byte
}

func (lc *loggedConn) Read(b []byte) (n int, err error) {
	n, err = lc.Conn.Read(b)
	if n > 0 {
		lc.logger.Printf("SMTP-TRAFFIC <<< %s", string(b[:n]))
	}
	return
}

func (lc *loggedConn) Write(b []byte) (n int, err error) {
	n, err = lc.Conn.Write(b)
	if n > 0 {
		lc.logger.Printf("SMTP-TRAFFIC >>> %s", string(b[:n]))
	}
	return
}

func sendEmailHandler(w http.ResponseWriter, r *http.Request) {
	logger.Printf("Neue Anfrage von %s", r.RemoteAddr)

	if r.Method != http.MethodPost {
		logger.Printf("Falsche HTTP-Methode: %s", r.Method)
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var req EmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Printf("JSON-Parsing-Fehler: %v", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if req.To == "" || req.Subject == "" || req.Body == "" {
		logger.Printf("Fehlende Pflichtfelder: to=%s, subject=%s, body=%s", req.To, req.Subject, req.Body)
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	logger.Printf("Verarbeite Anfrage für Empfänger: %s", req.To)
	if err := sendMailViaPostfix(req.To, req.Subject, req.Body); err != nil {
		logger.Printf("Fehler beim Mailversand: %v", err)
		http.Error(w, fmt.Sprintf("Failed to send email: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	logger.Printf("Anfrage erfolgreich verarbeitet")
}

func main() {
	logger.Printf("Starte Mail-API-Server auf Port %s", listenPort)
	http.HandleFunc("/send-email", sendEmailHandler)

	if err := http.ListenAndServe(":"+listenPort, nil); err != nil {
		logger.Fatalf("Serverfehler: %v", err)
	}
}
