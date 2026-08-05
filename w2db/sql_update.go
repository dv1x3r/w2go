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

	// Values lists values keyed by trusted column names.
	//
	// Values that implement Providable are skipped when IsProvided returns false.
	Values map[string]any

	// Where lists equality conditions keyed by trusted SQL expressions.
	Where map[string]any

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

	if len(opts.Values) == 0 {
		return 0, errors.New("opts.Values is required")
	}

	if len(opts.Where) == 0 {
		return 0, errors.New("opts.Where is required")
	}

	flavor := opts.Flavor
	if flavor == 0 {
		flavor = defaultFlavor
	}

	logger := opts.Logger
	if logger == nil {
		logger = defaultLogger
	}

	builder := sqlbuilder.Update(opts.Update)
	assigned := 0

	for col, value := range opts.Values {
		if p, ok := value.(Providable); ok && !p.IsProvided() {
			continue
		}
		builder.SetMore(builder.Assign(col, value))
		assigned++
	}

	if assigned == 0 {
		return 0, nil
	}

	for col, value := range opts.Where {
		builder.Where(builder.EQ(col, value))
	}

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
