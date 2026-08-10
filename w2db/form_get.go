package w2db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/dv1x3r/w2go/w2"
	"github.com/huandu/go-sqlbuilder"
)

// GetFormOptions configures GetForm and GetFormContext.
type GetFormOptions[T any] struct {
	// From is the table, view, or join expression used in the FROM clause.
	From string

	// IDField is the trusted SQL expression compared to req.RecID.
	IDField string

	// Select lists the SQL expressions returned for the form record.
	Select []string

	// Build customizes the SELECT query, for example by adding joins.
	Build func(sb *sqlbuilder.SelectBuilder)

	// Scan copies the selected row into record.
	Scan func(row *sql.Row, record *T) error

	// Flavor overrides the package default SQL dialect when non-zero.
	Flavor sqlbuilder.Flavor

	// Logger overrides the package default SQL logger when non-nil.
	Logger *slog.Logger
}

// GetForm loads one form record using context.Background.
func GetForm[T any](db QueryExecer, req w2.GetFormRequest, opts GetFormOptions[T]) (w2.GetFormResponse[T], error) {
	return GetFormContext(context.Background(), db, req, opts)
}

// GetFormContext loads one form record by req.RecID.
func GetFormContext[T any](ctx context.Context, db QueryExecer, req w2.GetFormRequest, opts GetFormOptions[T]) (w2.GetFormResponse[T], error) {
	if opts.From == "" {
		return w2.GetFormResponse[T]{}, errors.New("opts.From is required")
	}

	if opts.IDField == "" {
		return w2.GetFormResponse[T]{}, errors.New("opts.IDField is required")
	}

	if len(opts.Select) == 0 {
		return w2.GetFormResponse[T]{}, errors.New("opts.Select is required")
	}

	if opts.Scan == nil {
		return w2.GetFormResponse[T]{}, errors.New("opts.Scan is required")
	}

	flavor := opts.Flavor
	if flavor == 0 {
		flavor = defaultFlavor
	}

	logger := opts.Logger
	if logger == nil {
		logger = defaultLogger
	}

	builder := sqlbuilder.Select(opts.Select...).From(opts.From)
	if opts.Build != nil {
		opts.Build(builder)
	}

	builder.Where(builder.EQ(opts.IDField, req.RecID))
	query, args := builder.BuildWithFlavor(flavor)

	var record T

	begin := time.Now()
	row := db.QueryRowContext(ctx, query, args...)
	err := opts.Scan(row, &record)
	traceSQL(ctx, logger, begin, query, args, err)
	if err != nil {
		return w2.GetFormResponse[T]{}, fmt.Errorf("scan: %w", err)
	}

	return w2.NewGetFormResponse(record), nil
}
