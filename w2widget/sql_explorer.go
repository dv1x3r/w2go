package w2widget

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/dv1x3r/w2go/w2"
)

type SQLExplorerExecRequest struct {
	Query string `json:"query"`
}

type SQLExplorerExecResult struct {
	Status  w2.Status        `json:"status"`
	Columns []string         `json:"columns"`
	Records []map[string]any `json:"records"`
	Total   int              `json:"total"`
}

func NewSQLExplorerExecResult(columns []string, records []map[string]any, total int) SQLExplorerExecResult {
	return SQLExplorerExecResult{
		Status:  w2.StatusSuccess,
		Columns: columns,
		Records: records,
		Total:   total,
	}
}

func (res SQLExplorerExecResult) Write(w http.ResponseWriter) error {
	data, err := json.Marshal(res)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	return err
}

func SQLExplorerExecHandler(db *sql.DB) func(w http.ResponseWriter, r *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		var req SQLExplorerExecRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return err
		}

		res, err := SQLExplorerExecQuery(r.Context(), db, req.Query)
		if err != nil {
			return err
		}

		return res.Write(w)
	}
}

func SQLExplorerExecHTTPHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req SQLExplorerExecRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			res := w2.NewErrorResponse(err.Error())
			res.Write(w, http.StatusBadRequest)
			return
		}

		res, err := SQLExplorerExecQuery(r.Context(), db, req.Query)
		if err != nil {
			res := w2.NewErrorResponse(err.Error())
			res.Write(w, http.StatusInternalServerError)
			return
		}

		res.Write(w)
	}
}

func SQLExplorerExecQuery(ctx context.Context, db *sql.DB, query string) (SQLExplorerExecResult, error) {
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return SQLExplorerExecResult{}, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return SQLExplorerExecResult{}, err
	}

	records := []map[string]any{}

	for rows.Next() {
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return SQLExplorerExecResult{}, err
		}

		record := map[string]any{}
		for i, column := range columns {
			record[column] = values[i]
		}

		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return SQLExplorerExecResult{}, err
	}

	return NewSQLExplorerExecResult(columns, records, len(records)), nil
}

func SQLExplorerSchemaHandler(db *sql.DB) func(w http.ResponseWriter, r *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {

		return nil
	}
}

func SQLExplorerSchemaHTTPHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}
