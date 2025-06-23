package main

import (
	"log"
	"mailservice/internal/config"
	customhttp "mailservice/internal/delivery/http"
	"mailservice/internal/infrastructure/logging"
	"mailservice/internal/infrastructure/smtp"
	"mailservice/internal/mailer"
	"net/http"
)

func main() {
	// Konfiguration laden
	smtpCfg, serverCfg := config.LoadConfig()

	// Logger initialisieren
	mainLogger := logging.NewSMTPLogger()
	connLogger := logging.NewConnLogger()

	// SMTP-Client erstellen
	smtpClient := smtp.NewClient(smtpCfg, mainLogger, connLogger)

	// Mailer initialisieren mit SMTP-Client und Absender-Adresse aus der Konfiguration
	mailer := mailer.NewMailer(smtpClient, smtpCfg.DefaultSender)

	// HTTP-Handler erstellen
	handler := customhttp.NewHandler(mailer, mainLogger)

	// Server starten
	http.HandleFunc("/send-email", handler.SendEmail)
	mainLogger.Printf("Server starting on :%s", serverCfg.ListenPort)
	log.Fatal(http.ListenAndServe(":"+serverCfg.ListenPort, nil))
}
