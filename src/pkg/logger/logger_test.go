package logger_test

import (
	"context"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	"github.com/hoahm-ts/awesome-ai-skills/pkg/logger"
)

func TestNew_ReturnsLogger(t *testing.T) {
	t.Parallel()

	l := logger.New("test-service", "test", zerolog.InfoLevel)
	// A zero-value zerolog.Logger has a nil context; the logger returned by New
	// should be fully initialised and usable without panicking.
	l.Info().Msg("ok")
}

func TestWithContext_StoresLogger(t *testing.T) {
	t.Parallel()

	l := logger.New("test-service", "test", zerolog.DebugLevel)
	ctx := logger.WithContext(context.Background(), l)

	got := zerolog.Ctx(ctx)
	require.NotNil(t, got)
}

func TestLevelFromString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		give  string
		want  zerolog.Level
	}{
		{name: "debug", give: "debug", want: zerolog.DebugLevel},
		{name: "info", give: "info", want: zerolog.InfoLevel},
		{name: "warn", give: "warn", want: zerolog.WarnLevel},
		{name: "error", give: "error", want: zerolog.ErrorLevel},
		{name: "trace", give: "trace", want: zerolog.TraceLevel},
		{name: "unknown falls back to info", give: "invalid-level", want: zerolog.InfoLevel},
		{name: "empty string returns NoLevel", give: "", want: zerolog.NoLevel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := logger.LevelFromString(tt.give)
			require.Equal(t, tt.want, got)
		})
	}
}
