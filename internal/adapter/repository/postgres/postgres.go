package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/angusgmorrison/hexagonal/internal/adapter/envconfig"
	"github.com/jmoiron/sqlx"

	// Load postgres driver
	_ "github.com/lib/pq"
)

// DB wraps a config object and a *sqlx.DB, allowing us to write our own methods
// on the database struct.
type DB struct {
	config envconfig.DB
	sqlxDB *sqlx.DB
}

// NewDB returns a configured Postgres database that is ready to use, or an
// error if the connection can't be established.
func NewDB(cfg envconfig.DB) (*DB, error) {
	sqlxDB, err := sqlx.Open("postgres", cfg.URL())
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	db := DB{
		config: cfg,
		sqlxDB: sqlxDB,
	}

	db.sqlxDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
	db.sqlxDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.sqlxDB.SetMaxIdleConns(cfg.MaxIdleConns)
	db.sqlxDB.SetMaxOpenConns(cfg.MaxOpenConns)

	if err := db.ping(); err != nil {
		return nil, err
	}

	return &db, nil
}

// BeginTxx starts and returns a new sqlx transaction.
func (db *DB) BeginTxx(ctx context.Context, opts *sql.TxOptions) (*sqlx.Tx, error) {
	tx, err := db.sqlxDB.BeginTxx(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("BeginTxx: %w", err)
	}

	return tx, nil
}

// BindNamed binds a named query, replacing its named arguments with Postgres
// positional bindvars and producing a slice of args in the correct binding
// order.
func (db *DB) BindNamed(query string, arg any) (string, []any, error) {
	return db.sqlxDB.BindNamed(query, arg)
}

// Get a single row, scanning the result into dest. Placeholder parameters are
// replaced with supplied args.
func (db *DB) Get(dest any, query string, args ...any) error {
	return db.sqlxDB.Get(dest, query, args...)
}

// Select executes the query and scans each row into dest, which must be slice.
func (db *DB) Select(dest any, query string, args ...any) error {
	return db.sqlxDB.Select(dest, query, args...)
}

// Exec executes the query and returns the result.
func (db *DB) Exec(query string, args ...any) (sql.Result, error) {
	return db.sqlxDB.Exec(query, args...)
}

// NamedExec executes a query, replaced named arguments with fields from arg.
func (db *DB) NamedExec(query string, arg any) (sql.Result, error) {
	return db.sqlxDB.NamedExec(query, arg)
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

func (db *DB) ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), db.config.ConnTimeout)
	defer cancel()

	if err := db.sqlxDB.PingContext(ctx); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}

	return nil
}
