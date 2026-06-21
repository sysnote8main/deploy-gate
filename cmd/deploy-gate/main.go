package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/t1nyb0x/deploy-gate/internal/config"
	"github.com/t1nyb0x/deploy-gate/internal/deploy"
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

	// Bounded worker pool — prevents goroutine explosion under load.
	pool := deploy.NewPool(0) // 0 = auto (2x CPU cores)
	defer pool.Shutdown()

	app := fiber.New(fiber.Config{
		Concurrency:       256 * 1024,
		DisableKeepalive: false,
		EnablePrintRoutes: true,
		StrictRouting:     true,
		AppName:           "deploy-gate",
	})

	app.Use(recover.New())

	// Register webhook routes.
	routes := make([]struct {
		Path   string
		Script string
	}, len(cfg.Routes))
	for i, r := range cfg.Routes {
		routes[i] = struct {
			Path   string
			Script string
		}{Path: r.Path, Script: r.Script}
	}
	webhook.RegisterRoutes(app, secret, pool, routes)

	log.Println("listening on :9000")
	log.Fatal(app.Listen(":9000"))
}
