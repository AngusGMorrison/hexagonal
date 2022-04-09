package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/angusgmorrison/hexagonal/internal/adapter/envconfig"
	"github.com/angusgmorrison/hexagonal/internal/controller"
	"github.com/jmoiron/sqlx"
)

const (
	_getBankAccountByIBANQueryKey = "get_bank_account_by_iban.sql"
	_updateBankAccountQueryKey    = "update_bank_account.sql"
	_insertTransactionsQueryKey   = "insert_transactions.sql"
)

// TransferRepository provides access to the database tables of the transfers domain,
// including bank account and transactions.
type TransferRepository struct {
	db      *DB
	queries queries
}

// Statically verify that Repository satisfies controller.Repository.
var _ controller.TransferRepository = (*TransferRepository)(nil)

// NewTransferRepository returns a new TransferRepository.
func NewTransferRepository(db *DB, appConfig envconfig.App) (*TransferRepository, error) {
	q, err := loadQueries(appConfig.Root, transferQueryFilenames())
	if err != nil {
		return nil, fmt.Errorf("loadQueries: %w", err)
	}

	repo := TransferRepository{
		db:      db,
		queries: q,
	}

	return &repo, nil
}

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

// PerformBulkTransfer executes the provided transfers atomically against a
// single bank account. validate is called before committing the transaction. If
// validation fails, the validation error is returned and the transaction is
// rolled back.
func (r *TransferRepository) PerformBulkTransfer(
	ctx context.Context,
	bulkTransfer controller.BulkTransfer,
	validate controller.BulkTransferValidator,
) error {
	tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	defer tx.Rollback() //nolint:errcheck

	bulkTransfer, err = r.debitAccount(ctx, tx, bulkTransfer)
	if err != nil {
		return err
	}

	bulkTransfer, err = r.createTransactions(ctx, tx, bulkTransfer)
	if err != nil {
		return err
	}

	if err = validate(bulkTransfer); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// debitAccount fetches the target account from the DB by IBAN and updates its
// balance. The resulting state of the account is reflected in the returned
// BulkTransfer.
func (r *TransferRepository) debitAccount(
	ctx context.Context, tx *sqlx.Tx, bulkTransfer controller.BulkTransfer,
) (controller.BulkTransfer, error) {
	account, err := r.getBankAccountWhereIBANTx(ctx, tx, bulkTransfer.Account.OrganizationIBAN)
	if err != nil {
		return bulkTransfer, err
	}

	account.BalanceCents -= bulkTransfer.TotalCents()

	if account, err = r.updateBankAccountTx(ctx, tx, account); err != nil {
		return bulkTransfer, err
	}

	bulkTransfer.Account = account

	return bulkTransfer, nil
}

// createTransactions inserts the BulkTransfer's creditTransfers into the
// database as transactions using the bulkTransfer.BankAccount.ID as the foreign
// key. It then returns an updated BulkTransfer to reflect this change.
func (r *TransferRepository) createTransactions(
	ctx context.Context, tx *sqlx.Tx, bulkTransfer controller.BulkTransfer,
) (controller.BulkTransfer, error) {
	// Copy the credit transfers from the input bulkTranfer, assigning the
	// account ID to each. Copying preserves the original state of the input
	// bulkTransfer in the event of an error.
	creditTransfers := make([]controller.CreditTransfer, len(bulkTransfer.CreditTransfers))

	for i, transfer := range bulkTransfer.CreditTransfers {
		transfer.BankAccountID = bulkTransfer.Account.ID
		creditTransfers[i] = transfer
	}

	err := r.insertTransactionsTx(ctx, tx, creditTransfers)
	if err != nil {
		return bulkTransfer, err
	}

	return bulkTransfer, nil
}

func (r *TransferRepository) getBankAccountWhereIBANTx(
	ctx context.Context, tx *sqlx.Tx, iban string,
) (controller.BankAccount, error) {
	var row BankAccountRow
	if err := tx.GetContext(ctx, &row, r.queries[_getBankAccountByIBANQueryKey], iban); err != nil {
		return controller.BankAccount{}, fmt.Errorf("get bank account with IBAN %q: %w",
			iban, err)
	}

	return row.toDomain(), nil
}

func (r *TransferRepository) updateBankAccountTx(
	ctx context.Context, tx *sqlx.Tx, account controller.BankAccount,
) (controller.BankAccount, error) {
	stmt, err := tx.PrepareNamedContext(ctx, r.queries[_updateBankAccountQueryKey])
	if err != nil {
		return account, fmt.Errorf("prepare updateBankAccountQuery: %w", err)
	}

	row := bankAccountRowFromDomain(account)

	if err := stmt.QueryRowxContext(ctx, row).StructScan(&row); err != nil {
		return account, fmt.Errorf("update bank account with ID %d: %w", row.ID, err)
	}

	return row.toDomain(), nil
}

func (r *TransferRepository) insertTransactionsTx(
	ctx context.Context, tx *sqlx.Tx, creditTransfers []controller.CreditTransfer,
) error {
	transactionRows := transactionRowsFromDomain(creditTransfers)

	if _, err := tx.NamedExecContext(
		ctx, r.queries[_insertTransactionsQueryKey], transactionRows,
	); err != nil {
		return fmt.Errorf("insert transactions: %w", err)
	}

	return nil
}

func transferQueryFilenames() []string {
	return []string{
		_getBankAccountByIBANQueryKey,
		_updateBankAccountQueryKey,
		_insertTransactionsQueryKey,
	}
}
