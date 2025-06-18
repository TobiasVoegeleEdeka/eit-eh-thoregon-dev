package main

import (
	"log"
	"mailservice/internal/config"
	customhttp "mailservice/internal/delivery/http"
	"mailservice/internal/infrastructure/smtp"
	"mailservice/internal/usecase"
	"net/http"
	"os"
)

func main() {
	// Konfiguration laden
	smtpCfg, serverCfg := config.LoadConfig()

	// Logger initialisieren
	logger := log.New(os.Stdout, "[MAIL-API] ", log.LstdFlags|log.Lmicroseconds)

	// SMTP-Client erstellen
	smtpClient := smtp.NewClient(smtpCfg, logger)

	// Usecase initialisieren
	mailer := usecase.NewMailer(smtpClient)

	// HTTP-Handler erstellen
	handler := customhttp.NewHandler(mailer, logger)

	// Server starten
	http.HandleFunc("/send-email", handler.SendEmail)
	logger.Printf("Server starting on :%s", serverCfg.ListenPort)
	log.Fatal(http.ListenAndServe(":"+serverCfg.ListenPort, nil))
}
