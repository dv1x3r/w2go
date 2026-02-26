package w2db

import (
	"context"
	"errors"
	"fmt"

	"github.com/dv1x3r/w2go/w2"
	"github.com/huandu/go-sqlbuilder"
)

type GetDropdownOptions struct {
	From         string
	IDField      string
	TextField    string
	OrderByField string
	Flavor       sqlbuilder.Flavor
}

func GetDropdown(db QueryDB, req w2.GetDropdownRequest, opts GetDropdownOptions) (w2.GetDropdownResponse[w2.Dropdown], error) {
	return GetDropdownContext(context.Background(), db, req, opts)
}

func GetDropdownContext(ctx context.Context, db QueryDB, req w2.GetDropdownRequest, opts GetDropdownOptions) (w2.GetDropdownResponse[w2.Dropdown], error) {
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
		flavor = sqlbuilder.DefaultFlavor
	}

	builder := sqlbuilder.Select(opts.IDField, opts.TextField).From(opts.From)
	builder.Where(builder.Like(opts.TextField, fmt.Sprintf("%%%s%%", req.Search)))
	builder.OrderBy(opts.OrderByField)
	builder.Limit(req.Max)

	query, args := builder.BuildWithFlavor(flavor)
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return w2.GetDropdownResponse[w2.Dropdown]{}, err
	}
	defer rows.Close()

	var records []w2.Dropdown

	for rows.Next() {
		var record w2.Dropdown
		if err := rows.Scan(&record.ID, &record.Text); err != nil {
			return w2.GetDropdownResponse[w2.Dropdown]{}, fmt.Errorf("scan: %w", err)
		}
		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return w2.GetDropdownResponse[w2.Dropdown]{}, err
	}

	res := w2.NewGetDropdownResponse(records)
	return res, nil
}
