package w2db

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/huandu/go-sqlbuilder"
)

// InsertOptions configures Insert and InsertContext.
type InsertOptions struct {
	// Into is the trusted table name or insert target.
	Into string

	// Values lists values keyed by trusted column names.
	Values map[string]any

	// Flavor overrides the package default SQL dialect when non-zero.
	Flavor sqlbuilder.Flavor

	// Logger overrides the package default SQL logger when non-nil.
	Logger *slog.Logger
}

// Insert inserts one row using context.Background and returns LastInsertId.
func Insert(db QueryExecer, opts InsertOptions) (int, error) {
	return InsertContext(context.Background(), db, opts)
}

// InsertContext inserts one row and returns LastInsertId.
func InsertContext(ctx context.Context, db QueryExecer, opts InsertOptions) (int, error) {
	if opts.Into == "" {
		return 0, errors.New("opts.Into is required")
	}

	if len(opts.Values) == 0 {
		return 0, errors.New("opts.Values is required")
	}

	flavor := opts.Flavor
	if flavor == 0 {
		flavor = defaultFlavor
	}

	logger := opts.Logger
	if logger == nil {
		logger = defaultLogger
	}

	cols := make([]string, 0, len(opts.Values))
	values := make([]any, 0, len(opts.Values))
	for k, v := range opts.Values {
		cols = append(cols, k)
		values = append(values, v)
	}

	builder := sqlbuilder.InsertInto(opts.Into)
	builder.Cols(cols...)
	builder.Values(values...)

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
