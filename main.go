package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"AutoPuller/config"
	"AutoPuller/systemd"
	"AutoPuller/webhook"
)

func main() {
	if os.Getenv("WEBHOOK_SECRET") == "" {
		log.Fatal("WEBHOOK_SECRET environment variable not set")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Configuration loading failed:", err)
	}

	if _, err := systemd.GenerateAllServiceFiles(); err != nil {
		log.Fatal("Service generation failed:", err)
	}

	http.HandleFunc("/webhook", webhook.Handler)
	fmt.Printf("Starting server on port %d...\n", cfg.WebhookPort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", cfg.WebhookPort), nil))
}
