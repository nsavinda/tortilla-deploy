package main

import (
	"fmt"
	"log"
	"net/http"

	"AutoPuller/deploy"
	"AutoPuller/systemd"
)

func main() {

	systemd.GenerateServiceFile()

	http.HandleFunc("/webhook", deploy.WebhookHandler)
	fmt.Println("Listening on :9082...")
	log.Fatal(http.ListenAndServe(":9082", nil))
}
