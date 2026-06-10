package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/t1nyb0x/deploy-gate/internal/webhook"
)

func main() {
	secret := os.Getenv("DEPLOY_SECRET")
	queueDir := os.Getenv("QUEUE_DIR")

	if secret == "" {
		log.Fatal("DEPLOY_SECRET is required")
	}
	if queueDir == "" {
		log.Fatal("QUEUE_DIR is required")
	}


	mux := http.NewServeMux()
	mux.HandleFunc("/deploy/bot", webhook.Deploy(secret, "/scripts/deploy-bot.sh"))
	mux.HandleFunc("/deploy/dashboard", webhook.Deploy(secret, "/scripts/deploy-dashboard.sh"))

	server := &http.Server{
		Addr:              ":9000",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
	}

	log.Println("listening on :9000")
	log.Fatal(server.ListenAndServe())
}