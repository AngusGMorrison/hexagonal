package service

import "context"

// AtomicScribe represents a single-use, atomic connection to a repository.
// Implementors guarantee that all operations performed by the scribe before
// calling Commit are atomic.
type AtomicScribe interface {
	Begin(ctx context.Context) error
	Commit() error

	// Rollback must leave the data store in its original, clean state even in the
	// event of an error.
	//
	// After calling Commit, Rollback should return an error but otherwise have
	// no effect.
	Rollback() error
}
