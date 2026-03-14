// Package response provides helpers for writing consistent JSON HTTP responses.
package response

import (
	"encoding/json"
	"net/http"
	"time"
)

// Verdict is the machine-readable outcome code included in every API response.
type Verdict string

const (
	// VerdictSuccess indicates the request was handled successfully.
	VerdictSuccess Verdict = "success"
	// VerdictError indicates the request could not be fulfilled.
	VerdictError Verdict = "error"
)

// Envelope is the standard response wrapper returned by all API endpoints.
//
// Every response carries:
//   - verdict  — machine-readable outcome code ("success" or "error")
//   - message  — human-readable description; empty string for success responses
//   - time     — response timestamp in RFC 3339 format
//   - data     — payload; empty object for error responses
type Envelope struct {
	Data    any     `json:"data"`
	Message string  `json:"message"`
	Time    string  `json:"time"`
	Verdict Verdict `json:"verdict"`
}

// now returns the current time formatted as RFC 3339 with sub-second precision.
func now() string {
	return time.Now().Format(time.RFC3339)
}

// JSON writes a JSON-encoded body with the given HTTP status code.
func JSON(w http.ResponseWriter, statusCode int, payload Envelope) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, "encoding error", http.StatusInternalServerError)
	}
}

// OK writes a 200 OK JSON response with verdict "success".
func OK(w http.ResponseWriter, data any) {
	JSON(w, http.StatusOK, Envelope{
		Data:    data,
		Message: "",
		Time:    now(),
		Verdict: VerdictSuccess,
	})
}

// OKMsg writes a 200 OK JSON response with verdict "success" and a custom message.
func OKMsg(w http.ResponseWriter, data any, msg string) {
	JSON(w, http.StatusOK, Envelope{
		Data:    data,
		Message: msg,
		Time:    now(),
		Verdict: VerdictSuccess,
	})
}

// Created writes a 201 Created JSON response with verdict "success".
func Created(w http.ResponseWriter, data any) {
	JSON(w, http.StatusCreated, Envelope{
		Data:    data,
		Message: "",
		Time:    now(),
		Verdict: VerdictSuccess,
	})
}

// NoContent writes a 204 No Content response with no body.
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// BadRequest writes a 400 Bad Request JSON error response.
func BadRequest(w http.ResponseWriter, msg string) {
	JSON(w, http.StatusBadRequest, Envelope{
		Data:    nil,
		Message: msg,
		Time:    now(),
		Verdict: VerdictError,
	})
}

// Unauthorized writes a 401 Unauthorized JSON error response.
func Unauthorized(w http.ResponseWriter, msg string) {
	JSON(w, http.StatusUnauthorized, Envelope{
		Data:    nil,
		Message: msg,
		Time:    now(),
		Verdict: VerdictError,
	})
}

// Forbidden writes a 403 Forbidden JSON error response.
func Forbidden(w http.ResponseWriter, msg string) {
	JSON(w, http.StatusForbidden, Envelope{
		Data:    nil,
		Message: msg,
		Time:    now(),
		Verdict: VerdictError,
	})
}

// NotFound writes a 404 Not Found JSON error response.
func NotFound(w http.ResponseWriter, msg string) {
	JSON(w, http.StatusNotFound, Envelope{
		Data:    nil,
		Message: msg,
		Time:    now(),
		Verdict: VerdictError,
	})
}

// UnprocessableEntity writes a 422 Unprocessable Entity JSON error response.
func UnprocessableEntity(w http.ResponseWriter, msg string) {
	JSON(w, http.StatusUnprocessableEntity, Envelope{
		Data:    nil,
		Message: msg,
		Time:    now(),
		Verdict: VerdictError,
	})
}

// InternalServerError writes a 500 Internal Server Error JSON error response.
func InternalServerError(w http.ResponseWriter, msg string) {
	JSON(w, http.StatusInternalServerError, Envelope{
		Data:    nil,
		Message: msg,
		Time:    now(),
		Verdict: VerdictError,
	})
}
