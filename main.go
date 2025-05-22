package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"AutoPuller/systemd"
	"AutoPuller/webhook"
)

func main() {
	if os.Getenv("WEBHOOK_SECRET") == "" {
		log.Fatal("WEBHOOK_SECRET environment variable not set")
	}

	if _, err := systemd.GenerateServiceFiles(); err != nil {
		log.Fatal("Service generation failed:", err)
	}

	http.HandleFunc("/webhook", webhook.Handler)
	fmt.Println("Listening on :9082...")
	log.Fatal(http.ListenAndServe(":9082", nil))
}
