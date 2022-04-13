package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/angusgmorrison/hexagonal/internal/controller"
	"github.com/jmoiron/sqlx"
)

// BankAccountRow represents a row of the bank_accounts table.
type BankAccountRow struct {
	ID               int64  `db:"id"`
	OrganizationName string `db:"organization_name"`
	BalanceCents     int64  `db:"balance_cents"`
	IBAN             string `db:"iban"`
	BIC              string `db:"bic"`
}

func bankAccountRowFromDomain(domainAccount controller.BankAccount) BankAccountRow {
	return BankAccountRow{
		ID:               domainAccount.ID,
		OrganizationName: domainAccount.OrganizationName,
		IBAN:             domainAccount.OrganizationIBAN,
		BIC:              domainAccount.OrganizationBIC,
		BalanceCents:     domainAccount.BalanceCents,
	}
}

func (ba BankAccountRow) toDomain() controller.BankAccount {
	return controller.BankAccount{
		ID:               ba.ID,
		OrganizationName: ba.OrganizationName,
		OrganizationIBAN: ba.IBAN,
		OrganizationBIC:  ba.BIC,
		BalanceCents:     ba.BalanceCents,
	}
}

// TransactionRows is a convenience wrapper around one or more instances of
// transactionRow.
type TransactionRows []TransactionRow

func transactionRowsFromDomain(domainTransfers []controller.CreditTransfer) TransactionRows {
	rows := make(TransactionRows, len(domainTransfers))

	for i, dt := range domainTransfers {
		rows[i] = transactionRowFromDomain(dt)
	}

	return rows
}

// TransactionRow represents a row of the transactions table.
type TransactionRow struct {
	ID               int64  `db:"id"`
	BankAccountID    int64  `db:"bank_account_id"`
	CounterpartyName string `db:"counterparty_name"`
	CounterpartyIBAN string `db:"counterparty_iban"`
	CounterpartyBIC  string `db:"counterparty_bic"`
	AmountCents      int64  `db:"amount_cents"`
	AmountCurrency   string `db:"amount_currency"`
	Description      string `db:"description"`
}

func transactionRowFromDomain(domainTransfer controller.CreditTransfer) TransactionRow {
	return TransactionRow{
		ID:               domainTransfer.ID,
		BankAccountID:    domainTransfer.BankAccountID,
		CounterpartyName: domainTransfer.CounterpartyName,
		CounterpartyBIC:  domainTransfer.CounterpartyBIC,
		CounterpartyIBAN: domainTransfer.CounterpartyIBAN,
		AmountCents:      domainTransfer.AmountCents,
		AmountCurrency:   domainTransfer.Currency,
		Description:      domainTransfer.Description,
	}
}

const (
	_getBankAccountByIBAN QueryFilename = "get_bank_account_by_iban.sql"
	_updateBankAccount    QueryFilename = "update_bank_account.sql"
	_insertTransactions   QueryFilename = "insert_transactions.sql"
)

// TransferRepository provides the methods necessary to perform transfers.
// Satisfies contorller.AtomicTransferRepository.
type TransferRepository struct {
	db       *DB
	queryDir string
	queries  Queries
}

// Statically verify that Repository satisfies controller.Repository.
var _ controller.AtomicTransferRepository = (*TransferRepository)(nil)

// NewTransferRepository returns a new transfer repository with its SQL queries
// preloaded.
//
// queryDir is the absolute path to a directory containing the SQL files
// required by TransferRepository.
func NewTransferRepository(db *DB, queryDir string) (*TransferRepository, error) {
	repo := TransferRepository{
		db:       db,
		queryDir: queryDir,
	}

	if err := repo.loadQueries(); err != nil {
		return nil, err
	}

	return &repo, nil
}

// BeginSerializableTx starts a new sqlx transaction with isolation level
// Serializable and returns it as a controller.Transactor for use in atomic
// repository operations.
func (tr *TransferRepository) BeginSerializableTx(ctx context.Context) (controller.Transactor, error) {
	tx, err := tr.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return nil, fmt.Errorf("BeginSerializableTx: %w", err)
	}

	return tx, nil
}

// GetBankAccountByIBANTx retrieves the bank account with the given IBAN using the
// transactor. If there is no matching bank account, an error is returned.
func (tr *TransferRepository) GetBankAccountByIBANTx(
	ctx context.Context,
	transactor controller.Transactor,
	iban string,
) (controller.BankAccount, error) {
	tx, ok := transactor.(*sqlx.Tx)
	if !ok {
		return controller.BankAccount{}, TxTypeError{tx: tx}
	}

	var row BankAccountRow

	if err := tx.GetContext(
		ctx,
		&row,
		tr.queries[_getBankAccountByIBAN],
		iban,
	); err != nil {
		return controller.BankAccount{}, fmt.Errorf("get bank account with IBAN %q: %w",
			iban, err)
	}

	return row.toDomain(), nil
}

// UpdateBankAccountTx updates the bank account by ID using the transactor
// provided.
func (tr *TransferRepository) UpdateBankAccountTx(
	ctx context.Context,
	transactor controller.Transactor,
	ba controller.BankAccount,
) error {
	tx, ok := transactor.(*sqlx.Tx)
	if !ok {
		return TxTypeError{tx: tx}
	}

	baRow := bankAccountRowFromDomain(ba)

	if _, err := tx.NamedExecContext(
		ctx,
		tr.queries[_updateBankAccount],
		baRow,
	); err != nil {
		return fmt.Errorf("update bank account with ID %d: %w", baRow.ID, err)
	}

	return nil
}

// SaveCreditTransfersTx bulk inserts credit transfers using the transactor
// provided.
func (tr *TransferRepository) SaveCreditTransfersTx(
	ctx context.Context,
	transactor controller.Transactor,
	transfers controller.CreditTransfers,
) error {
	tx, ok := transactor.(*sqlx.Tx)
	if !ok {
		return TxTypeError{tx: tx}
	}

	rows := transactionRowsFromDomain(transfers)

	if _, err := tx.NamedExecContext(
		ctx,
		tr.queries[_insertTransactions],
		rows,
	); err != nil {
		return fmt.Errorf("insert transactions: %w", err)
	}

	return nil
}

func (tr *TransferRepository) loadQueries() error {
	queryFilenames := []QueryFilename{
		_getBankAccountByIBAN,
		_updateBankAccount,
		_insertTransactions,
	}

	if tr.queries == nil {
		tr.queries = make(Queries, len(queryFilenames))
	}

	if err := tr.queries.Load(tr.queryDir, queryFilenames); err != nil {
		return fmt.Errorf("load transfer queries: %w", err)
	}

	return nil
}
