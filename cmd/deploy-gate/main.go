package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/t1nyb0x/deploy-gate/internal/config"
	"github.com/t1nyb0x/deploy-gate/internal/webhook"
)

func main() {
	secret := os.Getenv("DEPLOY_SECRET")
	configPath := os.Getenv("DEPLOY_CONFIG")

	if secret == "" {
		log.Fatal("DEPLOY_SECRET is required")
	}

	if configPath == "" {
		log.Fatal("DEPLOY_CONFIG is required")
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	mux := http.NewServeMux()

	for _, route := range cfg.Routes {
		log.Printf("register route: path=%s script=%s", route.Path, route.Script)
		mux.HandleFunc(route.Path, webhook.Deploy(secret, route.Script))
	}

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