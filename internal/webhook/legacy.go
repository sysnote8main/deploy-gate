package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/t1nyb0x/deploy-gate/internal/deploy"
)

const legacyMaxBodySize = 1 << 20 // 1MB

type legacyDeployResponse struct {
	Status string `json:"status"`
	Output string `json:"output,omitempty"`
}

// DeployLegacy is the original net/http handler (kept for benchmark comparison).
func DeployLegacy(secret, script string, pool *deploy.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, legacyMaxBodySize)
		defer r.Body.Close()

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		sig := r.Header.Get("X-Hub-Signature-256")
		if !validateLegacy(body, sig, secret) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		pool.Run(script)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_, _ = io.WriteString(w, `{"status":"accepted"}`)
	}
}

func validateLegacy(body []byte, sigHeader, secret string) bool {
	if secret == "" {
		return false
	}
	if !strings.HasPrefix(sigHeader, "sha256=") {
		return false
	}
	got, err := hex.DecodeString(strings.TrimPrefix(sigHeader, "sha256="))
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expected := mac.Sum(nil)
	return hmac.Equal(got, expected)
}

// newLegacyServer creates a net/http test server with the legacy handler.
func newLegacyServer(secret, script string, pool *deploy.Pool) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/deploy", DeployLegacy(secret, script, pool))
	return &http.Server{
		Handler: mux,
	}
}

// newFiberApp creates a Fiber app with the optimized handler.
func newFiberApp(secret, script string, pool *deploy.Pool) *fiber.App {
	app := fiber.New()
	app.Post("/deploy", Handler(secret, pool, script))
	return app
}

// generateSignature creates a valid HMAC-SHA256 signature for testing.
func generateSignature(body []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

// suppress unused import warning
var _ = log.Println
