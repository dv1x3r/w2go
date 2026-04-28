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

type SaveGridOptions[T any] struct {
	Flavor      sqlbuilder.Flavor
	BuildUpdate func(change T) *sqlbuilder.UpdateBuilder
	Logger      *slog.Logger
}

func SaveGrid[T any](db ExecDB, req w2.SaveGridRequest[T], opts SaveGridOptions[T]) (int, error) {
	return SaveGridContext(context.Background(), db, req, opts)
}

func SaveGridContext[T any](ctx context.Context, db ExecDB, req w2.SaveGridRequest[T], opts SaveGridOptions[T]) (int, error) {
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

func saveGridContext[T any](ctx context.Context, db ExecDB, req w2.SaveGridRequest[T], opts SaveGridOptions[T]) (int, error) {
	if opts.BuildUpdate == nil {
		return 0, errors.New("opts.BuildUpdate is required")
	}

	if len(req.Changes) == 0 {
		return 0, nil
	}

	flavor := opts.Flavor
	if flavor == 0 {
		flavor = defaultFlavor
	}

	logger := opts.Logger
	if logger == nil {
		logger = defaultLogger
	}

	affected := 0

	for i, change := range req.Changes {
		builder := opts.BuildUpdate(change)
		query, args := builder.BuildWithFlavor(flavor)
		begin := time.Now()
		result, err := db.ExecContext(ctx, query, args...)
		traceSQL(ctx, logger, begin, query, args, err)
		if err != nil {
			return 0, fmt.Errorf("update [%d]: %w", i, err)
		}
		if n, err := result.RowsAffected(); err == nil {
			affected += int(n)
		}
	}

	return affected, nil
}
