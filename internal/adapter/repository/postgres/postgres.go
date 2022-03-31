package postgres

import (
	"context"
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
	*sqlx.DB
}

// NewDB returns a configured database that is ready to use, or an error if the
// connection can't be established.
func New(cfg envconfig.DB) (*Postgres, error) {
	db, err := sqlx.Open("postgres", cfg.URL())
	if err != nil {
		return nil, fmt.Errorf("failed to open DB: %w", err)
	}

	pg := Postgres{
		config: cfg,
		DB:     db,
	}

	pg.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
	pg.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	pg.SetMaxIdleConns(cfg.MaxIdleConns)
	pg.SetMaxOpenConns(cfg.MaxOpenConns)

	if err := pg.ping(); err != nil {
		return nil, err
	}

	return &pg, nil
}

func (p *Postgres) ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), p.config.ConnTimeout)
	defer cancel()

	if err := p.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	return nil
}
