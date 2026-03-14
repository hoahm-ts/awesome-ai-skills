// Package logger configures and provides a zerolog.Logger for the application.
// A single logger must be created in the composition root and propagated via
// context using WithContext / zerolog.Ctx(ctx).
package logger

import (
	"context"
	"io"
	"os"

	"github.com/rs/zerolog"
)

// New creates a configured zerolog.Logger instance.
// Call this once at the composition root and inject the result via DI.
func New(serviceName, env string, level zerolog.Level) zerolog.Logger {
	var out io.Writer = os.Stdout

	return zerolog.New(out).
		Level(level).
		With().
		Timestamp().
		Str("service", serviceName).
		Str("env", env).
		Logger()
}

// WithContext attaches l to ctx and returns the updated context.
// Downstream code retrieves the logger with zerolog.Ctx(ctx).
func WithContext(ctx context.Context, l zerolog.Logger) context.Context {
	return l.WithContext(ctx)
}

// LevelFromString parses a log level string and returns the matching zerolog.Level.
// Falls back to zerolog.InfoLevel when the string is unrecognised.
func LevelFromString(s string) zerolog.Level {
	l, err := zerolog.ParseLevel(s)
	if err != nil {
		return zerolog.InfoLevel
	}
	return l
}
