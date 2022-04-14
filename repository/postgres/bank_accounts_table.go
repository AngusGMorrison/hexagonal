package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"

	"github.com/angusgmorrison/hexagonal/envconfig"
	"github.com/angusgmorrison/hexagonal/service"
	"github.com/jmoiron/sqlx"
)

// BankAccount represents a row of the bank_accounts table.
type BankAccount struct {
	ID               int64  `db:"id"`
	OrganizationName string `db:"organization_name"`
	BalanceCents     int64  `db:"balance_cents"`
	IBAN             string `db:"iban"`
	BIC              string `db:"bic"`
}

func bankAccountFromDomain(cba service.BankAccount) BankAccount {
	return BankAccount{
		ID:               cba.ID,
		OrganizationName: cba.OrganizationName,
		IBAN:             cba.OrganizationIBAN,
		BIC:              cba.OrganizationBIC,
		BalanceCents:     cba.BalanceCents,
	}
}

func (ba BankAccount) toDomain() service.BankAccount {
	return service.BankAccount{
		ID:               ba.ID,
		OrganizationName: ba.OrganizationName,
		OrganizationIBAN: ba.IBAN,
		OrganizationBIC:  ba.BIC,
		BalanceCents:     ba.BalanceCents,
	}
}

// BankAccountRepository operates on the bank_accounts table.
type BankAccountRepository struct {
	db        *DB
	appConfig envconfig.App
	queries   Queries
}

// Statically verify that Repository satisfies service.Repository.
var _ service.AtomicBankAccountRepository = (*BankAccountRepository)(nil)

const (
	_countBankAccounts     QueryFilename = "count_bank_accounts.sql"
	_findBankAccountByIBAN QueryFilename = "find_bank_account_by_iban.sql"
	_findBankAccountByID   QueryFilename = "find_bank_account_by_id.sql"
	_insertBankAccounts    QueryFilename = "insert_bank_accounts.sql"
	_truncateBankAccounts  QueryFilename = "truncate_bank_accounts.sql"
	_updateBankAccount     QueryFilename = "update_bank_account.sql"
)

// NewBankAccountRepository returns a new BankAccountRepository with its queries
// preloaded.
func NewBankAccountRepository(db *DB, appConfig envconfig.App) (*BankAccountRepository, error) {
	repo := BankAccountRepository{
		db:        db,
		appConfig: appConfig,
	}

	if err := repo.loadQueries(); err != nil {
		return nil, err
	}

	return &repo, nil
}

// BeginSerializableTx starts a new sqlx transaction with isolation level
// Serializable and returns it as a service.Transactor for use in atomic
// repository operations.
func (bar *BankAccountRepository) BeginSerializableTx(
	ctx context.Context,
) (service.Transactor, error) {
	tx, err := bar.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return nil, fmt.Errorf("BeginSerializableTx: %w", err)
	}

	return tx, nil
}

// FindByIBANTx retrieves the bank account with the given IBAN using
// transactor. If there is no matching bank account, an error is returned.
func (bar *BankAccountRepository) FindByIBANTx(
	ctx context.Context,
	transactor service.Transactor,
	iban string,
) (service.BankAccount, error) {
	tx, ok := transactor.(*sqlx.Tx)
	if !ok {
		return service.BankAccount{}, TxTypeError{tx: tx}
	}

	var row BankAccount

	if err := tx.GetContext(
		ctx,
		&row,
		bar.queries[_findBankAccountByIBAN],
		iban,
	); err != nil {
		return service.BankAccount{}, fmt.Errorf("get bank account with IBAN %q: %w",
			iban, err)
	}

	return row.toDomain(), nil
}

// UpdateTx updates the bank account by ID using the transactor provided.
func (bar *BankAccountRepository) UpdateTx(
	ctx context.Context,
	transactor service.Transactor,
	ba service.BankAccount,
) error {
	tx, ok := transactor.(*sqlx.Tx)
	if !ok {
		return TxTypeError{tx: tx}
	}

	baRow := bankAccountFromDomain(ba)

	if _, err := tx.NamedExecContext(
		ctx,
		bar.queries[_updateBankAccount],
		baRow,
	); err != nil {
		return fmt.Errorf("update bank account with ID %d: %w", baRow.ID, err)
	}

	return nil
}

// Insert inserts a new row into the bank_accounts table, returning the inserted
// account.
func (bar *BankAccountRepository) Insert(
	ctx context.Context,
	ba service.BankAccount,
) (service.BankAccount, error) {
	row := bankAccountFromDomain(ba)

	query, args, err := bar.db.BindNamed(bar.queries[_insertBankAccounts], row)
	if err != nil {
		return ba, fmt.Errorf("prepare _insertBankAccount query: %w", err)
	}

	if err := bar.db.Get(ctx, &row, query, args...); err != nil {
		return ba, fmt.Errorf("insert into bank_accounts: %w", err)
	}

	return row.toDomain(), nil
}

// Count returns the number of rows in the bank_accounts table.
func (bar *BankAccountRepository) Count(ctx context.Context) (int64, error) {
	var count int64

	if err := bar.db.Get(ctx, &count, bar.queries[_countBankAccounts]); err != nil {
		return 0, fmt.Errorf("count bank_accounts: %w", err)
	}

	return count, nil
}

func (bar *BankAccountRepository) FindByID(
	ctx context.Context,
	id int64,
) (service.BankAccount, error) {
	var row BankAccount

	if err := bar.db.Get(
		ctx,
		&row,
		bar.queries[_findBankAccountByID],
		id,
	); err != nil {
		return service.BankAccount{}, fmt.Errorf("get bank account with ID %d: %w",
			id, err)
	}

	return row.toDomain(), nil
}

// Truncate truncates the bank_accounts table.
func (bar *BankAccountRepository) Truncate(ctx context.Context) error {
	if !truncationPermitted(bar.appConfig.Env) {
		return UnpermittedTruncationError{env: bar.appConfig.Env}
	}

	if _, err := bar.db.Exec(ctx, bar.queries[_truncateBankAccounts]); err != nil {
		return fmt.Errorf("truncate bank_accounts: %w", err)
	}

	return nil
}

func (bar *BankAccountRepository) loadQueries() error {
	queryDir := filepath.Join(bar.appConfig.Root, RelativeQueryDir(), "bank_accounts")
	queryFilenames := []QueryFilename{
		_countBankAccounts,
		_findBankAccountByIBAN,
		_findBankAccountByID,
		_insertBankAccounts,
		_truncateBankAccounts,
		_updateBankAccount,
	}

	if bar.queries == nil {
		bar.queries = make(Queries, len(queryFilenames))
	}

	if err := bar.queries.Load(queryDir, queryFilenames); err != nil {
		return fmt.Errorf("load bank account queries: %w", err)
	}

	return nil
}
