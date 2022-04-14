package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/angusgmorrison/hexagonal/internal/adapter/envconfig"
	"github.com/angusgmorrison/hexagonal/internal/service"
	"github.com/jmoiron/sqlx"

	// Load postgres driver
	_ "github.com/lib/pq"
)

// DB is a thin wrapper around an *sqlx.DB, allowing us to write our own methods
// on the database struct, expose only the methods required by our repositories
// and enforce conventions, such as always requiring a context to be passed when
// querying the DB.
type DB struct {
	config envconfig.DB
	sqlxDB *sqlx.DB
}

// NewDB returns a configured Postgres database that is ready to use, or an
// error if the connection can't be established.
func NewDB(dbConfig envconfig.DB) (*DB, error) {
	sqlxDB, err := sqlx.Open("postgres", dbConfig.URL())
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	db := DB{
		config: dbConfig,
		sqlxDB: sqlxDB,
	}

	db.configureConns()

	if err := db.ping(); err != nil {
		return nil, err
	}

	return &db, nil
}

// BindNamed binds a named query, replacing its named arguments with Postgres
// positional bindvars and producing a slice of args in the correct binding
// order.
func (db *DB) BindNamed(query string, arg any) (string, []any, error) {
	return db.sqlxDB.BindNamed(query, arg)
}

// Get a single row, scanning the result into dest. Placeholder parameters are
// replaced with supplied args.
func (db *DB) Get(ctx context.Context, dest any, query string, args ...any) error {
	return db.sqlxDB.GetContext(ctx, dest, query, args...)
}

// Select executes the query and scans each row into dest, which must be slice.
func (db *DB) Select(ctx context.Context, dest any, query string, args ...any) error {
	return db.sqlxDB.SelectContext(ctx, dest, query, args...)
}

// Exec executes the query and returns the result.
func (db *DB) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return db.sqlxDB.ExecContext(ctx, query, args...)
}

// NamedExec executes a query, replaced named arguments with fields from arg.
func (db *DB) NamedExec(ctx context.Context, query string, arg any) (sql.Result, error) {
	return db.sqlxDB.NamedExecContext(ctx, query, arg)
}

// LoadFile loads an entire SQL file into memory and executes it.
func (db *DB) LoadFile(path string) (*sql.Result, error) {
	return sqlx.LoadFile(db.sqlxDB, path)
}

// Close closes the underlying database connection.
func (db *DB) Close() error {
	if err := db.sqlxDB.Close(); err != nil {
		return fmt.Errorf("close inner database: %w", err)
	}

	return nil
}

// BeginTx returns a new database transaction.
func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sqlx.Tx, error) {
	tx, err := db.sqlxDB.BeginTxx(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("beginTx: %w", err)
	}

	return tx, nil
}

func (db *DB) configureConns() {
	db.sqlxDB.SetConnMaxIdleTime(db.config.ConnMaxIdleTime)
	db.sqlxDB.SetConnMaxLifetime(db.config.ConnMaxLifetime)
	db.sqlxDB.SetMaxIdleConns(db.config.MaxIdleConns)
	db.sqlxDB.SetMaxOpenConns(db.config.MaxOpenConns)
}

func (db *DB) ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), db.config.ConnTimeout)
	defer cancel()

	if err := db.sqlxDB.PingContext(ctx); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}

	return nil
}

func truncationPermitted(env string) bool {
	return env == "development" || env == "test"
}

// UnpermittedTruncationError is used to signal a truncation attempt in an
// environment which does not support it.
type UnpermittedTruncationError struct {
	env string
}

func (u UnpermittedTruncationError) Error() string {
	return fmt.Sprintf("truncation not permitted in environment %q", u.env)
}

// TxTypeError represents a failed conversion from a service.Transactor
// interface to an *sqlx.Tx.
type TxTypeError struct {
	tx service.Transactor
}

func (t TxTypeError) Error() string {
	return fmt.Sprintf("service.Transactor with concrete type *sqlx.Tx required; got %T", t.tx)
}
