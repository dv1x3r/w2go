package w2db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/huandu/go-sqlbuilder"
)

type SelectOptions[T any] struct {
	Build  func(sb *sqlbuilder.SelectBuilder)
	Scan   func(rows *sql.Rows, record *T) error
	Flavor sqlbuilder.Flavor
	Logger *slog.Logger
}

func Select[T any](db QueryExecer, opts SelectOptions[T]) ([]T, error) {
	return SelectContext[T](context.Background(), db, opts)
}

func SelectContext[T any](ctx context.Context, db QueryExecer, opts SelectOptions[T]) ([]T, error) {
	if opts.Build == nil {
		return nil, errors.New("opts.Build is required")
	}

	if opts.Scan == nil {
		return nil, errors.New("opts.Scan is required")
	}

	flavor := opts.Flavor
	if flavor == 0 {
		flavor = defaultFlavor
	}

	logger := opts.Logger
	if logger == nil {
		logger = defaultLogger
	}

	builder := sqlbuilder.NewSelectBuilder()
	opts.Build(builder)
	query, args := builder.BuildWithFlavor(flavor)

	begin := time.Now()
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		traceSQL(ctx, logger, begin, query, args, err)
		return nil, err
	}
	defer rows.Close()

	var records []T

	for rows.Next() {
		var record T
		if err := opts.Scan(rows, &record); err != nil {
			traceSQL(ctx, logger, begin, query, args, err)
			return records, fmt.Errorf("scan: %w", err)
		}
		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		traceSQL(ctx, logger, begin, query, args, err)
		return records, err
	}

	traceSQL(ctx, logger, begin, query, args, nil)
	return records, nil
}
