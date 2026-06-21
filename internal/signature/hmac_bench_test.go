package signature

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

// BenchmarkValidate benchmarks HMAC-SHA256 validation.
func BenchmarkValidate(b *testing.B) {
	secret := "test-secret-key"
	body := []byte(`{"ref":"refs/heads/main","repository":{"full_name":"test/repo"}}`)
	sig := "sha256=" + generateTestSig(body, secret)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		Validate(body, sig, secret)
	}
}

// BenchmarkValidateLargeBody benchmarks with a 10KB body.
func BenchmarkValidateLargeBody(b *testing.B) {
	secret := "test-secret-key"
	body := bytes.Repeat([]byte(`{"data":"payload"},`), 512) // ~10KB
	sig := "sha256=" + generateTestSig(body, secret)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		Validate(body, sig, secret)
	}
}

// BenchmarkValidateParallel benchmarks validation under concurrent load.
func BenchmarkValidateParallel(b *testing.B) {
	secret := "test-secret-key"
	body := []byte(`{"ref":"refs/heads/main","repository":{"full_name":"test/repo"}}`)
	sig := "sha256=" + generateTestSig(body, secret)

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			Validate(body, sig, secret)
		}
	})
}

func generateTestSig(body []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}
