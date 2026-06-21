package webhook

import (
	"github.com/gofiber/fiber/v2"
	"github.com/t1nyb0x/deploy-gate/internal/deploy"
	"github.com/t1nyb0x/deploy-gate/internal/signature"
)

const fiberMaxBodySize = 1 << 20 // 1MB

type fiberDeployResponse struct {
	Status string `json:"status"`
}

// Handler returns a Fiber handler that validates webhook signature and enqueues deploy.
// Used for benchmark comparison against the net/http version.
func Handler(secret string, pool *deploy.Pool, script string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Method() != fiber.MethodPost {
			return c.Status(fiber.StatusForbidden).SendString("forbidden")
		}

		body := c.Body()
		if len(body) > fiberMaxBodySize {
			return c.Status(fiber.StatusRequestEntityTooLarge).SendString("body too large")
		}

		sig := c.Get("X-Hub-Signature-256")
		if !signature.Validate(body, sig, secret) {
			return c.Status(fiber.StatusForbidden).SendString("forbidden")
		}

		pool.Run(script)

		return c.Status(fiber.StatusAccepted).JSON(fiberDeployResponse{Status: "accepted"})
	}
}
