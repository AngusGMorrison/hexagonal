package controller

import (
	"context"
	"errors"
	"fmt"
)

// ErrInsufficientFunds signals that a bank account did not have enough funds to
// complete a bulk transfer.
var ErrInsufficientFunds = errors.New("insufficient funds to settle bulk transfer")

// Transactor behaves like a database transaction.
type Transactor interface {
	Commit() error

	// Transactors must guarantee that a repository is left in its clean,
	// pre-transaction state even when Rollback returns an error.
	Rollback() error
}

// AtomicBankAccountRepository describes an object that can atomically fetch and
// persist BankAccounts from a data store.
type AtomicBankAccountRepository interface {
	BeginSerializableTx(ctx context.Context) (Transactor, error)
	FindByIBANTx(ctx context.Context, tx Transactor, iban string) (BankAccount, error)
	UpdateTx(ctx context.Context, tx Transactor, ba BankAccount) error
}

// AtomicTransactionRepository describes an object that can atomically fetch and
// persist Transactions from a data store.
type AtomicTransactionRepository interface {
	BulkInsertTx(ctx context.Context, tx Transactor, transfers Transactions) error
}

type logger interface {
	Printf(format string, args ...any)
}

// TransactionController provides the fields and methods required to perform bulk
// transfers.
type TransactionController struct {
	logger          logger
	bankAccountRepo AtomicBankAccountRepository
	transactionRepo AtomicTransactionRepository
}

// NewTransactionController configures and returns a Service.
func NewTransactionController(
	logger logger,
	bankAccountRepo AtomicBankAccountRepository,
	transactionRepo AtomicTransactionRepository,
) *TransactionController {
	return &TransactionController{
		logger:          logger,
		bankAccountRepo: bankAccountRepo,
		transactionRepo: transactionRepo,
	}
}

// BulkTransaction executes the given BulkTransaction atomically. If the target
// bank account has insufficient funds to settle all transactions, none are
// applied. Otherwise, the bank balance is updated and the transactions are
// persisted.
func (tc *TransactionController) BulkTransaction(ctx context.Context, bt BulkTransaction) error {
	tx, err := tc.bankAccountRepo.BeginSerializableTx(ctx)
	if err != nil {
		return fmt.Errorf("tc.repo.BeginTx: %w", err)
	}

	defer func() {
		if err := tx.Rollback(); err != nil {
			tc.logger.Printf("Failed to roll back transaction: %v\n", err)
		}
	}()

	bankAccount, err := tc.bankAccountRepo.FindByIBANTx(ctx, tx, bt.Account.OrganizationIBAN)
	if err != nil {
		return fmt.Errorf("tc.repo.GetBankAccountByIBAN: %w", err)
	}

	bankAccount.BalanceCents -= bt.TotalCents()
	if bankAccount.BalanceCents < 0 {
		return ErrInsufficientFunds
	}

	if err := tc.bankAccountRepo.UpdateTx(ctx, tx, bankAccount); err != nil {
		return fmt.Errorf("tc.repo.UpdateBankAccount: %w", err)
	}

	transactions := bt.Transactions.withBankAccountID(bankAccount.ID)

	if err := tc.transactionRepo.BulkInsertTx(ctx, tx, transactions); err != nil {
		return fmt.Errorf("tc.repo.SaveCreditTransfers: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("tx.Commit: %w", err)
	}

	return nil
}

// BulkTransaction represents a bulk transfer of funds to single bank account.
type BulkTransaction struct {
	Account      BankAccount
	Transactions Transactions
}

// TotalCents returns the total value of the bulk transfer in cents.
func (bt BulkTransaction) TotalCents() int64 {
	var total int64

	for _, transfer := range bt.Transactions {
		total += transfer.AmountCents
	}

	return total
}

// BankAccount represents a bank account involved in a bulk transfer and its
// associated business logic.
type BankAccount struct {
	ID               int64
	OrganizationName string
	OrganizationBIC  string
	OrganizationIBAN string
	BalanceCents     int64
}

// Transactions provides convenience methods on []Transaction.
type Transactions []Transaction

func (ts Transactions) withBankAccountID(id int64) Transactions {
	transfersWithID := make(Transactions, 0, len(ts))

	for _, transfer := range ts {
		transfer.BankAccountID = id
		transfersWithID = append(transfersWithID, transfer)
	}

	return transfersWithID
}

// Transaction represents a single credit transfer involved in a bulk
// transfer.
type Transaction struct {
	ID               int64
	BankAccountID    int64
	AmountCents      int64
	Currency         string
	CounterpartyName string
	CounterpartyBIC  string
	CounterpartyIBAN string
	Description      string
}
