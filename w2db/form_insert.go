package w2db

import (
	"context"
	"errors"
	"fmt"

	"github.com/dv1x3r/w2go/w2"
	"github.com/huandu/go-sqlbuilder"
)

type InsertFormOptions struct {
	Into   string
	Cols   []string
	Values []any
	Flavor sqlbuilder.Flavor
}

func InsertForm[T any](db ExecDB, req w2.SaveFormRequest[T], opts InsertFormOptions) (int, error) {
	return InsertFormContext(context.Background(), db, req, opts)
}

func InsertFormContext[T any](ctx context.Context, db ExecDB, req w2.SaveFormRequest[T], opts InsertFormOptions) (int, error) {
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
		flavor = sqlbuilder.DefaultFlavor
	}

	builder := sqlbuilder.InsertInto(opts.Into)
	builder.Cols(opts.Cols...)
	builder.Values(opts.Values...)

	query, args := builder.BuildWithFlavor(flavor)
	result, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("insert: %w", err)
	}

	lastInsertID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("last insert id: %w", err)
	}

	return int(lastInsertID), nil
}
