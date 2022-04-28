package scribe

import (
	"context"
	"errors"
	"fmt"

	"github.com/angusgmorrison/hexagonal/internal/repository/sql"
)

var (
	// ErrTransactionInProgress is returned when more than one concurrent
	// transaction is attempted.
	ErrTransactionInProgress = errors.New("transaction already in progress")

	// ErrTransactionNotStarted is returned when an atomic operation is
	// requested but the scribe hasn't begun a transaction.
	ErrTransactionNotStarted = errors.New("transaction not started")
)

// atomicScribe is a lightweight, disposable scribe that guarantees that its
// operations are atomic. Begin starts a transaction on the repository. Commit
// and Rollback end transactions.
//
// By design, atomicScribe is not thread-safe. Each instance can handle at most
// one concurrent transaction. The intended use pattern is to create a new
// scribe for each database operation, typically via a factory function.
type atomicScribe struct {
	db sql.Database
	tx sql.Transaction
}

// Begin starts a new transaction on the scribe.
func (as *atomicScribe) Begin(ctx context.Context) error {
	if as.tx != nil {
		return ErrTransactionInProgress
	}

	tx, err := as.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("(*atomicScribe).Begin: %w", err)
	}

	as.tx = tx

	return nil
}

// BeginSerializable starts a new serializable transaction on the scribe.
func (as *atomicScribe) BeginSerializable(ctx context.Context) error {
	if as.tx != nil {
		return ErrTransactionInProgress
	}

	tx, err := as.db.BeginSerializable(ctx)
	if err != nil {
		return fmt.Errorf("(*atomicScribe).BeginSerializable: %w", err)
	}

	as.tx = tx

	return nil
}

// Commit commits the current transaction.
func (as *atomicScribe) Commit() error {
	if as.tx == nil {
		return ErrTransactionNotStarted
	}

	defer func() { as.tx = nil }()

	if err := as.tx.Commit(); err != nil {
		return fmt.Errorf("(*atomicScribe).Commit: %w", err)
	}

	return nil
}

// Rollback rolls back the current transaction. If Rollback returns an error,
// the database will be left in its clean, pre-transaction state.
func (as *atomicScribe) Rollback() error {
	if as.tx == nil {
		return ErrTransactionNotStarted
	}

	defer func() { as.tx = nil }()

	if err := as.tx.Rollback(); err != nil {
		return fmt.Errorf("(*atomicScribe).Rollback: %w", err)
	}

	return nil
}

func (as *atomicScribe) getTx() (sql.Transaction, error) {
	if as.tx == nil {
		return nil, ErrTransactionNotStarted
	}

	return as.tx, nil
}
