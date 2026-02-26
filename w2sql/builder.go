package w2sql

import (
	"fmt"

	"github.com/dv1x3r/w2go/w2"
	"github.com/huandu/go-sqlbuilder"
)

// Set updates the field only if a value is provided.
// If the value is marked as valid, it sets the field to the provided value. Otherwise, it sets field to NULL.
func Set[T any](ub *sqlbuilder.UpdateBuilder, value w2.Field[T], field string) {
	if value.Provided {
		if value.Valid {
			ub.SetMore(ub.Assign(field, value.V))
		} else {
			ub.SetMore(ub.Assign(field, nil))
		}
	}
}

// SetNoNull updates the field only if a value is provided.
// Unlike Assign, it always sets the field to the provided value even if invalid, using the zero value instead of NULL.
func SetNoNull[T any](ub *sqlbuilder.UpdateBuilder, value w2.Field[T], field string) {
	if value.Provided {
		ub.SetMore(ub.Assign(field, value.V))
	}
}

// Limit sets the LIMIT in SELECT based on provided W2GridRequest.
func Limit(sb *sqlbuilder.SelectBuilder, r w2.GetGridRequest) {
	if r.Limit != 0 {
		sb.Limit(r.Limit)
	}
}

// Offset sets the LIMIT offset in SELECT based on provided W2GridRequest.
func Offset(sb *sqlbuilder.SelectBuilder, r w2.GetGridRequest) {
	if r.Offset != 0 {
		sb.Offset(r.Offset)
	}
}

// Where sets expressions of WHERE in SELECT based on provided W2GridRequest and field mapping.
func Where(sb *sqlbuilder.SelectBuilder, r w2.GetGridRequest, mapping map[string]string) {
	c := make([]string, 0, len(r.Search))

	for _, s := range r.Search {
		if field, ok := mapping[s.Field]; ok {
			switch s.Operator {
			case "=", "is":
				c = append(c, sb.EQ(field, s.Value))
			case ">":
				c = append(c, sb.GT(field, s.Value))
			case "<", "less":
				c = append(c, sb.LT(field, s.Value))
			case ">=", "more":
				c = append(c, sb.GTE(field, s.Value))
			case "<=":
				c = append(c, sb.LTE(field, s.Value))
			case "begins":
				c = append(c, sb.Like(field, fmt.Sprintf("%v%%", s.Value)))
			case "contains":
				c = append(c, sb.Like(field, fmt.Sprintf("%%%v%%", s.Value)))
			case "ends":
				c = append(c, sb.Like(field, fmt.Sprintf("%%%v", s.Value)))
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

// OrderBy sets columns of ORDER BY in SELECT based on provided W2GridRequest and field mapping.
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
