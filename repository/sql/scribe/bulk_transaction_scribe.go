package scribe

import (
	"context"
	"fmt"

	"github.com/angusgmorrison/hexagonal/repository/sql"
	"github.com/angusgmorrison/hexagonal/repository/sql/table/bankaccounts"
	"github.com/angusgmorrison/hexagonal/repository/sql/table/transactions"
	"github.com/angusgmorrison/hexagonal/service"
)

// BulkTransactionScribe is a single-use, serializable scribe that controls
// atomic updates to the bank_accounts and transactions tables.
type BulkTransactionScribe struct {
	atomicScribe *atomicScribe
}

var _ service.BulkTransactionRepository = (*BulkTransactionScribe)(nil)

// NewBulkTransactionScribeFactory returns a scribe factory function which has
// captured a reference to a database. This makes it trivial for business logic
// to instantiate a new scribe for a single database transaction,
// avoiding the need for scribe thread safety.
//
// Scribes returned by the factory contain a live transaction against which the
// scribe's database operations are run. For this reason, scribes should be used
// immediately after instantiation, since long-lived scribes will leak database
// connections. Attempts to reuse the scribe after the transaction is committed
// or rolled back return errors.
func NewBulkTransactionScribeFactory(db sql.Database) service.BulkTransactionRepositoryFactory {
	return func(ctx context.Context) (service.BulkTransactionRepository, error) {
		var (
			atomicScribe = atomicScribe{db: db}
			err          = atomicScribe.BeginSerializable(ctx)
			btScribe     = BulkTransactionScribe{atomicScribe: &atomicScribe}
		)

		return &btScribe, err
	}
}

// GetBankAccountByIBAN populates the BankAccount of the given BulkTransaction
// using its IBAN. Calling GetBankAccountByIBAN after calling Save is an error.
func (bts *BulkTransactionScribe) GetBankAccountByIBAN(
	ctx context.Context,
	bt service.BulkTransaction,
) (service.BulkTransaction, error) {
	bankAccount, err := bankaccounts.FindByIBAN(ctx, bts.atomicScribe.tx, bt.BankAccount.OrganizationIBAN)
	if err != nil {
		_ = bts.atomicScribe.Rollback()
		return bt, fmt.Errorf("LoadBankAccountFromIBAN: %w", err)
	}

	bt.BankAccount = bankAccount

	return bt, nil
}

// Save persists the BulkTransaction's BankAccount and Transactions and commits
// the underlying database transaction. Calling Save more than once is an error.
func (bts *BulkTransactionScribe) Save(
	ctx context.Context,
	bt service.BulkTransaction,
) (service.BulkTransaction, error) {
	defer func() { _ = bts.atomicScribe.Rollback() }()

	bankAccount, err := bankaccounts.Update(ctx, bts.atomicScribe.tx, bt.BankAccount)
	if err != nil {
		return bt, fmt.Errorf("save BankAccount: %w", err)
	}

	txns, err := transactions.BulkInsert(ctx, bts.atomicScribe.tx, bt.Transactions)
	if err != nil {
		return bt, fmt.Errorf("save Transactions: %w", err)
	}

	bt.BankAccount = bankAccount
	bt.Transactions = txns

	if err := bts.atomicScribe.Commit(); err != nil {
		return bt, fmt.Errorf("save BulkTransaction: %w", err)
	}

	return bt, nil
}

// Abort rolls back the database transaction. The database will be left in its
// original state even if Abort returns an error. Calling Abort after calling
// Save is an error.
func (bts *BulkTransactionScribe) Abort() error {
	if err := bts.atomicScribe.Rollback(); err != nil {
		return fmt.Errorf("abort bulk transaction: %w", err)
	}

	return nil
}
