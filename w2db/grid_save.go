package w2db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/dv1x3r/w2go/w2"
)

type SaveGridOptions[T any] struct {
	BuildOptions func(change T) UpdateOptions
}

func SaveGrid[T any](db QueryExecer, req w2.SaveGridRequest[T], opts SaveGridOptions[T]) (int, error) {
	return SaveGridContext(context.Background(), db, req, opts)
}

func SaveGridContext[T any](ctx context.Context, db QueryExecer, req w2.SaveGridRequest[T], opts SaveGridOptions[T]) (int, error) {
	// save requires a transaction for multiple row update,
	// but SQLite does not support nested transactions,
	// so begin one if db is not already a *sql.Tx transaction
	if sqlDB, ok := db.(*sql.DB); ok {
		// fmt.Println("begin tx")
		tx, err := sqlDB.BeginTx(ctx, nil)
		if err != nil {
			return 0, err
		}
		defer tx.Rollback()

		affected, err := saveGridContext(ctx, tx, req, opts)
		if err != nil {
			return 0, err
		}

		return affected, tx.Commit()
	} else {
		// db is already a *sql.Tx transaction
		return saveGridContext(ctx, db, req, opts)
	}
}

func saveGridContext[T any](ctx context.Context, db QueryExecer, req w2.SaveGridRequest[T], opts SaveGridOptions[T]) (int, error) {
	if opts.BuildOptions == nil {
		return 0, errors.New("opts.BuildOptions is required")
	}

	if len(req.Changes) == 0 {
		return 0, nil
	}

	affected := 0

	for i, change := range req.Changes {
		updateOpts := opts.BuildOptions(change)
		n, err := UpdateContext(ctx, db, updateOpts)
		if err != nil {
			return 0, fmt.Errorf("update [%d]: %w", i, err)
		}
		affected += int(n)
	}

	return affected, nil
}
