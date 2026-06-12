package webhook

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/t1nyb0x/deploy-gate/internal/deploy"
	"github.com/t1nyb0x/deploy-gate/internal/signature"
)

const maxBodySize = 1 << 20 // 1MB

type deployResponse struct {
	Status string `json:"status"`
	Output string `json:"output,omitempty"`
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func Deploy(secret, script string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
		defer r.Body.Close()

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		sig := r.Header.Get("X-Hub-Signature-256")
		if !signature.Validate(body, sig, secret) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		go func() {
			output, err := deploy.Run(script)
			if err != nil {
				log.Printf("deploy failed: script=%s error=%v output=%s", script, err, output)
				return
			}
			log.Printf("deploy succeeded: script=%s output=%s", script, output)
		}()

		writeJSON(w, http.StatusAccepted, deployResponse{Status: "accepted"})
	}
}