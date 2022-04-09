package transferrepo

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/angusgmorrison/hexagonal/internal/adapter/envconfig"
	"github.com/angusgmorrison/hexagonal/internal/adapter/repository/postgres"
	"github.com/angusgmorrison/hexagonal/internal/controller"
	"github.com/jmoiron/sqlx"
)

// Repository provides access to the database tables of the transfers domain,
// including bank account and transactions.
type Repository struct {
	postgres *postgres.Postgres
	queries  queries
}

// Statically verify that Repository satisfies controller.Repository.
var _ controller.TransferRepository = (*Repository)(nil)

// New returns a new Repository.
func New(pg *postgres.Postgres, appConfig envconfig.App) (*Repository, error) {
	q, err := loadQueries(appConfig.Root)
	if err != nil {
		return nil, fmt.Errorf("loadQueries: %w", err)
	}

	repo := Repository{
		postgres: pg,
		queries:  q,
	}

	return &repo, nil
}

// PerformBulkTransfer executes the provided transfers atomically against a
// single bank account. validate is called before committing the transaction. If
// validation fails, the validation error is returned and the transaction is
// rolled back.
func (r *Repository) PerformBulkTransfer(
	ctx context.Context,
	bulkTransfer controller.BulkTransfer,
	validate controller.BulkTransferValidator,
) error {
	tx, err := r.postgres.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
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
func (r *Repository) debitAccount(
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
func (r *Repository) createTransactions(
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

func (r *Repository) getBankAccountWhereIBANTx(
	ctx context.Context, tx *sqlx.Tx, iban string,
) (controller.BankAccount, error) {
	var row BankAccountRow
	if err := tx.GetContext(ctx, &row, r.queries.getBankAccountByIBAN(), iban); err != nil {
		return controller.BankAccount{}, fmt.Errorf("get bank account with IBAN %q: %w",
			iban, err)
	}

	return row.toDomain(), nil
}

func (r *Repository) updateBankAccountTx(
	ctx context.Context, tx *sqlx.Tx, account controller.BankAccount,
) (controller.BankAccount, error) {
	stmt, err := tx.PrepareNamedContext(ctx, r.queries.updateBankAccount())
	if err != nil {
		return account, fmt.Errorf("prepare updateBankAccountQuery: %w", err)
	}

	row := bankAccountRowFromDomain(account)

	if err := stmt.QueryRowxContext(ctx, row).StructScan(&row); err != nil {
		return account, fmt.Errorf("update bank account with ID %d: %w", row.ID, err)
	}

	return row.toDomain(), nil
}

func (r *Repository) insertTransactionsTx(
	ctx context.Context, tx *sqlx.Tx, creditTransfers []controller.CreditTransfer,
) error {
	transactionRows := transactionRowsFromDomain(creditTransfers)

	if _, err := tx.NamedExecContext(ctx, r.queries.insertTransactions(), transactionRows); err != nil {
		return fmt.Errorf("insert transactions: %w", err)
	}

	return nil
}
