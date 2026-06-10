package webhook

import (
	"io"
	"net/http"

	"github.com/t1nyb0x/deploy-gate/internal/deploy"
	"github.com/t1nyb0x/deploy-gate/internal/signature"
)

const maxBodySize = 1 << 20 // 1MB

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

		if err := deploy.Run(script); err != nil {
			http.Error(w, "error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}