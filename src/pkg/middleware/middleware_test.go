package middleware_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	"github.com/hoahm-ts/awesome-ai-skills/pkg/middleware"
)

func TestRequestLogger_ForwardsToNextHandler(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	log := zerolog.New(&buf)

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.RequestLogger(log)(next)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestRequestLogger_WritesStructuredLogEntry(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	log := zerolog.New(&buf)

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})

	handler := middleware.RequestLogger(log)(next)

	req := httptest.NewRequest(http.MethodPost, "/items", nil)
	req.Header.Set("X-Request-ID", "req-123")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	logOutput := buf.String()
	require.Contains(t, logOutput, "[api]")
	require.Contains(t, logOutput, "POST")
	require.Contains(t, logOutput, "/items")
	require.Contains(t, logOutput, "req-123")
}

func TestRequestLogger_LogsStatusCode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		giveStatus int
	}{
		{name: "200 OK", giveStatus: http.StatusOK},
		{name: "404 Not Found", giveStatus: http.StatusNotFound},
		{name: "500 Internal Server Error", giveStatus: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			log := zerolog.New(&buf)

			next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.giveStatus)
			})

			handler := middleware.RequestLogger(log)(next)
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			require.Equal(t, tt.giveStatus, rec.Code)
			require.Contains(t, buf.String(), "api request")
		})
	}
}
