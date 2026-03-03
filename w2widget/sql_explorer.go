package w2widget

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/dv1x3r/w2go/w2"
)

type SQLExecRequest struct {
	Query string `json:"query"`
}

type SQLExecResult struct {
	Status  w2.Status        `json:"status"`
	Columns []string         `json:"columns"`
	Records []map[string]any `json:"records"`
	Total   int              `json:"total"`
}

func NewSQLExecResult(columns []string, records []map[string]any, total int) SQLExecResult {
	return SQLExecResult{
		Status:  w2.StatusSuccess,
		Columns: columns,
		Records: records,
		Total:   total,
	}
}

func (res SQLExecResult) Write(w http.ResponseWriter) error {
	data, err := json.Marshal(res)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	return err
}

func SQLExecHandler(db *sql.DB) func(w http.ResponseWriter, r *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		var req SQLExecRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return err
		}

		res, err := SQLExecQuery(r.Context(), db, req.Query)
		if err != nil {
			return err
		}

		return res.Write(w)
	}
}

func SQLExecHTTPHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req SQLExecRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			res := w2.NewErrorResponse(err.Error())
			res.Write(w, http.StatusBadRequest)
			return
		}

		res, err := SQLExecQuery(r.Context(), db, req.Query)
		if err != nil {
			res := w2.NewErrorResponse(err.Error())
			res.Write(w, http.StatusInternalServerError)
			return
		}

		res.Write(w)
	}
}

func SQLExecQuery(ctx context.Context, db *sql.DB, query string) (SQLExecResult, error) {
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return SQLExecResult{}, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return SQLExecResult{}, err
	}

	records := []map[string]any{}

	for rows.Next() {
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return SQLExecResult{}, err
		}

		record := map[string]any{}
		for i, column := range columns {
			record[column] = values[i]
		}

		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return SQLExecResult{}, err
	}

	return NewSQLExecResult(columns, records, len(records)), nil
}

func SQLiteSchemaHandler(db *sql.DB) func(w http.ResponseWriter, r *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		res, err := SQLiteSelectSchema(r.Context(), db)
		if err != nil {
			return err
		}

		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write([]byte(res))
		return err
	}
}

func SQLiteSchemaHTTPHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res, err := SQLiteSelectSchema(r.Context(), db)
		if err != nil {
			res := w2.NewErrorResponse(err.Error())
			res.Write(w, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(res))
	}
}

func SQLiteSelectSchema(ctx context.Context, db *sql.DB) (string, error) {
	const query = `
SELECT json_object(
  'databases', json_group_array(
    json_object(
      'name', d.[name],
      'tables', (
        SELECT json_group_array(
          json_object(
            'name', t.[name],
            'columns', (
              SELECT json_group_array(
                json_object(
                  'name', col.[name],
                  'type', col.[type],
                  'notnull', col.[notnull],
                  'default', col.[dflt_value],
                  'pk', col.[pk]
                )
              )
              FROM pragma_table_info(t.[name]) col
            )
          )
        )
        FROM sqlite_master t
        WHERE t.type = 'table'
          AND t.name NOT LIKE 'sqlite_%'
      )
    )
  )
) AS [schema]
FROM pragma_database_list d
		`

	var schema string
	row := db.QueryRowContext(ctx, query)
	return schema, row.Scan(&schema)
}
