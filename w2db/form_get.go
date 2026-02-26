package w2db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/dv1x3r/w2go/w2"
	"github.com/huandu/go-sqlbuilder"
)

type GetFormOptions[T any] struct {
	From        string
	IDField     string
	Select      []string
	Flavor      sqlbuilder.Flavor
	BuildSelect func(sb *sqlbuilder.SelectBuilder)
	Scan        func(row *sql.Row) (T, error)
}

func GetForm[T any](db QueryDB, req w2.GetFormRequest, opts GetFormOptions[T]) (w2.GetFormResponse[T], error) {
	return GetFormContext(context.Background(), db, req, opts)
}

func GetFormContext[T any](ctx context.Context, db QueryDB, req w2.GetFormRequest, opts GetFormOptions[T]) (w2.GetFormResponse[T], error) {
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
		flavor = sqlbuilder.DefaultFlavor
	}

	builder := sqlbuilder.Select(opts.Select...).From(opts.From)
	if opts.BuildSelect != nil {
		opts.BuildSelect(builder)
	}

	builder.Where(builder.EQ(opts.IDField, req.RecID))

	query, args := builder.BuildWithFlavor(flavor)
	row := db.QueryRowContext(ctx, query, args...)
	record, err := opts.Scan(row)
	if err != nil {
		return w2.GetFormResponse[T]{}, fmt.Errorf("scan: %w", err)
	}

	return w2.NewGetFormResponse(record), nil
}
