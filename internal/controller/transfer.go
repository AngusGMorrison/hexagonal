package controller

import (
	"context"
	"errors"
	"fmt"
)

// ErrInsufficientFunds signals that a bank account did not have enough funds to
// complete a bulk transfer.
var ErrInsufficientFunds = errors.New("insufficient funds to settle bulk transfer")

// Logger specificies a simple logging interface.
type Logger interface {
	Printf(format string, args ...any)
}

// Transactor behaves like a transaction.
type Transactor interface {
	Commit() error

	// Transactors must guarantee that a repository is left in its clean,
	// pre-transaction state even when Rollback returns an error.
	Rollback() error
}

// AtomicTransferRepository specifies the methods required to save BulkTransfers
// to a data store atomically.
type AtomicTransferRepository interface {
	// BeginSerializableTx must return a Transactor that guarantees a
	// transaction that is serializable with all concurrent repository
	// operations.
	BeginSerializableTx(ctx context.Context) (Transactor, error)

	GetBankAccountByIBANTx(ctx context.Context, tx Transactor, iban string) (BankAccount, error)
	UpdateBankAccountTx(ctx context.Context, tx Transactor, ba BankAccount) error
	SaveCreditTransfersTx(ctx context.Context, tx Transactor, transfers CreditTransfers) error
}

// TransferController provides the fields and methods required to perform bulk
// transfers.
type TransferController struct {
	logger Logger
	repo   AtomicTransferRepository
}

// NewTransferController configures and returns a Service.
func NewTransferController(logger Logger, repo AtomicTransferRepository) *TransferController {
	return &TransferController{
		logger: logger,
		repo:   repo,
	}
}

// PerformBulkTransfer exposes the underlying repo's PerformBulkTransfer method
// as a convenience.
func (tc *TransferController) PerformBulkTransfer(ctx context.Context, bt BulkTransfer) error {
	tx, err := tc.repo.BeginSerializableTx(ctx)
	if err != nil {
		return fmt.Errorf("tc.repo.BeginTx: %w", err)
	}

	defer func() {
		if err := tx.Rollback(); err != nil {
			tc.logger.Printf("Failed to roll back transaction: %v\n", err)
		}
	}()

	bankAccount, err := tc.repo.GetBankAccountByIBANTx(ctx, tx, bt.Account.OrganizationIBAN)
	if err != nil {
		return fmt.Errorf("tc.repo.GetBankAccountByIBAN: %w", err)
	}

	bankAccount.BalanceCents -= bt.TotalCents()
	if bankAccount.BalanceCents < 0 {
		return ErrInsufficientFunds
	}

	if err := tc.repo.UpdateBankAccountTx(ctx, tx, bankAccount); err != nil {
		return fmt.Errorf("tc.repo.UpdateBankAccount: %w", err)
	}

	creditTransfers := bt.CreditTransfers.assignBankAccountID(bankAccount.ID)

	if err := tc.repo.SaveCreditTransfersTx(ctx, tx, creditTransfers); err != nil {
		return fmt.Errorf("tc.repo.SaveCreditTransfers: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("tx.Commit: %w", err)
	}

	return nil
}

// BulkTransfer represents a bulk transfer.
type BulkTransfer struct {
	Account         BankAccount
	CreditTransfers CreditTransfers
}

// TotalCents returns the total value of the bulk transfer in cents.
func (bt BulkTransfer) TotalCents() int64 {
	var total int64

	for _, transfer := range bt.CreditTransfers {
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

// CreditTransfers wraps []CreditTransfer to provide convenience methods on the
// slice.
type CreditTransfers []CreditTransfer

func (cts CreditTransfers) assignBankAccountID(id int64) CreditTransfers {
	transfersWithID := make(CreditTransfers, 0, len(cts))

	for _, transfer := range cts {
		transfer.BankAccountID = id
		transfersWithID = append(transfersWithID, transfer)
	}

	return transfersWithID
}

// CreditTransfer represents a single credit transfer involved in a bulk
// transfer.
type CreditTransfer struct {
	ID               int64
	BankAccountID    int64
	AmountCents      int64
	Currency         string
	CounterpartyName string
	CounterpartyBIC  string
	CounterpartyIBAN string
	Description      string
}
