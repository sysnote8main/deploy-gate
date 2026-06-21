package webhook

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/t1nyb0x/deploy-gate/internal/deploy"
	"github.com/t1nyb0x/deploy-gate/internal/signature"
)

const maxBodySize = 1 << 20 // 1MB

type deployResponse struct {
	Status string `json:"status"`
}

// Handler returns a Fiber handler that validates webhook signature and enqueues deploy.
func Handler(secret string, pool *deploy.Pool, script string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Method() != fiber.MethodPost {
			return c.Status(fiber.StatusForbidden).SendString("forbidden")
		}

		body := c.Body()
		if len(body) > maxBodySize {
			return c.Status(fiber.StatusRequestEntityTooLarge).SendString("body too large")
		}

		sig := c.Get("X-Hub-Signature-256")
		if !signature.Validate(body, sig, secret) {
			return c.Status(fiber.StatusForbidden).SendString("forbidden")
		}

		pool.Run(script)

		return c.Status(fiber.StatusAccepted).JSON(deployResponse{Status: "accepted"})
	}
}

// RegisterRoutes registers all configured routes on the Fiber app.
func RegisterRoutes(app *fiber.App, secret string, pool *deploy.Pool, routes []struct {
	Path   string
	Script string
}) {
	for _, route := range routes {
		log.Printf("register route: path=%s script=%s", route.Path, route.Script)
		app.Post(route.Path, Handler(secret, pool, route.Script))
	}
}
