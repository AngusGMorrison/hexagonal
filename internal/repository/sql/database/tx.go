package database

import (
	"context"
	"fmt"

	hexsql "github.com/angusgmorrison/hexagonal/internal/repository/sql"
	"github.com/jmoiron/sqlx"
)

// Tx is a thin wrapper around *sqlx.Tx, allowing us to satisfy sql.Transaction.
type Tx struct {
	sqlxTx *sqlx.Tx
}

var _ hexsql.Transaction = (*Tx)(nil)

// Commit commits the transaction to the database.
func (tx *Tx) Commit() error {
	if err := tx.sqlxTx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// Rollback rolls back a transaction in progress. The underlying sql package
// guarantees that Rollback leaves the database in a clean state even when it
// returns an error.
func (tx *Tx) Rollback() error {
	if err := tx.sqlxTx.Rollback(); err != nil {
		return fmt.Errorf("roll back transaction: %w", err)
	}

	return nil
}

// Bind converts a query with named parameters and its struct or slice of struct
// argument into a query with the positional parameters of the database driver.
//
// Each element of a slice arg is treated as a separate row to be used in bulk
// operations.
func (tx *Tx) Bind(query string, arg any) (string, []any, error) {
	boundQuery, positionalArgs, err := tx.sqlxTx.BindNamed(query, arg)
	if err != nil {
		return "", nil, fmt.Errorf("db.Bind: %w", err)
	}

	return boundQuery, positionalArgs, nil
}

// Query executes the query and scans each row into dest, which must be a
// pointer to a slice.
func (tx *Tx) Query(ctx context.Context, dest any, query string, args ...any) error {
	if err := tx.sqlxTx.SelectContext(ctx, dest, query, args...); err != nil {
		return fmt.Errorf("tx.Query: %w", err)
	}

	return nil
}

// Execute executes a query that returns no result.
func (tx *Tx) Execute(ctx context.Context, query string, args ...any) error {
	_, err := tx.sqlxTx.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("tx.Execute: %w", err)
	}

	return nil
}
