package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/angusgmorrison/hexagonal/envconfig"
	hexsql "github.com/angusgmorrison/hexagonal/repository/sql"
	"github.com/jmoiron/sqlx"

	// Load postgres driver
	_ "github.com/lib/pq"
)

// DB is a thin wrapper around an *sqlx.DB, allowing us to write our own methods
// on the database struct and implement the interfaces of package sql.
type DB struct {
	config envconfig.DB
	sqlxDB *sqlx.DB
}

var _ hexsql.Database = (*DB)(nil)

// New returns a configured Postgres database that is ready to use, or an
// error if the connection can't be established.
func New(dbConfig envconfig.DB) (*DB, error) {
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

// Execute executes a query that returns no result.
func (db *DB) Execute(ctx context.Context, query string, args ...any) error {
	_, err := db.sqlxDB.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("db.Execute: %w", err)
	}

	return nil
}

// Query executes the query and scans each row into dest, which must be a
// pointer to a slice.
func (db *DB) Query(ctx context.Context, dest any, query string, args ...any) error {
	if err := db.sqlxDB.SelectContext(ctx, dest, query, args...); err != nil {
		return fmt.Errorf("db.Select: %w", err)
	}

	return nil
}

// Bind converts a query with named parameters and its struct or slice of struct
// argument into a query with the positional parameters of the database driver.
//
// Each element of a slice arg is treated as a separate row to be used in bulk
// operations.
func (db *DB) Bind(query string, arg any) (string, []any, error) {
	boundQuery, positionalArgs, err := db.sqlxDB.BindNamed(query, arg)
	if err != nil {
		return "", nil, fmt.Errorf("db.Bind: %w", err)
	}

	return boundQuery, positionalArgs, nil
}

// BeginSerializable returns a new, serializable database transaction.
func (db *DB) BeginSerializable(ctx context.Context) (hexsql.Transaction, error) {
	return db.begin(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
}

// Begin returns a new database transaction at the database's current isolation
// level.
func (db *DB) Begin(ctx context.Context) (hexsql.Transaction, error) {
	return db.begin(ctx, nil)
}

func (db *DB) begin(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	sqlxTx, err := db.sqlxDB.BeginTxx(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}

	return &Tx{sqlxTx: sqlxTx}, nil
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
