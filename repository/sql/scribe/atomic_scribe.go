package scribe

import (
	"context"
	"errors"
	"fmt"

	"github.com/angusgmorrison/hexagonal/repository/sql"
)

var (
	// ErrTransactionInProgress is returned when more than one concurrent
	// transaction is attempted.
	ErrTransactionInProgress = errors.New("transaction already in progress")

	// ErrTransactionNotStarted is returned when an atomic operation is
	// requested but the scribe hasn't begun a transaction.
	ErrTransactionNotStarted = errors.New("transaction not started")

	// ErrDone is returned when an operation is requested on a scribe that has
	// already completed its transaction.
	ErrDone = errors.New("scribe already used")
)

// atomicScribe is a lightweight, disposable scribe that guarantees that its
// operations are atomic. Begin starts a transaction on the repository. Commit
// and Rollback end transactions.
//
// atomicScribe is single-use by design. After the first transaction, the scribe
// will return errors to all attempts to reuse it. So long as the scribe's
// transaction takes place in a single goroutine, this eliminates the need for
// thread safety protections.
//
// The recommended use pattern is to create a factory function that returns a
// new scribe for each database operation.
type atomicScribe struct {
	db   sql.Database
	tx   sql.Transaction
	done bool
}

// Begin starts a new transaction on the scribe.
func (as *atomicScribe) Begin(ctx context.Context) error {
	if as.done {
		return ErrDone
	}

	if as.tx != nil {
		return ErrTransactionInProgress
	}

	tx, err := as.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("(*AtomicScribe).Begin: %w", err)
	}

	as.tx = tx

	return nil
}

// BeginSerializable starts a new serializable transaction on the scribe.
func (as *atomicScribe) BeginSerializable(ctx context.Context) error {
	if as.done {
		return ErrDone
	}

	if as.tx != nil {
		return ErrTransactionInProgress
	}

	tx, err := as.db.BeginSerializable(ctx)
	if err != nil {
		return fmt.Errorf("(*AtomicScribe).BeginSerializable: %w", err)
	}

	as.tx = tx

	return nil
}

// Commit commits the current transaction.
func (as *atomicScribe) Commit() error {
	if as.done {
		return ErrDone
	}

	if as.tx == nil {
		return ErrTransactionNotStarted
	}

	defer func() { as.done = true }()

	if err := as.tx.Commit(); err != nil {
		return fmt.Errorf("(*AtomicScribe).Commit: %w", err)
	}

	return nil
}

// Rollback rolls back the current transaction. If Rollback returns an error,
// the database will be left in its clean, pre-transaction state.
func (as *atomicScribe) Rollback() error {
	if as.done {
		return ErrDone
	}

	if as.tx == nil {
		return ErrTransactionNotStarted
	}

	defer func() { as.done = true }()

	if err := as.tx.Rollback(); err != nil {
		return fmt.Errorf("(*AtomicScribe).Rollback: %w", err)
	}

	return nil
}
