package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/angusgmorrison/hexagonal/envconfig"
	"github.com/angusgmorrison/hexagonal/service"
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

// Select executes the query and scans each row into dest, which must be slice.
func (db *DB) Select(ctx context.Context, dest any, query string, args ...any) error {
	if err := db.sqlxDB.SelectContext(ctx, dest, query, args...); err != nil {
		return fmt.Errorf("db.Select: %w", err)
	}

	return nil
}

// Execute executes a query that returns no result.
func (db *DB) Execute(ctx context.Context, query string, args ...any) error {
	_, err := db.sqlxDB.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("db.Execute: %w", err)
	}

	return nil
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
func (db *DB) BeginSerializableTx(ctx context.Context) (*Tx, error) {
	sqlxTx, err := db.sqlxDB.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return nil, fmt.Errorf("beginTx: %w", err)
	}

	return &Tx{sqlxTx: sqlxTx}, nil
}

// BeginTx returns a new database transaction.
func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	sqlxTx, err := db.sqlxDB.BeginTxx(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("beginTx: %w", err)
	}

	return &Tx{sqlxTx: sqlxTx}, nil
}

type Tx struct {
	sqlxTx *sqlx.Tx
}

func (t *Tx) Commit() error {
	if err := t.sqlxTx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func (t *Tx) Rollback() error {
	if err := t.sqlxTx.Rollback(); err != nil {
		return fmt.Errorf("roll back transaction: %w", err)
	}

	return nil
}

// Select executes the query and scans each row into dest, which must be slice.
func (tx *Tx) Select(ctx context.Context, dest any, query string, args ...any) error {
	if err := tx.sqlxTx.SelectContext(ctx, dest, query, args...); err != nil {
		return fmt.Errorf("tx.Select: %w", err)
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

type TxTypeError struct {
	tx service.Transactor
}

func (t TxTypeError) Error() string {
	return fmt.Sprintf("service.Transactor with concrete type *sqlx.Tx required; got %T", t.tx)
}
