// Package response provides helpers for writing consistent JSON HTTP responses.
package response

import (
	"encoding/json"
	"net/http"
)

// Envelope is the standard response wrapper returned by all API endpoints.
type Envelope struct {
	Data  any    `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

// JSON writes a JSON-encoded body with the given HTTP status code.
func JSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "encoding error", http.StatusInternalServerError)
	}
}

// OK writes a 200 OK JSON response.
func OK(w http.ResponseWriter, data any) {
	JSON(w, http.StatusOK, Envelope{Data: data})
}

// Created writes a 201 Created JSON response.
func Created(w http.ResponseWriter, data any) {
	JSON(w, http.StatusCreated, Envelope{Data: data})
}

// NoContent writes a 204 No Content response with no body.
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// BadRequest writes a 400 Bad Request JSON error response.
func BadRequest(w http.ResponseWriter, msg string) {
	JSON(w, http.StatusBadRequest, Envelope{Error: msg})
}

// Unauthorized writes a 401 Unauthorized JSON error response.
func Unauthorized(w http.ResponseWriter, msg string) {
	JSON(w, http.StatusUnauthorized, Envelope{Error: msg})
}

// Forbidden writes a 403 Forbidden JSON error response.
func Forbidden(w http.ResponseWriter, msg string) {
	JSON(w, http.StatusForbidden, Envelope{Error: msg})
}

// NotFound writes a 404 Not Found JSON error response.
func NotFound(w http.ResponseWriter, msg string) {
	JSON(w, http.StatusNotFound, Envelope{Error: msg})
}

// UnprocessableEntity writes a 422 Unprocessable Entity JSON error response.
func UnprocessableEntity(w http.ResponseWriter, msg string) {
	JSON(w, http.StatusUnprocessableEntity, Envelope{Error: msg})
}

// InternalServerError writes a 500 Internal Server Error JSON error response.
func InternalServerError(w http.ResponseWriter, msg string) {
	JSON(w, http.StatusInternalServerError, Envelope{Error: msg})
}
