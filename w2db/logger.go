package w2db

import (
	"context"
	"log/slog"
	"time"
)

var defaultLogger *slog.Logger
var slowThreshold = 250 * time.Millisecond

// SetLogger sets the package default logger for SQL traces.
//
// Passing nil disables package-level logging. Individual option structs can
// still provide their own Logger.
func SetLogger(logger *slog.Logger) {
	defaultLogger = logger
}

// SetSlowThreshold sets the duration at which successful SQL traces are logged
// at Warn level instead of Debug level.
//
// Set threshold to zero or a negative value to log successful queries at Debug
// level regardless of duration.
func SetSlowThreshold(threshold time.Duration) {
	slowThreshold = threshold
}

func traceSQL(ctx context.Context, logger *slog.Logger, begin time.Time, query string, args []any, err error) {
	if logger == nil {
		return
	}

	elapsed := time.Since(begin)
	logArgs := []any{
		slog.String("sql", query),
		"args", append([]any(nil), args...),
		slog.Duration("elapsed", elapsed),
	}

	if err != nil {
		logger.ErrorContext(ctx, "w2db", append(logArgs, "err", err)...)
		return
	}

	if slowThreshold > 0 && elapsed >= slowThreshold {
		logger.WarnContext(ctx, "w2db", logArgs...)
		return
	}

	logger.DebugContext(ctx, "w2db", logArgs...)
}
