package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
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

	queueFile := filepath.Join(queueDir, "deploy.queue")

	mux := http.NewServeMux()
	mux.HandleFunc("/deploy", webhook.Deploy(secret, queueFile))

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