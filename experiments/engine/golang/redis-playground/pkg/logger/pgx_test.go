package logger

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/tracelog"
	"github.com/stretchr/testify/require"
)

func TestPGXLogger_ImplementsTraceLogger(t *testing.T) {
	t.Parallel()

	var _ tracelog.Logger = NewPGXLogger(New("debug"), 500*time.Millisecond)
}

func TestPGXLogger_Log(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		level tracelog.LogLevel
		data  map[string]any
	}{
		{
			name:  "trace",
			level: tracelog.LogLevelTrace,
			data:  nil,
		},
		{
			name:  "debug",
			level: tracelog.LogLevelDebug,
			data: map[string]any{
				"sql": "SELECT 1",
			},
		},
		{
			name:  "info",
			level: tracelog.LogLevelInfo,
			data: map[string]any{
				"rows": 10,
			},
		},
		{
			name:  "warn",
			level: tracelog.LogLevelWarn,
			data: map[string]any{
				"elapsed": "2s",
			},
		},
		{
			name:  "error",
			level: tracelog.LogLevelError,
			data: map[string]any{
				"err": "connection timeout",
			},
		},
		{
			name:  "empty map",
			level: tracelog.LogLevelDebug,
			data:  map[string]any{},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			logger := NewPGXLogger(New("debug"), 500*time.Millisecond)

			require.NotPanics(t, func() {
				logger.Log(
					context.Background(),
					tt.level,
					"test message",
					tt.data,
				)
			})
		})
	}
}
