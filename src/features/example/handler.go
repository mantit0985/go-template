package example

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/mantit0985/go-template/src/core/api"
)

func RegisterHandlers(router *api.Router) {
	svc := NewService()
	router.HandleFunc("/example", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			if err := api.Problem(w, http.StatusMethodNotAllowed, "Method Not Allowed", "Only POST is supported", "about:blank"); err != nil {
				slog.Error("failed to write problem response", slog.Any("error", err))
			}
			return
		}

		var req ExampleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			if err := api.BadRequest(w, "Invalid request payload", "about:blank"); err != nil {
				slog.Error("failed to write problem response", slog.Any("error", err))
			}
			return
		}

		res := ExampleResponse{
			Message: svc.DoWork(req.Name),
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(res); err != nil {
			if err := api.InternalServerError(w, "failed to encode response", "about:blank"); err != nil {
				slog.Error("failed to write problem response", slog.Any("error", err))
			}
		}
	})
}
