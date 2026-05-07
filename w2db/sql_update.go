package w2db

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/huandu/go-sqlbuilder"
)

type UpdateOptions struct {
	Update  string
	Cols    []string
	Values  []any
	IDField string
	IDValue any
	Flavor  sqlbuilder.Flavor
	Logger  *slog.Logger
}

func Update(db QueryExecer, opts UpdateOptions) (int, error) {
	return UpdateContext(context.Background(), db, opts)
}

func UpdateContext(ctx context.Context, db QueryExecer, opts UpdateOptions) (int, error) {
	if opts.Update == "" {
		return 0, errors.New("opts.Update is required")
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

	if opts.IDField == "" {
		return 0, errors.New("opts.IDField is required")
	}

	if opts.IDValue == nil {
		return 0, errors.New("opts.IDValue is required")
	}

	flavor := opts.Flavor
	if flavor == 0 {
		flavor = defaultFlavor
	}

	logger := opts.Logger
	if logger == nil {
		logger = defaultLogger
	}

	var assigned int

	builder := sqlbuilder.Update(opts.Update)
	for i := range opts.Cols {
		if f, ok := opts.Values[i].(Providable); ok {
			if f.IsProvided() {
				builder.SetMore(builder.Assign(opts.Cols[i], opts.Values[i]))
				assigned++
			}
		} else {
			builder.SetMore(builder.Assign(opts.Cols[i], opts.Values[i]))
			assigned++
		}
	}

	if assigned == 0 {
		return 0, nil
	}

	builder.Where(builder.EQ(opts.IDField, opts.IDValue))
	query, args := builder.BuildWithFlavor(flavor)

	begin := time.Now()
	result, err := db.ExecContext(ctx, query, args...)
	traceSQL(ctx, logger, begin, query, args, err)
	if err != nil {
		return 0, fmt.Errorf("update: %w", err)
	}

	affected, _ := result.RowsAffected()
	return int(affected), nil
}
