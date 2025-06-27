package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"smtp-worker/infrastructure/config"
	"smtp-worker/infrastructure/logging"
	"smtp-worker/infrastructure/smtp"
	"smtp-worker/internal/natsclient"
	"smtp-worker/internal/worker"
)

func main() {

	smtpCfg, workerCfg := config.LoadSMTPAndWorkerConfig()
	natsURL := os.Getenv("NATS_URL")

	mainLogger := logging.NewSMTPLogger()
	connLogger := logging.NewConnLogger()

	mainLogger.Printf("SMTP Worker wird initialisiert...")

	smtpClient := smtp.NewClient(smtpCfg, mainLogger, connLogger)
	sub, err := natsclient.NewPullSubscriber(
		natsURL,
		worker.JobsStream,
		worker.JobsSubject,
		"SMTP_WORKER_GROUP",
		mainLogger,
	)
	if err != nil {
		log.Fatalf("NATS-Initialisierung fehlgeschlagen: %v", err)
	}
	mainLogger.Printf("NATS-Abonnement ist bereit.")

	smtpWorker := worker.New(sub, smtpClient, mainLogger, workerCfg)

	go func() {
		mainLogger.Printf("SMTP Worker wird gestartet und wartet auf Aufträge...")
		if err := smtpWorker.Run(); err != nil {
			log.Fatalf("Fehler beim Ausführen des Workers: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	mainLogger.Printf("Shutdown-Signal empfangen, Worker wird heruntergefahren...")
	smtpWorker.Shutdown()
	mainLogger.Printf("Worker wurde sauber heruntergefahren.")
}
