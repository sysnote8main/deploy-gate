package webhook

import (
	"bytes"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/t1nyb0x/deploy-gate/internal/deploy"
)

// --- HTTP/Net Benchmark Helpers ---

func setupHTTPServer(b *testing.B) (string, func()) {
	b.Helper()
	pool := deploy.NewPool(8)

	mux := http.NewServeMux()
	mux.HandleFunc("/deploy", DeployLegacy(benchSecret, benchScript, pool))

	server := &http.Server{
		Addr:              ":0",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		b.Fatal(err)
	}

	go server.Serve(ln)

	return "http://" + ln.Addr().String(), func() {
		server.Close()
		pool.Shutdown()
	}
}

func setupFiberServer(b *testing.B) (string, func()) {
	b.Helper()
	pool := deploy.NewPool(8)

	app := fiber.New()
	app.Post("/deploy", Handler(benchSecret, pool, benchScript))

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		b.Fatal(err)
	}

	go app.Listener(ln)

	return "http://" + ln.Addr().String(), func() {
		app.Shutdown()
		pool.Shutdown()
	}
}

// --- Network I/O Benchmarks ---

// BenchmarkHTTPNet benchmarks the HTTP handler over real TCP.
func BenchmarkHTTPNet(b *testing.B) {
	baseURL, cleanup := setupHTTPServer(b)
	defer cleanup()

	// Warmup connection.
	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
		},
	}
	req, _ := http.NewRequest(http.MethodPost, baseURL+"/deploy", bytes.NewReader(benchBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Hub-Signature-256", generateSignature(benchBody, benchSecret))
	resp, _ := client.Do(req)
	resp.Body.Close()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest(http.MethodPost, baseURL+"/deploy", bytes.NewReader(benchBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Hub-Signature-256", generateSignature(benchBody, benchSecret))

		resp, err := client.Do(req)
		if err != nil {
			b.Fatal(err)
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}

// BenchmarkFiberNet benchmarks the Fiber handler over real TCP.
func BenchmarkFiberNet(b *testing.B) {
	baseURL, cleanup := setupFiberServer(b)
	defer cleanup()

	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	// Warmup.
	req, _ := http.NewRequest(http.MethodPost, baseURL+"/deploy", bytes.NewReader(benchBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Hub-Signature-256", generateSignature(benchBody, benchSecret))
	resp, _ := client.Do(req)
	resp.Body.Close()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest(http.MethodPost, baseURL+"/deploy", bytes.NewReader(benchBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Hub-Signature-256", generateSignature(benchBody, benchSecret))

		resp, err := client.Do(req)
		if err != nil {
			b.Fatal(err)
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}

// BenchmarkHTTPNetLargeBody benchmarks HTTP over TCP with a 10KB body.
func BenchmarkHTTPNetLargeBody(b *testing.B) {
	baseURL, cleanup := setupHTTPServer(b)
	defer cleanup()

	var largeBody []byte
	chunk := []byte(`{"data":"` + strings.Repeat("x", 250) + `"}`)
	for len(largeBody) < 10240 {
		largeBody = append(largeBody, chunk...)
		largeBody = append(largeBody, ',')
	}

	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest(http.MethodPost, baseURL+"/deploy", bytes.NewReader(largeBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Hub-Signature-256", generateSignature(largeBody, benchSecret))

		resp, err := client.Do(req)
		if err != nil {
			b.Fatal(err)
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}

// BenchmarkFiberNetLargeBody benchmarks Fiber over TCP with a 10KB body.
func BenchmarkFiberNetLargeBody(b *testing.B) {
	baseURL, cleanup := setupFiberServer(b)
	defer cleanup()

	var largeBody []byte
	chunk := []byte(`{"data":"` + strings.Repeat("x", 250) + `"}`)
	for len(largeBody) < 10240 {
		largeBody = append(largeBody, chunk...)
		largeBody = append(largeBody, ',')
	}

	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest(http.MethodPost, baseURL+"/deploy", bytes.NewReader(largeBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Hub-Signature-256", generateSignature(largeBody, benchSecret))

		resp, err := client.Do(req)
		if err != nil {
			b.Fatal(err)
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}

// BenchmarkHTTPNetParallel benchmarks HTTP over TCP with concurrent clients.
func BenchmarkHTTPNetParallel(b *testing.B) {
	baseURL, cleanup := setupHTTPServer(b)
	defer cleanup()

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		client := &http.Client{
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 100,
				IdleConnTimeout:     90 * time.Second,
			},
		}

		for pb.Next() {
			req, _ := http.NewRequest(http.MethodPost, baseURL+"/deploy", bytes.NewReader(benchBody))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Hub-Signature-256", generateSignature(benchBody, benchSecret))

			resp, err := client.Do(req)
			if err != nil {
				b.Error(err)
				continue
			}
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
	})
}

// BenchmarkFiberNetParallel benchmarks Fiber over TCP with concurrent clients.
func BenchmarkFiberNetParallel(b *testing.B) {
	baseURL, cleanup := setupFiberServer(b)
	defer cleanup()

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		client := &http.Client{
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 100,
				IdleConnTimeout:     90 * time.Second,
			},
		}

		for pb.Next() {
			req, _ := http.NewRequest(http.MethodPost, baseURL+"/deploy", bytes.NewReader(benchBody))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Hub-Signature-256", generateSignature(benchBody, benchSecret))

			resp, err := client.Do(req)
			if err != nil {
				b.Error(err)
				continue
			}
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
	})
}
