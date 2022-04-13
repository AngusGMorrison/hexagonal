//go:build integration

package integration_test

import (
	"fmt"
	"log"
	"net/http"

	"github.com/angusgmorrison/hexagonal/internal/adapter/envconfig"
	"github.com/angusgmorrison/hexagonal/internal/adapter/repository/postgres"
)

type infrastructure struct {
	logger          *log.Logger
	db              *postgres.DB
	bankAccountRepo *postgres.BankAccountRepository
	transactionRepo *postgres.TransactionRepository
	client          *http.Client
}

func newInfrastructure(envConfig envconfig.EnvConfig, logger *log.Logger) (*infrastructure, error) {
	db, err := postgres.NewDB(envConfig.DB)
	if err != nil {
		return nil, fmt.Errorf("create database: %w", err)
	}

	bankAccountRepo, err := postgres.NewBankAccountRepository(db, envConfig.App)
	if err != nil {
		return nil, fmt.Errorf("create BankAccountRepository: %w", err)
	}

	transactionRepo, err := postgres.NewTransactionRepository(db, envConfig.App)
	if err != nil {
		return nil, fmt.Errorf("create TransactionRepository: %w", err)
	}

	infra := infrastructure{
		logger:          logger,
		db:              db,
		bankAccountRepo: bankAccountRepo,
		transactionRepo: transactionRepo,
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
