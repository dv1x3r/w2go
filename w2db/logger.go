package w2db

import (
	"context"
	"log/slog"
	"time"
)

var defaultLogger *slog.Logger
var slowThreshold = 250 * time.Millisecond

func SetLogger(logger *slog.Logger) {
	defaultLogger = logger
}

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
