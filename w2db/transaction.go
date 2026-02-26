package w2db

import (
	"context"
	"database/sql"
)

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
