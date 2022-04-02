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

// NewDB returns a configured database that is ready to use, or an error if the
// connection can't be established.
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

func (pg *Postgres) BeginTxx(ctx context.Context, opts *sql.TxOptions) (*sqlx.Tx, error) {
	tx, err := pg.db.BeginTxx(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("BeginTxx: %w", err)
	}

	return tx, nil
}

func (pg *Postgres) Close() error {
	if err := pg.db.Close(); err != nil {
		return fmt.Errorf("close inner database: %w", err)
	}

	return nil
}

func (p *Postgres) ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), p.config.ConnTimeout)
	defer cancel()

	if err := p.db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}

	return nil
}
