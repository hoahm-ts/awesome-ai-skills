// Package middleware provides reusable HTTP middleware for the chi router.
package middleware

import (
	"net/http"
	"time"

	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
)

// RequestLogger returns a chi-compatible middleware that logs every HTTP request
// using zerolog structured fields.
//
// Fields emitted per request:
//   - marker      "[api]" — fixed tag to enable fast log filtering
//   - method      HTTP verb
//   - path        request URL path
//   - status      response status code
//   - duration    elapsed time
//   - request_id  value of the X-Request-ID header (if present)
func RequestLogger(log zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := chiMiddleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r)

			log.Info().
				Str("marker", "[api]").
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Int("status", ww.Status()).
				Dur("duration", time.Since(start)).
				Str("request_id", r.Header.Get("X-Request-ID")).
				Msg("api request")
		})
	}
}
