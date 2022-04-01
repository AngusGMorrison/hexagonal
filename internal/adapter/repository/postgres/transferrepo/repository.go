package transferrepo

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/angusgmorrison/hexagonal/internal/adapter/repository/postgres"
	"github.com/angusgmorrison/hexagonal/internal/app/transferdomain"
	"github.com/jmoiron/sqlx"
)

// Repository provides access to the database tables of the transfers domain,
// including bank account and transactions.
type Repository struct {
	logger *log.Logger
	*postgres.Postgres
}

// Statically verify that Repository satisfies transferdomain.Repository.
var _ transferdomain.Repository = (*Repository)(nil)

func New(p *postgres.Postgres, logger *log.Logger) *Repository {
	return &Repository{
		logger:   logger,
		Postgres: p,
	}
}

const getBankAccountQuery = `
SELECT id, organization_name, balance_cents, iban, bic
FROM bank_accounts
WHERE iban = $1;
`

const updateBankAccountQuery = `
UPDATE bank_accounts
SET
	balance_cents = :balance_cents, organization_name = :organization_name,
	iban = :iban, bic = :bic
WHERE id = :id
RETURNING id, balance_cents, organization_name, iban, bic;
`

const insertTransactionsQuery = `
INSERT INTO transactions (
	counterparty_name, counterparty_iban, counterparty_bic, amount_cents,
	amount_currency, bank_account_id, description
) VALUES (
	:counterparty_name, :counterparty_iban, :counterparty_bic, :amount_cents,
	:amount_currency, :bank_account_id, :description
) RETURNING id, counterparty_name, counterparty_iban, counterparty_bic
	amount_cents, amount_currency, bank_account_id, description;`

// PerformBulkTransfer executes the provided transfers atomically against a
// single bank account. validate is called before committing the transaction. If
// validation fails, the validation error is returned and the transaction is
// rolled back.
func (r *Repository) PerformBulkTransfer(ctx context.Context, bulkTransfer transferdomain.BulkTransfer, validate transferdomain.BulkTransferValidator) error {
	tx, err := r.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("repository.BulkTransfer failed to begin transaction: %w", err)
	}

	defer tx.Rollback() //nolint:errcheck

	bulkTransfer, err = debitAccount(ctx, tx, bulkTransfer)
	if err != nil {
		return err
	}

	bulkTransfer, err = createTransactions(ctx, tx, bulkTransfer)
	if err != nil {
		return err
	}

	if err = validate(bulkTransfer); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("repository.BulkTransfer failed to commit transaction: %w", err)
	}

	return nil
}

// debitAccount fetches the target account from the DB by IBAN and updates its
// balance. The resulting state of the account is reflected in the returned
// BulkTransfer.
func debitAccount(ctx context.Context, tx *sqlx.Tx, bulkTransfer transferdomain.BulkTransfer) (transferdomain.BulkTransfer, error) {
	account, err := getBankAccountWhereIBANTx(ctx, tx, bulkTransfer.Account.OrganizationIBAN)
	if err != nil {
		return bulkTransfer, err
	}

	account.BalanceCents -= bulkTransfer.TotalCents()

	if account, err = updateBankAccountTx(ctx, tx, account); err != nil {
		return bulkTransfer, err
	}

	bulkTransfer.Account = account

	return bulkTransfer, nil
}

// createTransactions inserts the BulkTransfer's creditTransfers into the
// database as transactions using the bulkTransfer.BankAccount.ID as the foreign
// key. It then returns an updated BulkTransfer to reflect this change.
func createTransactions(ctx context.Context, tx *sqlx.Tx, bulkTransfer transferdomain.BulkTransfer) (transferdomain.BulkTransfer, error) {
	// Copy the credit transfers from the input bulkTranfer, assigning the
	// account ID to each. Copying preserves the original state of the input
	// bulkTransfer in the event of an error.
	creditTransfers := make([]transferdomain.CreditTransfer, len(bulkTransfer.CreditTransfers))

	for i, transfer := range bulkTransfer.CreditTransfers {
		transfer.BankAccountID = bulkTransfer.Account.ID
		creditTransfers[i] = transfer
	}

	err := insertTransactionsTx(ctx, tx, creditTransfers)
	if err != nil {
		return bulkTransfer, err
	}

	return bulkTransfer, nil
}

func getBankAccountWhereIBANTx(ctx context.Context, tx *sqlx.Tx, iban string) (transferdomain.BankAccount, error) {
	var row bankAccountRow
	if err := tx.GetContext(ctx, &row, getBankAccountQuery, iban); err != nil {
		return transferdomain.BankAccount{}, fmt.Errorf("failed to get bank account with IBAN %q: %w",
			iban, err)
	}

	return row.toDomain(), nil
}

func updateBankAccountTx(ctx context.Context, tx *sqlx.Tx, account transferdomain.BankAccount) (transferdomain.BankAccount, error) {
	stmt, err := tx.PrepareNamedContext(ctx, updateBankAccountQuery)
	if err != nil {
		return account, fmt.Errorf("failed to prepare updateBankAccountQuery: %w", err)
	}

	row := bankAccountRowFromDomain(account)

	if err := stmt.QueryRowxContext(ctx, row).StructScan(&row); err != nil {
		return account, fmt.Errorf("failed to update bank account with ID %d: %w", row.ID, err)
	}

	return row.toDomain(), nil
}

func insertTransactionsTx(ctx context.Context, tx *sqlx.Tx, creditTransfers []transferdomain.CreditTransfer) error {
	transactionRows := transactionRowsFromDomain(creditTransfers)

	if _, err := tx.NamedExecContext(ctx, insertTransactionsQuery, transactionRows); err != nil {
		return fmt.Errorf("failed to insert transactions: %w", err)
	}

	return nil
}
