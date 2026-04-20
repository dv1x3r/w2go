package w2db

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/dv1x3r/w2go/w2"
	"github.com/huandu/go-sqlbuilder"
)

type RemoveGridOptions struct {
	From    string
	IDField string
	Flavor  sqlbuilder.Flavor
	Logger  *slog.Logger
}

func RemoveGrid(db ExecDB, req w2.RemoveGridRequest, opts RemoveGridOptions) (int, error) {
	return RemoveGridContext(context.Background(), db, req, opts)
}

func RemoveGridContext(ctx context.Context, db ExecDB, req w2.RemoveGridRequest, opts RemoveGridOptions) (int, error) {
	if opts.From == "" {
		return 0, errors.New("opts.From is required")
	}

	if opts.IDField == "" {
		return 0, errors.New("opts.IDField is required")
	}

	if len(req.ID) == 0 {
		return 0, errors.New("req.ID must not be empty")
	}

	flavor := opts.Flavor
	if flavor == 0 {
		flavor = sqlbuilder.DefaultFlavor
	}

	logger := opts.Logger
	if logger == nil {
		logger = defaultLogger
	}

	builder := sqlbuilder.DeleteFrom(opts.From)
	builder.Where(builder.In(opts.IDField, sqlbuilder.List(req.ID)))
	query, args := builder.BuildWithFlavor(flavor)

	begin := time.Now()
	result, err := db.ExecContext(ctx, query, args...)
	traceSQL(ctx, logger, begin, query, args, err)
	if err != nil {
		return 0, fmt.Errorf("delete: %w", err)
	}

	affected, _ := result.RowsAffected()
	return int(affected), nil
}
