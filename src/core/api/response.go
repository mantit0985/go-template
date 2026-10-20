package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

// ProblemDetail represents an RFC 7807 problem detail.
type ProblemDetail struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail,omitempty"`
	Instance string `json:"instance,omitempty"`
}

// WriteProblem writes an RFC 7807 problem detail response.
func WriteProblem(w http.ResponseWriter, problem ProblemDetail) error {
	if problem.Status < 100 || problem.Status > 599 {
		return fmt.Errorf("invalid HTTP status code: %d", problem.Status)
	}
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(problem.Status)
	if err := json.NewEncoder(w).Encode(problem); err != nil {
		slog.Error("error encoding problem detail", slog.Any("error", err))
		return err
	}
	return nil
}

// BadRequest is a helper for 400 Bad Request errors.
func BadRequest(w http.ResponseWriter, detail string, typeUri string) error {
	if typeUri == "" {
		typeUri = "about:blank"
	}
	return WriteProblem(w, ProblemDetail{
		Type:   typeUri,
		Title:  "Bad Request",
		Status: http.StatusBadRequest,
		Detail: detail,
	})
}

// NotFound is a helper for 404 Not Found errors.
func NotFound(w http.ResponseWriter, detail string, typeUri string) error {
	if typeUri == "" {
		typeUri = "about:blank"
	}
	return WriteProblem(w, ProblemDetail{
		Type:   typeUri,
		Title:  "Not Found",
		Status: http.StatusNotFound,
		Detail: detail,
	})
}

// InternalServerError is a helper for 500 Internal Server Error errors.
func InternalServerError(w http.ResponseWriter, internalDetail string, typeUri string) error {
	if typeUri == "" {
		typeUri = "about:blank"
	}
	// Log the internal detail, but return a generic message to the client
	slog.Error("internal server error", slog.String("detail", internalDetail))
	return WriteProblem(w, ProblemDetail{
		Type:   typeUri,
		Title:  "Internal Server Error",
		Status: http.StatusInternalServerError,
		Detail: "An unexpected error occurred. Please contact support.",
	})
}

// Problem is a generic helper for creating custom problem responses.
func Problem(w http.ResponseWriter, status int, title, detail, typeUri string) error {
	if typeUri == "" {
		typeUri = "about:blank"
	}
	return WriteProblem(w, ProblemDetail{
		Type:   typeUri,
		Title:  title,
		Status: status,
		Detail: detail,
	})
}
