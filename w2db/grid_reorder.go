package w2db

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/dv1x3r/w2go/w2"
	"github.com/dv1x3r/w2go/w2sort"
	"github.com/huandu/go-sqlbuilder"
)

type ReorderGridOptions struct {
	Update     string
	IDField    string
	SetField   string
	GroupField string
	Flavor     sqlbuilder.Flavor
	Logger     *slog.Logger
}

func ReorderGrid(db QueryExecDB, req w2.ReorderGridRequest, opts ReorderGridOptions) (int, error) {
	return ReorderGridContext(context.Background(), db, req, opts)
}

func ReorderGridContext(ctx context.Context, db QueryExecDB, req w2.ReorderGridRequest, opts ReorderGridOptions) (int, error) {
	if opts.Update == "" {
		return 0, errors.New("opts.Update is required")
	}

	if opts.IDField == "" {
		return 0, errors.New("opts.IDField is required")
	}

	if opts.SetField == "" {
		return 0, errors.New("opts.SetField is required")
	}

	flavor := opts.Flavor
	if flavor == 0 {
		flavor = defaultFlavor
	}

	logger := opts.Logger
	if logger == nil {
		logger = defaultLogger
	}

	selectBuilder := sqlbuilder.Select(opts.IDField).From(opts.Update)

	if opts.GroupField != "" {
		sub := sqlbuilder.Select(opts.GroupField).From(opts.Update)
		sub.Where(sub.EQ(opts.IDField, req.RecID))
		selectBuilder.Where(selectBuilder.In(opts.GroupField, sub))
	}

	selectBuilder.OrderByAsc(opts.SetField).OrderByDesc(opts.IDField)
	query, args := selectBuilder.BuildWithFlavor(flavor)

	begin := time.Now()
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		traceSQL(ctx, logger, begin, query, args, err)
		return 0, err
	}
	defer rows.Close()

	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			traceSQL(ctx, logger, begin, query, args, err)
			return 0, fmt.Errorf("scan: %w", err)
		}
		ids = append(ids, id)
	}

	if err := rows.Err(); err != nil {
		traceSQL(ctx, logger, begin, query, args, err)
		return 0, err
	}
	traceSQL(ctx, logger, begin, query, args, nil)

	if err := w2sort.ReorderArray(ids, req); err != nil {
		return 0, fmt.Errorf("reorder: %w", err)
	}

	whenClauses := make([]string, len(ids))
	for i, id := range ids {
		whenClauses[i] = fmt.Sprintf("WHEN %d THEN %d", id, i+1)
	}

	setClause := fmt.Sprintf("CASE %s %s ELSE %s END",
		opts.IDField, strings.Join(whenClauses, " "), opts.SetField,
	)

	updateBuilder := sqlbuilder.Update(opts.Update)
	updateBuilder.Set(updateBuilder.Assign(opts.SetField, sqlbuilder.Raw(setClause)))
	updateBuilder.Where(updateBuilder.In(opts.IDField, sqlbuilder.List(ids)))
	query, args = updateBuilder.BuildWithFlavor(flavor)

	begin = time.Now()
	result, err := db.ExecContext(ctx, query, args...)
	traceSQL(ctx, logger, begin, query, args, err)
	if err != nil {
		return 0, fmt.Errorf("update: %w", err)
	}

	affected, _ := result.RowsAffected()
	return int(affected), nil
}
