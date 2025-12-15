package github

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestHandler_ServeHTTP(t *testing.T) {
	secret := "my-secret"
	// Set env var before creating handler
	os.Setenv("GITHUB_WEBHOOK_SECRET", secret)
	defer os.Unsetenv("GITHUB_WEBHOOK_SECRET")

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	h := NewHandler(WithLogger(logger))

	tests := []struct {
		name           string
		body           string
		signature      string
		wantStatusCode int
	}{
		{
			name:           "valid signature",
			body:           `{"foo":"bar"}`,
			signature:      generateSignature(t, secret, `{"foo":"bar"}`),
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "invalid signature",
			body:           `{"foo":"bar"}`,
			signature:      "sha256=invalid",
			wantStatusCode: http.StatusUnauthorized,
		},
		{
			name:           "missing signature",
			body:           `{"foo":"bar"}`,
			signature:      "",
			wantStatusCode: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/v1/github/webhooks", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-GitHub-Event", "push") // Required for ParseWebHook
			if tt.signature != "" {
				req.Header.Set("X-Hub-Signature-256", tt.signature)
			}

			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatusCode {
				t.Errorf("handler returned wrong status code: got %v want %v",
					rr.Code, tt.wantStatusCode)
			}
		})
	}
}

func generateSignature(t *testing.T, secret, body string) string {
	t.Helper()
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(body))
	return fmt.Sprintf("sha256=%s", hex.EncodeToString(mac.Sum(nil)))
}
