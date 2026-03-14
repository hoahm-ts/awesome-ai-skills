package response_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hoahm-ts/awesome-ai-skills/pkg/response"
)

// decodeEnvelope reads and decodes the JSON body of the recorder into an Envelope.
func decodeEnvelope(t *testing.T, rec *httptest.ResponseRecorder) response.Envelope {
	t.Helper()
	body, err := io.ReadAll(rec.Body)
	require.NoError(t, err)

	var env response.Envelope
	require.NoError(t, json.Unmarshal(body, &env))
	return env
}

func TestOK(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	response.OK(rec, map[string]string{"key": "value"})

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	env := decodeEnvelope(t, rec)
	require.Equal(t, response.VerdictSuccess, env.Verdict)
	require.Empty(t, env.Message)
	require.NotEmpty(t, env.Time)
}

func TestOKMsg(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	response.OKMsg(rec, struct{}{}, "pong")

	require.Equal(t, http.StatusOK, rec.Code)

	env := decodeEnvelope(t, rec)
	require.Equal(t, response.VerdictSuccess, env.Verdict)
	require.Equal(t, "pong", env.Message)
}

func TestCreated(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	response.Created(rec, map[string]int{"id": 1})

	require.Equal(t, http.StatusCreated, rec.Code)

	env := decodeEnvelope(t, rec)
	require.Equal(t, response.VerdictSuccess, env.Verdict)
	require.Empty(t, env.Message)
}

func TestNoContent(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	response.NoContent(rec)

	require.Equal(t, http.StatusNoContent, rec.Code)
	require.Empty(t, rec.Body.String())
}

func TestErrorResponses(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		giveMsg    string
		call       func(w http.ResponseWriter, msg string)
		wantStatus int
	}{
		{
			name:       "BadRequest",
			giveMsg:    "bad input",
			call:       response.BadRequest,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Unauthorized",
			giveMsg:    "not logged in",
			call:       response.Unauthorized,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "Forbidden",
			giveMsg:    "no permission",
			call:       response.Forbidden,
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "NotFound",
			giveMsg:    "resource missing",
			call:       response.NotFound,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "UnprocessableEntity",
			giveMsg:    "validation failed",
			call:       response.UnprocessableEntity,
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "InternalServerError",
			giveMsg:    "unexpected failure",
			call:       response.InternalServerError,
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rec := httptest.NewRecorder()
			tt.call(rec, tt.giveMsg)

			require.Equal(t, tt.wantStatus, rec.Code)
			require.Equal(t, "application/json", rec.Header().Get("Content-Type"))

			env := decodeEnvelope(t, rec)
			require.Equal(t, response.VerdictError, env.Verdict)
			require.Equal(t, tt.giveMsg, env.Message)
			require.NotEmpty(t, env.Time)
		})
	}
}

func TestJSON_SetsContentType(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	env := response.Envelope{
		Verdict: response.VerdictSuccess,
		Message: "test",
		Time:    "2025-01-01T00:00:00Z",
		Data:    nil,
	}
	response.JSON(rec, http.StatusOK, env)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "application/json", rec.Header().Get("Content-Type"))
}
