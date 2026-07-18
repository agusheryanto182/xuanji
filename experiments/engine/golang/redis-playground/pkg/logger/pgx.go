package logger

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/tracelog"
	"github.com/rs/zerolog"
)

type PGXLogger struct {
	logger             *Logger
	slowQueryThreshold time.Duration
}

var _ tracelog.Logger = (*PGXLogger)(nil)

func NewPGXLogger(
	logger *Logger,
	slowQueryThreshold time.Duration,
) *PGXLogger {
	return &PGXLogger{
		logger:             logger,
		slowQueryThreshold: slowQueryThreshold,
	}
}

func (l *PGXLogger) Log(
	_ context.Context,
	level tracelog.LogLevel,
	msg string,
	data map[string]any,
) {
	// Mapping pgx level -> zerolog level
	zl := toZeroLogLevel(level)

	// Detect slow query
	if msg == "Query" {
		if d, ok := data["time"].(time.Duration); ok &&
			l.slowQueryThreshold > 0 &&
			d >= l.slowQueryThreshold {

			zl = zerolog.WarnLevel
			msg = "Slow Query"

			data["slow_query"] = true
		}
	}

	event := l.logger.Event(zl)

	for k, v := range data {
		switch k {
		case "time":
			// pgx "time" = duration
			if d, ok := v.(time.Duration); ok {
				event = event.Str("duration", d.String())
				continue
			}
		}

		event = event.Interface(k, v)
	}

	event.
		Str("component", "pgx").
		Msg(msg)
}

func toZeroLogLevel(level tracelog.LogLevel) zerolog.Level {
	switch level {
	case tracelog.LogLevelTrace:
		return zerolog.TraceLevel
	case tracelog.LogLevelDebug:
		return zerolog.DebugLevel
	case tracelog.LogLevelInfo:
		return zerolog.InfoLevel
	case tracelog.LogLevelWarn:
		return zerolog.WarnLevel
	case tracelog.LogLevelError:
		return zerolog.ErrorLevel
	default:
		return zerolog.InfoLevel
	}
}
