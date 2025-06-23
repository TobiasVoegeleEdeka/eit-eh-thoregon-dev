package main

import (
	"bounceservice/internal/bouncer"
	"bounceservice/internal/server"
	"log"
	"net/http"
)

func main() {

	b := bouncer.New("/data/mail.log")

	s := server.New(b)
	s.RegisterRoutes()

	go b.Watch()

	log.Println("Bounce service listening on :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
