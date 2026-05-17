package w2db

import (
	"context"
	"database/sql"
)

// WithinTransaction runs fn inside a database transaction.
//
// The transaction is committed when fn returns nil. If fn returns an error, or
// commit fails, the deferred rollback leaves the transaction closed.
func WithinTransaction(db *sql.DB, fn func(tx *sql.Tx) error) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit()
}

// WithinTransactionContext runs fn inside a database transaction created with ctx.
//
// The same context is passed to fn so nested calls can use it for query cancellation and logging.
func WithinTransactionContext(ctx context.Context, db *sql.DB, fn func(ctx context.Context, tx *sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := fn(ctx, tx); err != nil {
		return err
	}

	return tx.Commit()
}
