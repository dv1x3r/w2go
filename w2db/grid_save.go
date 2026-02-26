package w2db

import (
	"context"
	"errors"
	"fmt"

	"github.com/dv1x3r/w2go/w2"
	"github.com/huandu/go-sqlbuilder"
)

type SaveGridOptions[T any] struct {
	Flavor      sqlbuilder.Flavor
	BuildUpdate func(change T) *sqlbuilder.UpdateBuilder
}

func SaveGrid[T any](db ExecDB, req w2.SaveGridRequest[T], opts SaveGridOptions[T]) (int, error) {
	return SaveGridContext(context.Background(), db, req, opts)
}

func SaveGridContext[T any](ctx context.Context, db ExecDB, req w2.SaveGridRequest[T], opts SaveGridOptions[T]) (int, error) {
	if opts.BuildUpdate == nil {
		return 0, errors.New("opts.BuildUpdate is required")
	}

	if len(req.Changes) == 0 {
		return 0, nil
	}

	flavor := opts.Flavor
	if flavor == 0 {
		flavor = sqlbuilder.DefaultFlavor
	}

	affected := 0

	for i, change := range req.Changes {
		builder := opts.BuildUpdate(change)
		query, args := builder.BuildWithFlavor(flavor)
		result, err := db.ExecContext(ctx, query, args...)
		if err != nil {
			return 0, fmt.Errorf("update [%d]: %w", i, err)
		}
		if n, err := result.RowsAffected(); err == nil {
			affected += int(n)
		}
	}

	return affected, nil
}
