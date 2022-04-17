package service

import (
	"context"
	"errors"
	"fmt"
)

// ErrInsufficientFunds signals that a bank account did not have enough funds to
// complete a bulk transfer.
var ErrInsufficientFunds = errors.New("insufficient funds to settle bulk transfer")

type logger interface {
	Printf(format string, args ...any)
}

// BulkTransactionRepository represents an atomic, serializable store of bulk
// transaction data. Implementers must guarantee that all operations are atomic
// and serializable.
type BulkTransactionRepository interface {
	GetBankAccountByIBAN(ctx context.Context, bt BulkTransaction) (BulkTransaction, error)

	// Save the BulkTransaction to the data store.
	Save(ctx context.Context, bt BulkTransaction) (BulkTransaction, error)

	// Abort the bulk transaction. The repository must be left in the the state
	// it was in before the bulk transaction was attempted, even in the event of
	// an error.
	//
	// Abort should have no effect after calling Save.
	Abort() error
}

// BulkTransactionRepositoryFactory functions return new repositories that are
// ready to be used and can be discarded when done.
type BulkTransactionRepositoryFactory func(ctx context.Context) (BulkTransactionRepository, error)

// BulkTransactionService specifies the fields and methods required to perform
// bulk transactions.
//
// Exporting an interface for which this package also provides
// an implemention makes it simple for dependent packages to mock the service
// while ensuring that the service package is authoritative.
type BulkTransactionService interface {
	BulkTransaction(ctx context.Context, bt BulkTransaction) error
}

// bulkTransactionService implements the BulkTransactionService interface.
type bulkTransactionService struct {
	logger logger

	// newRepository allows the service to create new, disposable atomic
	// repositories on the fly, avoiding the need to handle transactions in the
	// service layer.
	newRepository BulkTransactionRepositoryFactory
}

// NewBulkTransactionService configures and returns a Service.
func NewBulkTransactionService(
	logger logger,
	repoFactory BulkTransactionRepositoryFactory,
) BulkTransactionService {
	return &bulkTransactionService{
		logger:        logger,
		newRepository: repoFactory,
	}
}

// BulkTransaction executes the given BulkTransaction atomically. If the target
// bank account has insufficient funds to settle all transactions, none are
// applied. Otherwise, the bank balance is updated and the transactions are
// persisted.
func (tc *bulkTransactionService) BulkTransaction(ctx context.Context, bt BulkTransaction) error {
	// Spawn a new atomic repository for the lifetime of the function.
	repo, err := tc.newRepository(ctx)
	if err != nil {
		return fmt.Errorf("tc.newRepository: %w", err)
	}

	defer func() {
		if err := repo.Abort(); err != nil {
			tc.logger.Printf("BulkTransactionRepsitory.Rollback: %v\n", err)
		}
	}()

	bt, err = repo.GetBankAccountByIBAN(ctx, bt)
	if err != nil {
		return fmt.Errorf("(*TransactionService).BulkTransaction: %w", err)
	}

	if bt.insufficientFunds() {
		return ErrInsufficientFunds
	}

	bt = bt.applyTransactions()

	bt, err = repo.Save(ctx, bt)
	if err != nil {
		return fmt.Errorf("*TransactionService).BulkTransaction: %w", err)
	}

	return nil
}

// BulkTransaction represents a bulk transfer of funds to single bank account.
type BulkTransaction struct {
	BankAccount  BankAccount
	Transactions Transactions
}

// TotalCents returns the total value of the bulk transfer in cents.
func (bt BulkTransaction) totalCents() int64 {
	var total int64

	for _, transfer := range bt.Transactions {
		total += transfer.AmountCents
	}

	return total
}

func (bt BulkTransaction) insufficientFunds() bool {
	return bt.totalCents() > bt.BankAccount.BalanceCents
}

func (bt BulkTransaction) applyTransactions() BulkTransaction {
	bt.BankAccount.BalanceCents -= bt.totalCents()

	return bt.assignBankAccountIDToTransactions()
}

func (bt BulkTransaction) assignBankAccountIDToTransactions() BulkTransaction {
	txnsWithID := make(Transactions, 0, len(bt.Transactions))

	for _, transaction := range bt.Transactions {
		transaction.BankAccountID = bt.BankAccount.ID
		txnsWithID = append(txnsWithID, transaction)
	}

	bt.Transactions = txnsWithID

	return bt
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

func (ts Transactions) assignBankAccountID(id int64) {
	for i, transfer := range ts {
		transfer.BankAccountID = id
		ts[i] = transfer
	}
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
