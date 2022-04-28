// Package sql describes a complete set of driver-agnostic interfaces required
// to work with relational databases.
package sql

import "context"

// Committer commits atomic operations to the database.
type Committer interface {
	Commit() error
}

// Rollbacker rolls back an atomic operation. Rollbackers must guarantee that
// the database is left in its clean, pre-transaction state even when Rollback
// returns an error.
type Rollbacker interface {
	Rollback() error
}

// Transactor represents the methods of a transaction concerned with atomicity.
type Transactor interface {
	Committer
	Rollbacker
}

// Execer executes a query that returns no result.
type Execer interface {
	Execute(ctx context.Context, query string, args ...any) error
}

// Queryer executes a query that populates dest with the returned rows. Since
// the number of rows is unknown, dest must be a pointer to a slice.
type Queryer interface {
	Query(ctx context.Context, dest any, query string, args ...any) error
}

// Binder converts a query with named parameters and its tagged struct or
// slice arguments to a query that uses the bind vars of the underlying database
// driver and a slice of positional arguments corresponding to those bindings.
//
// Bind should interpret each element of a slice argument as a separate row,
// e.g. as part of a bulk insertion.
type Binder interface {
	Bind(query string, arg any) (boundQuery string, positionalArgs []any, err error)
}

// BindQueryer can bind and execute a query with named parameters.
type BindQueryer interface {
	Binder
	Queryer
}

// Rebinder accepts a query with bind vars of one form (e.g. '?') and returns a
// query with bind vars appropriate to the underlying database driver. This is
// typically useful where queries are built dynamically, such as queries with IN
// clauses. The table can construct the query without knowing the bind var of
// the database.
type Rebinder interface {
	Rebind(query string) string
}

// RebindQueryer can rebind and execute a query.
type RebindQueryer interface {
	Rebinder
	Queryer
}

// Transaction provides all the methods required to query a database atomically.
type Transaction interface {
	Transactor
	Execer
	Queryer
	Binder
	Rebinder
}

// Beginner represents an object that can begin a transaction at some default
// isolation level.
type Beginner interface {
	Begin(ctx context.Context) (Transaction, error)
}

// Serializer represents an object that can begin a serializable transaction.
type Serializer interface {
	BeginSerializable(ctx context.Context) (Transaction, error)
}

// Database provides only the functionality of a database that application code
// needs to be concerned with.
type Database interface {
	Beginner
	Serializer
	Execer
	BindQueryer
}
