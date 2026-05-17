package w2db

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/huandu/go-sqlbuilder"
)

// UpdateOptions configures Update and UpdateContext.
type UpdateOptions struct {
	// Update is the trusted table name or update target.
	Update string

	// Cols lists the columns to assign.
	Cols []string

	// Values lists the values assigned to Cols.
	//
	// Values that implement Providable are skipped when IsProvided returns false.
	Values []any

	// IDField is the trusted SQL expression used to locate the row.
	IDField string

	// IDValue is the value compared with IDField.
	IDValue any

	// Flavor overrides the package default SQL dialect when non-zero.
	Flavor sqlbuilder.Flavor

	// Logger overrides the package default SQL logger when non-nil.
	Logger *slog.Logger
}

// Update updates one row using context.Background and returns RowsAffected.
func Update(db QueryExecer, opts UpdateOptions) (int, error) {
	return UpdateContext(context.Background(), db, opts)
}

// UpdateContext updates one row and returns RowsAffected.
//
// If every value is a Providable value that was not provided, no SQL statement
// is executed and the function returns zero rows affected.
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
