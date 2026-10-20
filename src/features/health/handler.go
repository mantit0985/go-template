package health

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type Handler struct {
	service *Service
}

func NewHandler(s *Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	status := h.service.Check()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(status); err != nil {
		slog.Error("failed to encode health status", slog.Any("error", err))
	}
}

func (h *Handler) ReadyCheck(w http.ResponseWriter, r *http.Request) {
	status := h.service.Ready()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(status); err != nil {
		slog.Error("failed to encode ready status", slog.Any("error", err))
	}
}

func (h *Handler) VersionCheck(w http.ResponseWriter, r *http.Request) {
	version := h.service.Version()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(version); err != nil {
		slog.Error("failed to encode version", slog.Any("error", err))
	}
}
