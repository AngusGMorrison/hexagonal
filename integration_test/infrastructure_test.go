//go:build integration

package integration_test

import (
	"fmt"
	"log"
	"net/http"

	"github.com/angusgmorrison/hexagonal/internal/envconfig"
	"github.com/angusgmorrison/hexagonal/internal/repository/sql/database"
)

type infrastructure struct {
	logger *log.Logger
	db     *database.DB
	client *http.Client
}

func newInfrastructure(envConfig envconfig.EnvConfig, logger *log.Logger) (*infrastructure, error) {
	db, err := database.New(envConfig.DB)
	if err != nil {
		return nil, fmt.Errorf("create database: %w", err)
	}

	infra := infrastructure{
		logger: logger,
		db:     db,
		client: &http.Client{
			Timeout: envConfig.HTTP.ClientTimeout,
		},
	}

	return &infra, nil
}

// cleanup is designed to be passed to (*testing.T).Cleanup to shut down the
// infrastructure.
func (i *infrastructure) cleanup() {
	if err := i.db.Close(); err != nil {
		i.logger.Printf("Close database: %v", err)
	}
}
