package main

import (
	"fmt"
	"log"
	"net/http"

	"AutoPuller/config"
	"AutoPuller/deploy"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed")
		return
	}

	http.HandleFunc("/webhook", deploy.WebhookHandler)
	fmt.Println("Listening on :9082...")
	log.Fatal(http.ListenAndServe(":9082", nil))
}
