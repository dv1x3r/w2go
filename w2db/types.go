package w2db

import (
	"context"
	"database/sql"
)

type QueryDB interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

type ExecDB interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type QueryExecDB interface {
	QueryDB
	ExecDB
}
