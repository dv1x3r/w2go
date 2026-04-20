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

type GetGridOptions[T any] struct {
	From           string
	Select         []string
	CountExpr      string
	WhereMapping   map[string]string
	OrderByMapping map[string]string
	Flavor         sqlbuilder.Flavor
	BuildBase      func(sb *sqlbuilder.SelectBuilder)
	BuildSelect    func(sb *sqlbuilder.SelectBuilder)
	Scan           func(rows *sql.Rows) (T, error)
	Logger         *slog.Logger
}

func GetGrid[T any](db QueryDB, req w2.GetGridRequest, opts GetGridOptions[T]) (w2.GetGridResponse[T], error) {
	return GetGridContext(context.Background(), db, req, opts)
}

func GetGridContext[T any](ctx context.Context, db QueryDB, req w2.GetGridRequest, opts GetGridOptions[T]) (w2.GetGridResponse[T], error) {
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
		flavor = sqlbuilder.DefaultFlavor
	}

	logger := opts.Logger
	if logger == nil {
		logger = defaultLogger
	}

	var total int
	var records []T

	countBuilder := sqlbuilder.Select(countExpr).From(opts.From)
	countBuilder.SetFlavor(flavor)
	if opts.BuildBase != nil {
		opts.BuildBase(countBuilder)
	}

	w2sql.Where(countBuilder, req, opts.WhereMapping)
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
	if opts.BuildBase != nil {
		opts.BuildBase(dataBuilder)
	}
	if opts.BuildSelect != nil {
		opts.BuildSelect(dataBuilder)
	}

	w2sql.Where(dataBuilder, req, opts.WhereMapping)
	w2sql.OrderBy(dataBuilder, req, opts.OrderByMapping)
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
		record, err := opts.Scan(rows)
		if err != nil {
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
