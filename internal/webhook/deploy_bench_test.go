package webhook

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/t1nyb0x/deploy-gate/internal/deploy"
)

const benchSecret = "test-secret-key-for-benchmark"
const benchScript = "/usr/bin/true"

var benchBody = []byte(`{"ref":"refs/heads/main","repository":{"full_name":"test/repo"}}`)

// BenchmarkHTTPHandler benchmarks the original net/http handler.
func BenchmarkHTTPHandler(b *testing.B) {
	pool := deploy.NewPool(8)
	defer pool.Shutdown()

	handler := DeployLegacy(benchSecret, benchScript, pool)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/deploy", bytes.NewReader(benchBody))
		req.Header.Set("X-Hub-Signature-256", generateSignature(benchBody, benchSecret))

		w := httptest.NewRecorder()
		handler(w, req)
		_ = w.Result()
	}
}

// BenchmarkFiberHandler benchmarks the GoFiber handler.
func BenchmarkFiberHandler(b *testing.B) {
	pool := deploy.NewPool(8)
	defer pool.Shutdown()

	app := fiber.New()
	app.Post("/deploy", Handler(benchSecret, pool, benchScript))

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/deploy", bytes.NewReader(benchBody))
		req.Header.Set("X-Hub-Signature-256", generateSignature(benchBody, benchSecret))

		resp, _ := app.Test(req, -1)
		_ = resp
	}
}

// BenchmarkHTTPHandlerLargeBody benchmarks with a 10KB body.
func BenchmarkHTTPHandlerLargeBody(b *testing.B) {
	pool := deploy.NewPool(8)
	defer pool.Shutdown()

	handler := DeployLegacy(benchSecret, benchScript, pool)

	// Generate large body (~10KB).
	var largeBody []byte
	chunk := []byte(`{"data":"` + strings.Repeat("x", 250) + `"}`)
	for len(largeBody) < 10240 {
		largeBody = append(largeBody, chunk...)
		largeBody = append(largeBody, ',')
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/deploy", bytes.NewReader(largeBody))
		req.Header.Set("X-Hub-Signature-256", generateSignature(largeBody, benchSecret))

		w := httptest.NewRecorder()
		handler(w, req)
		_ = w.Result()
	}
}

// BenchmarkFiberHandlerLargeBody benchmarks Fiber with a 10KB body.
func BenchmarkFiberHandlerLargeBody(b *testing.B) {
	pool := deploy.NewPool(8)
	defer pool.Shutdown()

	app := fiber.New()
	app.Post("/deploy", Handler(benchSecret, pool, benchScript))

	var largeBody []byte
	chunk := []byte(`{"data":"` + strings.Repeat("x", 250) + `"}`)
	for len(largeBody) < 10240 {
		largeBody = append(largeBody, chunk...)
		largeBody = append(largeBody, ',')
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/deploy", bytes.NewReader(largeBody))
		req.Header.Set("X-Hub-Signature-256", generateSignature(largeBody, benchSecret))

		resp, _ := app.Test(req, -1)
		_ = resp
	}
}

// BenchmarkHTTPHandlerParallel benchmarks net/http under concurrent load.
func BenchmarkHTTPHandlerParallel(b *testing.B) {
	pool := deploy.NewPool(8)
	defer pool.Shutdown()

	handler := DeployLegacy(benchSecret, benchScript, pool)

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest(http.MethodPost, "/deploy", bytes.NewReader(benchBody))
			req.Header.Set("X-Hub-Signature-256", generateSignature(benchBody, benchSecret))

			w := httptest.NewRecorder()
			handler(w, req)
			io.Copy(io.Discard, w.Result().Body)
		}
	})
}

// BenchmarkFiberHandlerParallel benchmarks Fiber under concurrent load.
func BenchmarkFiberHandlerParallel(b *testing.B) {
	pool := deploy.NewPool(8)
	defer pool.Shutdown()

	app := fiber.New()
	app.Post("/deploy", Handler(benchSecret, pool, benchScript))

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest(http.MethodPost, "/deploy", bytes.NewReader(benchBody))
			req.Header.Set("X-Hub-Signature-256", generateSignature(benchBody, benchSecret))

			resp, _ := app.Test(req, -1)
			_ = resp
		}
	})
}
