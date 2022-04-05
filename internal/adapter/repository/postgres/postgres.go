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

// Postgres wraps a config object and a *sqlx.DB, allowing us to write our own methods
// on the database struct.
type Postgres struct {
	config envconfig.DB
	db     *sqlx.DB
}

// New returns a configured Postgres database that is ready to use, or an error
// if the connection can't be established.
func New(cfg envconfig.DB) (*Postgres, error) {
	db, err := sqlx.Open("postgres", cfg.URL())
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	pg := Postgres{
		config: cfg,
		db:     db,
	}

	pg.db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
	pg.db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	pg.db.SetMaxIdleConns(cfg.MaxIdleConns)
	pg.db.SetMaxOpenConns(cfg.MaxOpenConns)

	if err := pg.ping(); err != nil {
		return nil, err
	}

	return &pg, nil
}

// BeginTxx starts and returns a new sqlx transaction.
func (pg *Postgres) BeginTxx(ctx context.Context, opts *sql.TxOptions) (*sqlx.Tx, error) {
	tx, err := pg.db.BeginTxx(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("BeginTxx: %w", err)
	}

	return tx, nil
}

// BindNamed binds a named query, replacing its named arguments with Postgres
// positional bindvars and producing a slice of args in the correct binding
// order.
func (pg *Postgres) BindNamed(query string, arg any) (string, []any, error) {
	return pg.db.BindNamed(query, arg)
}

// Get a single row, scanning the result into dest. Placeholder parameters are
// replaced with supplied args.
func (pg *Postgres) Get(dest any, query string, args ...any) error {
	return pg.db.Get(dest, query, args...)
}

// Select executes the query and scans each row into dest, which must be slice.
func (pg *Postgres) Select(dest any, query string, args ...any) error {
	return pg.db.Select(dest, query, args...)
}

// Exec executes the query and returns the result.
func (pg *Postgres) Exec(query string, args ...any) (sql.Result, error) {
	return pg.db.Exec(query, args...)
}

// NamedExec executes a query, replaced named arguments with fields from arg.
func (pg *Postgres) NamedExec(query string, arg any) (sql.Result, error) {
	return pg.db.NamedExec(query, arg)
}

// LoadFile loads an entire SQL file into memory and executes it.
func (pg *Postgres) LoadFile(path string) (*sql.Result, error) {
	return sqlx.LoadFile(pg.db, path)
}

// Close closes the underlying database connection.
func (pg *Postgres) Close() error {
	if err := pg.db.Close(); err != nil {
		return fmt.Errorf("close inner database: %w", err)
	}

	return nil
}

func (pg *Postgres) ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), pg.config.ConnTimeout)
	defer cancel()

	if err := pg.db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}

	return nil
}
