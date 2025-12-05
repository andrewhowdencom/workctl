package github

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/google/go-github/v60/github"
)

type Handler struct {
	logger *slog.Logger
	secret []byte
}

func NewHandler(logger *slog.Logger) *Handler {
	secret := os.Getenv("GITHUB_WEBHOOK_SECRET")
	return &Handler{
		logger: logger,
		secret: []byte(secret),
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	payload, err := github.ValidatePayload(r, h.secret)
	if err != nil {
		h.logger.Error("failed to validate webhook signature", "error", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	h.logger.Info("webhook received", "payload_size", len(payload))
	w.WriteHeader(http.StatusOK)
}
