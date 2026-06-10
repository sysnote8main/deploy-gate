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
	botScript := os.Getenv("DEPLOY_BOT_SCRIPT")
	dashboardScript := os.Getenv("DEPLOY_DASHBOARD_SCRIPT")

	if secret == "" {
		log.Fatal("DEPLOY_SECRET is required")
	}
	if botScript == "" {
		log.Fatal("DEPLOY_BOT_SCRIPT is required")
	}
	if dashboardScript == "" {
		log.Fatal("DEPLOY_DASHBOARD_SCRIPT is required")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/deploy/bot", webhook.Deploy(secret, botScript))
	mux.HandleFunc("/deploy/dashboard", webhook.Deploy(secret, dashboardScript))

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