package w2db

import (
	"context"
	"errors"
	"fmt"

	"github.com/dv1x3r/w2go/w2"
	"github.com/huandu/go-sqlbuilder"
)

type UpdateFormOptions struct {
	Update  string
	IDField string
	Cols    []string
	Values  []any
	Flavor  sqlbuilder.Flavor
}

func UpdateForm[T any](db ExecDB, req w2.SaveFormRequest[T], opts UpdateFormOptions) (int, error) {
	return UpdateFormContext(context.Background(), db, req, opts)
}

func UpdateFormContext[T any](ctx context.Context, db ExecDB, req w2.SaveFormRequest[T], opts UpdateFormOptions) (int, error) {
	if opts.Update == "" {
		return 0, errors.New("opts.Update is required")
	}

	if opts.IDField == "" {
		return 0, errors.New("opts.IDField is required")
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

	builder := sqlbuilder.Update(opts.Update)
	for i := range opts.Cols {
		builder.SetMore(builder.Assign(opts.Cols[i], opts.Values[i]))
	}

	builder.Where(builder.EQ(opts.IDField, req.RecID))

	query, args := builder.BuildWithFlavor(flavor)
	result, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("update: %w", err)
	}

	affected, _ := result.RowsAffected()
	return int(affected), nil
}
