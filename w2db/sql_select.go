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

// SelectOptions configures Select and SelectContext.
type SelectOptions[T any] struct {
	// Query is the raw SQL statement executed as is. It is required when Build is nil.
	Query string

	// Args lists the placeholder values bound to Query.
	Args []any

	// Build builds the SELECT query with go-sqlbuilder.
	//
	// It is required when Query is empty, and replaces Query and Args when non-nil.
	Build func(sb *sqlbuilder.SelectBuilder)

	// Scan copies the current data row into record.
	Scan func(rows *sql.Rows, record *T) error

	// Flavor overrides the package default SQL dialect when non-zero.
	//
	// It only applies to Build, raw Query placeholders are left untouched.
	Flavor sqlbuilder.Flavor

	// Logger overrides the package default SQL logger when non-nil.
	Logger *slog.Logger
}

// Select runs a SELECT query using context.Background and returns all scanned records.
func Select[T any](db QueryExecer, opts SelectOptions[T]) ([]T, error) {
	return SelectContext[T](context.Background(), db, opts)
}

// SelectContext runs a SELECT query and returns all scanned records.
func SelectContext[T any](ctx context.Context, db QueryExecer, opts SelectOptions[T]) ([]T, error) {
	if opts.Query == "" && opts.Build == nil {
		return nil, errors.New("opts.Query or opts.Build is required")
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

	query, args := opts.Query, opts.Args
	if opts.Build != nil {
		builder := sqlbuilder.NewSelectBuilder()
		opts.Build(builder)
		query, args = builder.BuildWithFlavor(flavor)
	}

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

// SelectRowOptions configures SelectRow and SelectRowContext.
type SelectRowOptions[T any] struct {
	// Query is the raw SQL statement executed as is. It is required when Build is nil.
	Query string

	// Args lists the placeholder values bound to Query.
	Args []any

	// Build builds the SELECT query with go-sqlbuilder.
	//
	// It is required when Query is empty, and replaces Query and Args when non-nil.
	Build func(sb *sqlbuilder.SelectBuilder)

	// Scan copies the data row into record.
	Scan func(row *sql.Row, record *T) error

	// Flavor overrides the package default SQL dialect when non-zero.
	//
	// It only applies to Build, raw Query placeholders are left untouched.
	Flavor sqlbuilder.Flavor

	// Logger overrides the package default SQL logger when non-nil.
	Logger *slog.Logger
}

// SelectRow runs a SELECT query using context.Background and returns the first
// record, reporting whether a row was found.
func SelectRow[T any](db QueryExecer, opts SelectRowOptions[T]) (T, bool, error) {
	return SelectRowContext[T](context.Background(), db, opts)
}

// SelectRowContext runs a SELECT query and returns the first record, reporting
// whether a row was found. Missing rows are not an error.
func SelectRowContext[T any](ctx context.Context, db QueryExecer, opts SelectRowOptions[T]) (T, bool, error) {
	var record T

	if opts.Query == "" && opts.Build == nil {
		return record, false, errors.New("opts.Query or opts.Build is required")
	}

	if opts.Scan == nil {
		return record, false, errors.New("opts.Scan is required")
	}

	flavor := opts.Flavor
	if flavor == 0 {
		flavor = defaultFlavor
	}

	logger := opts.Logger
	if logger == nil {
		logger = defaultLogger
	}

	query, args := opts.Query, opts.Args
	if opts.Build != nil {
		builder := sqlbuilder.NewSelectBuilder()
		opts.Build(builder)
		query, args = builder.BuildWithFlavor(flavor)
	}

	begin := time.Now()
	row := db.QueryRowContext(ctx, query, args...)
	if err := opts.Scan(row, &record); err != nil {
		traceSQL(ctx, logger, begin, query, args, err)
		if err != sql.ErrNoRows {
			return record, false, fmt.Errorf("scan: %w", err)
		} else {
			return record, false, nil
		}
	}

	traceSQL(ctx, logger, begin, query, args, nil)
	return record, true, nil
}
