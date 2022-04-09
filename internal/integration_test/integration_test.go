//go:build integration

// Package integration_test contains integration tests for the hexagonal
// application.
package integration_test

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/angusgmorrison/hexagonal/internal/adapter/envconfig"
	"github.com/angusgmorrison/hexagonal/internal/adapter/repository/postgres"
	"github.com/angusgmorrison/hexagonal/internal/adapter/rest"
	server "github.com/angusgmorrison/hexagonal/internal/adapter/rest"
	"github.com/angusgmorrison/hexagonal/internal/controller"
)

const (
	_serverPort = 7532
	_dbPort     = 5432
	_dbName     = "hexagonal_test"
)

func TestMain(m *testing.M) {
	server, err := NewServer()
	if err != nil {
		panic(err)
	}

	go func() {
		if err := server.Run(); err != nil {
			panic(err)
		}
	}()

	defer func() {
		if err := server.GracefulShutdown(); err != nil {
			panic(err)
		}
	}()

	os.Exit(m.Run())
}

func NewServer() (*server.Server, error) {
	logger := log.New(os.Stdout, "hexagonal_test ", log.LstdFlags)

	envConfig := defaultEnvConfig()

	db, err := postgres.NewDB(envConfig.DB)
	if err != nil {
		return nil, fmt.Errorf("create database: %w", err)
	}

	transferRepo, err := postgres.NewTransferRepository(db, envConfig.App)
	if err != nil {
		return nil, fmt.Errorf("create transfer Repository: %w", err)
	}

	transferService := controller.NewTransferController(transferRepo)
	server := rest.NewServer(logger, envConfig, transferService)

	return server, nil
}

const (
	_insertBankAccountQuery = `
		INSERT INTO bank_accounts (organization_name, iban, bic, balance_cents)
		VALUES (:organization_name, :iban, :bic, :balance_cents)
		RETURNING *;`

	_getBankAccountByIDQuery = `
		SELECT id, organization_name, iban, bic, balance_cents
		FROM bank_accounts
		WHERE id = $1;
	`

	_countBankAccountsQuery = `SELECT COUNT(*) FROM bank_accounts;`

	_truncateBankAccountsQuery = `TRUNCATE TABLE bank_accounts CASCADE;`

	_countTransactionsQuery = `SELECT COUNT(*) FROM transactions;`

	_selectTransactionByCounterpartyNameQuery = `
		SELECT id, bank_account_id, counterparty_name, counterparty_iban,
			counterparty_bic, amount_cents, amount_currency, description
		FROM transactions
		WHERE counterparty_name = $1;`

	_truncateTransactionsQuery = `TRUNCATE TABLE transactions;`
)

type repository struct {
	db *postgres.DB
}

func (r *repository) insertBankAccount(
	br postgres.BankAccountRow,
) (postgres.BankAccountRow, error) {
	query, args, err := r.db.BindNamed(_insertBankAccountQuery, br)
	if err != nil {
		return br, fmt.Errorf("db.BindNamed: %w", err)
	}

	if err := r.db.Get(&br, query, args...); err != nil {
		return br, fmt.Errorf("insertBankAccount: %w", err)
	}

	return br, nil
}

func (r *repository) getBankAccountByID(id int64) (postgres.BankAccountRow, error) {
	var row postgres.BankAccountRow

	err := r.db.Get(&row, _getBankAccountByIDQuery, id)
	if err != nil {
		return row, fmt.Errorf("getBankAccount")
	}

	return row, nil
}

func (r *repository) countBankAccounts() (int64, error) {
	var count int64

	if err := r.db.Get(&count, _countBankAccountsQuery); err != nil {
		return 0, fmt.Errorf("countBankAccounts: %w", err)
	}

	return count, nil
}

func (r *repository) truncateBankAccounts() error {
	if _, err := r.db.Exec(_truncateBankAccountsQuery); err != nil {
		return fmt.Errorf("truncateBankAccounts: %w", err)
	}

	return nil
}

func (r *repository) countTransactions() (int64, error) {
	var count int64

	if err := r.db.Get(&count, _countTransactionsQuery); err != nil {
		return 0, fmt.Errorf("countTransactions: %w", err)
	}

	return count, nil
}

func (r *repository) selectTransactionsByCounterpartyName(
	name string,
) (postgres.TransactionRows, error) {
	var rows postgres.TransactionRows

	if err := r.db.Select(&rows, _selectTransactionByCounterpartyNameQuery, name); err != nil {
		return nil, fmt.Errorf("selectTransactionByCounterpartyName: %w", err)
	}

	return rows, nil
}

func (r *repository) truncateTransactions() error {
	if _, err := r.db.Exec(_truncateTransactionsQuery); err != nil {
		return fmt.Errorf("truncateTransactions: %w", err)
	}

	return nil
}

type infrastructure struct {
	repo   *repository
	client *http.Client
}

func newInfrastructure(env envconfig.EnvConfig) (*infrastructure, error) {
	db, err := postgres.NewDB(env.DB)
	if err != nil {
		return nil, fmt.Errorf("create database: %w", err)
	}

	infra := infrastructure{
		repo: &repository{
			db: db,
		},
		client: &http.Client{
			Timeout: env.HTTP.ClientTimeout,
		},
	}

	return &infra, nil
}

// mustCleanup is designed to be passed to (*testing.T).Cleanup to shut down the
// infrastructure. Panics if shutdown fails.
func (i *infrastructure) mustCleanup() {
	if err := i.repo.db.Close(); err != nil {
		panic(fmt.Errorf("db.Close: %w", err))
	}
}

func defaultEnvConfig() envconfig.EnvConfig {
	return envconfig.EnvConfig{
		App: envconfig.App{
			Name: "hexagonal",
			Env:  "test",
			Root: defaultAppRoot(),
		},
		HTTP: envconfig.HTTP{
			Host:                "",
			Port:                _serverPort,
			ClientTimeout:       5 * time.Second,
			ReadTimeout:         5 * time.Second,
			WriteTimeout:        5 * time.Second,
			ShutdownGracePeriod: 0,
		},
		DB: envconfig.DB{
			Host:            "postgres",
			Port:            _dbPort,
			Username:        "postgres",
			Password:        "postgres",
			Name:            _dbName,
			SSLMode:         "disable",
			ConnTimeout:     5 * time.Second,
			ConnMaxIdleTime: 0,
			ConnMaxLifetime: 0,
			MaxIdleConns:    20,
			MaxOpenConns:    20,
		},
	}
}

func defaultAppRoot() string {
	return filepath.Join(string(filepath.Separator), "usr", "src", "app")
}

func bulkTransferURL() string {
	return serverURL() + "/bulk_transfer"
}

func serverURL() string {
	return fmt.Sprintf("http://0.0.0.0:%d", _serverPort)
}

func defaultBankAccount() postgres.BankAccountRow {
	return postgres.BankAccountRow{
		OrganizationName: "ACME Corp",
		BIC:              "OIVUSCLQXXX",
		IBAN:             "FR10474608000002006107XXXXX",
		BalanceCents:     10000000,
	}
}
