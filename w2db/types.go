package w2db

import (
	"context"
	"database/sql"
)

// QueryExecer is the database interface required by this package.
//
// It is satisfied by *sql.DB and *sql.Tx.
type QueryExecer interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// Providable marks a value that knows whether it was sent by the client.
//
// Update skips Providable values when IsProvided returns false. w2.Field implements this interface.
type Providable interface {
	IsProvided() bool
}
