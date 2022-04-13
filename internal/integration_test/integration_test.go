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
	logger := log.New(os.Stdout, "hexagonal_test ", log.LstdFlags)

	server, err := NewServer(logger)
	if err != nil {
		logger.Fatalf("Create server: %v\n", err)
	}

	go func() {
		if err := server.Run(); err != nil {
			logger.Fatalf("Run server: %v\n", err)
		}
	}()

	defer func() {
		if err := server.GracefulShutdown(); err != nil {
			logger.Fatalf("Shut down server: %v\n", err)
		}
	}()

	os.Exit(m.Run())
}

func NewServer(logger *log.Logger) (*server.Server, error) {
	envConfig := defaultEnvConfig()

	db, err := postgres.NewDB(envConfig.DB)
	if err != nil {
		return nil, fmt.Errorf("create database: %w", err)
	}

	transferQueryDir := filepath.Join(envConfig.App.Root, postgres.RelativeQueryDir(), "transfer")
	repo, err := postgres.NewTransferRepository(db, transferQueryDir)
	if err != nil {
		return nil, fmt.Errorf("create TransferRepository: %w", err)
	}

	transferService := controller.NewTransferController(logger, repo)
	server := rest.NewServer(logger, envConfig, transferService)

	return server, nil
}

const (
	_insertBankAccount                    postgres.QueryFilename = "insert_bank_account.sql"
	_getBankAccountByID                   postgres.QueryFilename = "get_bank_account_by_id.sql"
	_countBankAccounts                    postgres.QueryFilename = "count_bank_accounts.sql"
	_truncateBankAccounts                 postgres.QueryFilename = "truncate_bank_accounts.sql"
	_countTransactions                    postgres.QueryFilename = "count_transactions.sql"
	_selectTransactionsByCounterpartyName postgres.QueryFilename = "select_transactions_by_counterparty_name.sql"
	_truncateTransactions                 postgres.QueryFilename = "truncate_transactions.sql"
)

type testRepository struct {
	db       *postgres.DB
	queryDir string
	queries  postgres.Queries
}

func newTestRepository(db *postgres.DB, queryDir string) (*testRepository, error) {
	repo := testRepository{
		db:       db,
		queryDir: queryDir,
	}

	if err := repo.loadQueries(); err != nil {
		return nil, err
	}

	return &repo, nil
}

func (tr *testRepository) loadQueries() error {
	queryFilenames := []postgres.QueryFilename{
		_insertBankAccount,
		_getBankAccountByID,
		_countBankAccounts,
		_truncateBankAccounts,
		_countTransactions,
		_selectTransactionsByCounterpartyName,
		_truncateTransactions,
	}

	if tr.queries == nil {
		tr.queries = make(postgres.Queries, len(queryFilenames))
	}

	if err := tr.queries.Load(tr.queryDir, queryFilenames); err != nil {
		return fmt.Errorf("load test queries: %w", err)
	}

	return nil
}

func (tr *testRepository) insertBankAccount(
	br postgres.BankAccountRow,
) (postgres.BankAccountRow, error) {
	query, args, err := tr.db.BindNamed(tr.queries[_insertBankAccount], br)
	if err != nil {
		return br, fmt.Errorf("db.BindNamed: %w", err)
	}

	if err := tr.db.Get(&br, query, args...); err != nil {
		return br, fmt.Errorf("insertBankAccount: %w", err)
	}

	return br, nil
}

func (tr *testRepository) getBankAccountByID(id int64) (postgres.BankAccountRow, error) {
	var row postgres.BankAccountRow

	err := tr.db.Get(&row, tr.queries[_getBankAccountByID], id)
	if err != nil {
		return row, fmt.Errorf("getBankAccount")
	}

	return row, nil
}

func (tr *testRepository) countBankAccounts() (int64, error) {
	var count int64

	if err := tr.db.Get(&count, tr.queries[_countBankAccounts]); err != nil {
		return 0, fmt.Errorf("countBankAccounts: %w", err)
	}

	return count, nil
}

func (tr *testRepository) truncateBankAccounts() error {
	if _, err := tr.db.Exec(tr.queries[_truncateBankAccounts]); err != nil {
		return fmt.Errorf("truncateBankAccounts: %w", err)
	}

	return nil
}

func (tr *testRepository) countTransactions() (int64, error) {
	var count int64

	if err := tr.db.Get(&count, tr.queries[_countTransactions]); err != nil {
		return 0, fmt.Errorf("countTransactions: %w", err)
	}

	return count, nil
}

func (tr *testRepository) selectTransactionsByCounterpartyName(
	name string,
) (postgres.TransactionRows, error) {
	var rows postgres.TransactionRows

	if err := tr.db.Select(&rows, tr.queries[_selectTransactionsByCounterpartyName], name); err != nil {
		return nil, fmt.Errorf("selectTransactionByCounterpartyName: %w", err)
	}

	return rows, nil
}

func (tr *testRepository) truncateTransactions() error {
	if _, err := tr.db.Exec(tr.queries[_truncateTransactions]); err != nil {
		return fmt.Errorf("truncateTransactions: %w", err)
	}

	return nil
}

type infrastructure struct {
	logger *log.Logger
	repo   *testRepository
	client *http.Client
}

func newInfrastructure(envConfig envconfig.EnvConfig, logger *log.Logger) (*infrastructure, error) {
	db, err := postgres.NewDB(envConfig.DB)
	if err != nil {
		return nil, fmt.Errorf("create database: %w", err)
	}

	testQueryDir := filepath.Join(envConfig.App.Root, postgres.RelativeQueryDir(), "transfer")
	repo, err := newTestRepository(db, testQueryDir)
	if err != nil {
		return nil, fmt.Errorf("create testRepository: %w", err)
	}

	infra := infrastructure{
		logger: logger,
		repo:   repo,
		client: &http.Client{
			Timeout: envConfig.HTTP.ClientTimeout,
		},
	}

	return &infra, nil
}

// cleanup is designed to be passed to (*testing.T).Cleanup to shut down the
// infrastructure. Panics if shutdown fails.
func (i *infrastructure) cleanup() {
	if err := i.repo.db.Close(); err != nil {
		i.logger.Printf("Close database: %v", err)
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
