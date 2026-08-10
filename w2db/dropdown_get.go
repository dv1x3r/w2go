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

// GetDropdownOptions configures GetDropdown and GetDropdownContext.
type GetDropdownOptions struct {
	// From is the table, view, or join expression used in the FROM clause.
	From string

	// IDField is the trusted SQL expression returned as each option ID.
	IDField string

	// TextField is the trusted SQL expression returned as each option label.
	TextField string

	// OrderByField is the trusted SQL expression used to order options.
	OrderByField string

	// Build customizes the SELECT query, for example by adding joins or fixed filters.
	Build func(sb *sqlbuilder.SelectBuilder)

	// Flavor overrides the package default SQL dialect when non-zero.
	Flavor sqlbuilder.Flavor

	// Logger overrides the package default SQL logger when non-nil.
	Logger *slog.Logger
}

// GetDropdown loads dropdown options using context.Background.
func GetDropdown(db QueryExecer, req w2.GetDropdownRequest, opts GetDropdownOptions) (w2.GetDropdownResponse[w2.Dropdown], error) {
	return GetDropdownContext(context.Background(), db, req, opts)
}

// GetDropdownContext loads dropdown options and filters them by req.Search.
//
// The response records use w2.Dropdown, with ID and Text scanned as w2.Field
// values so nullable database values round-trip correctly.
func GetDropdownContext(ctx context.Context, db QueryExecer, req w2.GetDropdownRequest, opts GetDropdownOptions) (w2.GetDropdownResponse[w2.Dropdown], error) {
	if opts.From == "" {
		return w2.GetDropdownResponse[w2.Dropdown]{}, errors.New("opts.From is required")
	}

	if opts.IDField == "" {
		return w2.GetDropdownResponse[w2.Dropdown]{}, errors.New("opts.IDField is required")
	}

	if opts.TextField == "" {
		return w2.GetDropdownResponse[w2.Dropdown]{}, errors.New("opts.TextField is required")
	}

	if opts.OrderByField == "" {
		return w2.GetDropdownResponse[w2.Dropdown]{}, errors.New("opts.OrderByField is required")
	}

	flavor := opts.Flavor
	if flavor == 0 {
		flavor = defaultFlavor
	}

	logger := opts.Logger
	if logger == nil {
		logger = defaultLogger
	}

	builder := sqlbuilder.Select(opts.IDField, opts.TextField).From(opts.From)
	if opts.Build != nil {
		opts.Build(builder)
	}

	if req.Search != "" {
		if flavor == sqlbuilder.SQLite {
			expr := sqlbuilder.Buildf("INSTR(LOWER(%v), LOWER(%v)) > 0", sqlbuilder.Raw(opts.TextField), req.Search)
			builder.Where(builder.Var(expr))
		} else {
			builder.Where(builder.Like(opts.TextField, "%"+req.Search+"%"))
		}
	}

	builder.OrderBy(opts.OrderByField)
	builder.Limit(req.Max)
	query, args := builder.BuildWithFlavor(flavor)

	begin := time.Now()
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		traceSQL(ctx, logger, begin, query, args, err)
		return w2.GetDropdownResponse[w2.Dropdown]{}, err
	}
	defer rows.Close()

	var records []w2.Dropdown

	for rows.Next() {
		var record w2.Dropdown
		if err := rows.Scan(&record.ID, &record.Text); err != nil {
			traceSQL(ctx, logger, begin, query, args, err)
			return w2.GetDropdownResponse[w2.Dropdown]{}, fmt.Errorf("scan: %w", err)
		}
		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		traceSQL(ctx, logger, begin, query, args, err)
		return w2.GetDropdownResponse[w2.Dropdown]{}, err
	}

	traceSQL(ctx, logger, begin, query, args, nil)
	res := w2.NewGetDropdownResponse(records)
	return res, nil
}
