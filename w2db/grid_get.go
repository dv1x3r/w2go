package w2db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/dv1x3r/w2go/w2"
	"github.com/dv1x3r/w2go/w2sql"
	"github.com/huandu/go-sqlbuilder"
)

// GetGridOptions configures GetGrid and GetGridContext.
type GetGridOptions[T any] struct {
	// From is the table, view, or join expression used in the FROM clause.
	From string

	// Select lists the SQL expressions returned for each grid row.
	Select []string

	// CountExpr is the aggregate used for the total row count. It defaults to "count(*)".
	CountExpr string

	// Where maps w2grid search field names to trusted SQL expressions.
	Where map[string]string

	// OrderBy maps w2grid sort field names to trusted SQL expressions.
	OrderBy map[string]string

	// Build customizes the SELECT query, for example by adding joins or fixed filters.
	Build func(sb *sqlbuilder.SelectBuilder)

	// Scan copies the current data row into record.
	Scan func(rows *sql.Rows, record *T) error

	// Flavor overrides the package default SQL dialect when non-zero.
	Flavor sqlbuilder.Flavor

	// Logger overrides the package default SQL logger when non-nil.
	Logger *slog.Logger
}

// GetGrid loads records for a w2grid request using context.Background.
func GetGrid[T any](db QueryExecer, req w2.GetGridRequest, opts GetGridOptions[T]) (w2.GetGridResponse[T], error) {
	return GetGridContext(context.Background(), db, req, opts)
}

// GetGridContext loads a filtered, sorted, and paginated w2grid response.
//
// It runs one count query for the filtered total and one data query for the
// current page. Search and sort fields are ignored unless they exist in the
// corresponding mapping.
func GetGridContext[T any](ctx context.Context, db QueryExecer, req w2.GetGridRequest, opts GetGridOptions[T]) (w2.GetGridResponse[T], error) {
	if opts.From == "" {
		return w2.GetGridResponse[T]{}, errors.New("opts.From is required")
	}

	if len(opts.Select) == 0 {
		return w2.GetGridResponse[T]{}, errors.New("opts.Select is required")
	}

	if opts.Scan == nil {
		return w2.GetGridResponse[T]{}, errors.New("opts.Scan is required")
	}

	countExpr := opts.CountExpr
	if countExpr == "" {
		countExpr = "count(*)"
	}

	flavor := opts.Flavor
	if flavor == 0 {
		flavor = defaultFlavor
	}

	logger := opts.Logger
	if logger == nil {
		logger = defaultLogger
	}

	var total int
	var records []T

	countBuilder := sqlbuilder.Select(countExpr).From(opts.From)
	countBuilder.SetFlavor(flavor)
	if opts.Build != nil {
		opts.Build(countBuilder)
	}

	w2sql.Where(countBuilder, req, opts.Where)
	query, args := countBuilder.Build()

	begin := time.Now()
	row := db.QueryRowContext(ctx, query, args...)
	err := row.Scan(&total)
	traceSQL(ctx, logger, begin, query, args, err)
	if errors.Is(err, sql.ErrNoRows) {
		return w2.NewGetGridResponse(records, 0), nil
	} else if err != nil {
		return w2.GetGridResponse[T]{}, err
	}

	dataBuilder := sqlbuilder.Select(opts.Select...).From(opts.From)
	dataBuilder.SetFlavor(flavor)
	if opts.Build != nil {
		opts.Build(dataBuilder)
	}

	w2sql.Where(dataBuilder, req, opts.Where)
	w2sql.OrderBy(dataBuilder, req, opts.OrderBy)
	w2sql.Limit(dataBuilder, req)
	w2sql.Offset(dataBuilder, req)
	query, args = dataBuilder.Build()

	begin = time.Now()
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		traceSQL(ctx, logger, begin, query, args, err)
		return w2.GetGridResponse[T]{}, err
	}
	defer rows.Close()

	capacity := total
	if req.Limit > 0 {
		capacity = min(total, req.Limit)
	}
	records = make([]T, 0, capacity)

	for rows.Next() {
		var record T
		if err := opts.Scan(rows, &record); err != nil {
			traceSQL(ctx, logger, begin, query, args, err)
			return w2.GetGridResponse[T]{}, fmt.Errorf("scan: %w", err)
		}
		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		traceSQL(ctx, logger, begin, query, args, err)
		return w2.GetGridResponse[T]{}, err
	}

	traceSQL(ctx, logger, begin, query, args, nil)
	return w2.NewGetGridResponse(records, total), nil
}
