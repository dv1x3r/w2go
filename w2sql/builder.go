package w2sql

import (
	"fmt"

	"github.com/dv1x3r/w2go/w2"
	"github.com/huandu/go-sqlbuilder"
)

// Set adds an UPDATE assignment for field when value was provided by the client.
//
// A provided and valid value writes that value. A provided but invalid value
// writes SQL NULL through w2.Field's driver.Valuer implementation. A value that
// was not provided is skipped.
func Set[T any](ub *sqlbuilder.UpdateBuilder, value w2.Field[T], field string) {
	if value.Provided {
		ub.SetMore(ub.Assign(field, value))
	}
}

// Limit applies r.Limit to sb when the request includes a non-zero limit.
func Limit(sb *sqlbuilder.SelectBuilder, r w2.GetGridRequest) {
	if r.Limit != 0 {
		sb.Limit(r.Limit)
	}
}

// Offset applies r.Offset to sb when the request includes a non-zero offset.
func Offset(sb *sqlbuilder.SelectBuilder, r w2.GetGridRequest) {
	if r.Offset != 0 {
		sb.Offset(r.Offset)
	}
}

// Where applies w2grid search filters to sb using mapping as a field whitelist.
//
// Keys in mapping are client-side field names from w2.GridSearch.Field; values
// are trusted SQL column names or expressions. Search rules whose fields are not
// present in mapping are ignored.
func Where(sb *sqlbuilder.SelectBuilder, r w2.GetGridRequest, mapping map[string]string) {
	c := make([]string, 0, len(r.Search))

	for _, s := range r.Search {
		if field, ok := mapping[s.Field]; ok {
			switch s.Operator {
			case "=", "is":
				c = append(c, sb.EQ(field, s.Value))
			case ">", "more":
				c = append(c, sb.GT(field, s.Value))
			case "<", "less":
				c = append(c, sb.LT(field, s.Value))
			case ">=":
				c = append(c, sb.GTE(field, s.Value))
			case "<=":
				c = append(c, sb.LTE(field, s.Value))
			case "null", "is null":
				c = append(c, sb.IsNull(field))
			case "not null":
				c = append(c, sb.IsNotNull(field))
			case "begins":
				if s.Value != "" {
					if sb.Flavor() == sqlbuilder.SQLite {
						expr := sqlbuilder.Buildf("INSTR(LOWER(%v), LOWER(%v)) = 1", sqlbuilder.Raw(field), s.Value)
						c = append(c, sb.Var(expr))
					} else {
						c = append(c, sb.Like(field, fmt.Sprintf("%v%%", s.Value)))
					}
				}
			case "contains":
				if s.Value != "" {
					if sb.Flavor() == sqlbuilder.SQLite {
						expr := sqlbuilder.Buildf("INSTR(LOWER(%v), LOWER(%v)) > 0", sqlbuilder.Raw(field), s.Value)
						c = append(c, sb.Var(expr))
					} else {
						c = append(c, sb.Like(field, fmt.Sprintf("%%%v%%", s.Value)))
					}
				}
			case "ends":
				if s.Value != "" {
					if sb.Flavor() == sqlbuilder.SQLite {
						expr := sqlbuilder.Buildf(
							"INSTR(LOWER(%v), LOWER(%v)) = LENGTH(%v) - LENGTH(%v) + 1",
							sqlbuilder.Raw(field), s.Value,
							sqlbuilder.Raw(field), s.Value,
						)
						c = append(c, sb.Var(expr))
					} else {
						c = append(c, sb.Like(field, fmt.Sprintf("%%%v", s.Value)))
					}
				}
			case "between":
				if values, ok := s.Value.([]any); ok && len(values) == 2 {
					c = append(c, sb.Between(field, values[0], values[1]))
				}
			case "in":
				if values, ok := s.Value.([]any); ok {
					ids := make([]any, 0, len(values))
					for i := range values {
						if value, ok := values[i].(map[string]any); ok {
							ids = append(ids, value["id"])
						}
					}
					c = append(c, sb.In(field, ids...))
				}
			case "not in":
				if values, ok := s.Value.([]any); ok {
					ids := make([]any, 0, len(values))
					for i := range values {
						if value, ok := values[i].(map[string]any); ok {
							ids = append(ids, value["id"])
						}
					}
					c = append(c, sb.NotIn(field, ids...))
				}
			}
		}
	}

	if len(c) > 0 {
		if r.SearchLogic == "AND" {
			sb.Where(sb.And(c...))
		} else {
			sb.Where(sb.Or(c...))
		}
	}
}

// OrderBy applies w2grid sort rules to sb using mapping as a field whitelist.
//
// Sort rules whose fields are not present in mapping are ignored.
func OrderBy(sb *sqlbuilder.SelectBuilder, r w2.GetGridRequest, mapping map[string]string) {
	for _, s := range r.Sort {
		if field, ok := mapping[s.Field]; ok {
			if s.Direction == "desc" {
				sb.OrderByDesc(field)
			} else {
				sb.OrderByAsc(field)
			}
		}
	}
}
