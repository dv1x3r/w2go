package w2db

import (
	"context"
	"database/sql"
)

type QueryExecer interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type Providable interface {
	IsProvided() bool
}
