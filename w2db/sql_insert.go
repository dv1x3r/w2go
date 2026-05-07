package w2db

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/huandu/go-sqlbuilder"
)

type InsertOptions struct {
	Into   string
	Cols   []string
	Values []any
	Flavor sqlbuilder.Flavor
	Logger *slog.Logger
}

func Insert(db QueryExecer, opts InsertOptions) (int, error) {
	return InsertContext(context.Background(), db, opts)
}

func InsertContext(ctx context.Context, db QueryExecer, opts InsertOptions) (int, error) {
	if opts.Into == "" {
		return 0, errors.New("opts.Into is required")
	}

	if len(opts.Cols) == 0 {
		return 0, errors.New("opts.Cols is required")
	}

	if len(opts.Values) == 0 {
		return 0, errors.New("opts.Values is required")
	}

	if len(opts.Cols) != len(opts.Values) {
		return 0, errors.New("opts.Cols and opts.Values must have same length")
	}

	flavor := opts.Flavor
	if flavor == 0 {
		flavor = defaultFlavor
	}

	logger := opts.Logger
	if logger == nil {
		logger = defaultLogger
	}

	builder := sqlbuilder.InsertInto(opts.Into)
	builder.Cols(opts.Cols...)
	builder.Values(opts.Values...)
	query, args := builder.BuildWithFlavor(flavor)

	begin := time.Now()
	result, err := db.ExecContext(ctx, query, args...)
	traceSQL(ctx, logger, begin, query, args, err)
	if err != nil {
		return 0, fmt.Errorf("insert: %w", err)
	}

	lastInsertID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("last insert id: %w", err)
	}

	return int(lastInsertID), nil
}
