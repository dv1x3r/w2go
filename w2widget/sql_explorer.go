package w2widget

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/dv1x3r/w2go/w2"
)

// SQLExecField is the cell type returned by the SQL explorer result grid.
type SQLExecField = w2.Field[string]

// SQLExecRow is one SQL explorer result row keyed by column name.
type SQLExecRow map[string]SQLExecField

// SQLExecResult is the JSON response returned by SQL explorer query execution.
type SQLExecResult struct {
	// Status is set to w2.StatusSuccess by NewSQLExecResult.
	Status w2.Status `json:"status"`

	// Columns contains the result column names in display order.
	Columns []string `json:"columns"`

	// Records contains the scanned query rows.
	Records []SQLExecRow `json:"records"`

	// Total is the number of result rows.
	Total int `json:"total"`
}

// SQLExecRequest is the JSON request body for SQL explorer query execution.
type SQLExecRequest struct {
	// Query is the SQL text to execute.
	Query string `json:"query"`
}

// NewSQLExecResult returns a successful SQL explorer result.
func NewSQLExecResult(columns []string, records []SQLExecRow, total int) SQLExecResult {
	return SQLExecResult{
		Status:  w2.StatusSuccess,
		Columns: columns,
		Records: records,
		Total:   total,
	}
}

// Write sends the SQL explorer result as application/json.
func (res SQLExecResult) Write(w http.ResponseWriter) error {
	data, err := json.Marshal(res)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	return err
}

// SQLExecHandler returns a query-execution handler that reports errors to the
// caller instead of writing error responses itself.
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

// SQLExecHTTPHandler returns an http.HandlerFunc for SQL explorer query execution.
//
// The handler writes JSON error responses for malformed requests and query failures.
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

// SQLExecQuery executes query and scans the returned rows for the SQL explorer.
//
// Empty queries return an error. The query text is executed as-is, so callers
// must restrict access to this function when query text comes from a user.
func SQLExecQuery(ctx context.Context, db *sql.DB, query string) (SQLExecResult, error) {
	if strings.TrimSpace(query) == "" {
		return SQLExecResult{}, errors.New("query is empty")
	}

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return SQLExecResult{}, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return SQLExecResult{}, err
	}

	if columns == nil {
		columns = []string{}
	}

	records := []SQLExecRow{}

	for rows.Next() {
		values := make([]SQLExecField, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return SQLExecResult{}, err
		}

		record := SQLExecRow{}
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

// SQLiteSchemaHandler returns a SQLite schema handler that reports errors to the caller instead of writing error responses itself.
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

// SQLiteSchemaHTTPHandler returns an http.HandlerFunc that writes the SQLite schema JSON used by the SQL explorer sidebar.
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

// SQLiteSelectSchema returns SQLite database, table, and column metadata as a JSON document for the SQL explorer sidebar.
func SQLiteSelectSchema(ctx context.Context, db *sql.DB) (string, error) {
	const query = `
SELECT json_object(
  'databases', (
    SELECT json_group_array(
      json_object(
        'name', db.[name],
        'tables', (
          SELECT json_group_array(
            json_object(
              'name', t.[name],
              'type', t.[type],
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
          FROM (
            SELECT [name], [type]
            FROM pragma_table_list
            WHERE [schema] = db.[name]
              AND [name] NOT LIKE 'sqlite_%'
            ORDER BY [name]
          ) t
        )
      )
    )
    FROM pragma_database_list db
  )
) AS [schema]
		`

	var schema string
	row := db.QueryRowContext(ctx, query)
	return schema, row.Scan(&schema)
}
