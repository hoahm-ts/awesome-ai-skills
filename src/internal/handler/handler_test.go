package handler_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hoahm-ts/awesome-ai-skills/internal/handler"
	"github.com/hoahm-ts/awesome-ai-skills/pkg/response"
)

func TestPing_ReturnsOK(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()

	handler.Ping(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestPing_ContentTypeIsJSON(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()

	handler.Ping(rec, req)

	require.Equal(t, "application/json", rec.Header().Get("Content-Type"))
}

func TestPing_BodyContainsPong(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()

	handler.Ping(rec, req)

	body, err := io.ReadAll(rec.Body)
	require.NoError(t, err)

	var env response.Envelope
	require.NoError(t, json.Unmarshal(body, &env))
	require.Equal(t, response.VerdictSuccess, env.Verdict)
	require.Equal(t, "pong", env.Message)
}
