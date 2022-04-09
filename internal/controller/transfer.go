package controller

import (
	"context"
	"errors"
	"fmt"
)

// ErrInsufficientFunds signals that a bank account did not have enough funds to
// complete a bulk transfer.
var ErrInsufficientFunds = errors.New("insufficient funds to settle bulk transfer")

type Logger interface {
	Printf(format string, args ...any)
}

// Transactor is an object capable of behaving like a transaction.
type Transactor interface {
	Commit() error
	Rollback() error
}

// AtomicTransferRepository specifies the methods required to save BulkTransfers
// to a data store atomically.
type AtomicTransferRepository interface {
	BeginTx(ctx context.Context) (Transactor, error)
	GetBankAccountByIBAN(ctx context.Context, tx Transactor, iban string) (BankAccount, error)
	UpdateBankAccount(ctx context.Context, tx Transactor, ba BankAccount) error
	SaveCreditTransfers(ctx context.Context, tx Transactor, transfers []CreditTransfer) error
}

// TransferController provides the fields and methods required to perform bulk
// transfers.
type TransferController struct {
	repo   AtomicTransferRepository
	logger Logger
}

// NewTransferController configures and returns a Service.
func NewTransferController(repo AtomicTransferRepository) *TransferController {
	return &TransferController{repo: repo}
}

// PerformBulkTransfer exposes the underlying repo's PerformBulkTransfer method
// as a convenience.
func (tc *TransferController) PerformBulkTransfer(ctx context.Context, bt BulkTransfer) error {
	tx, err := tc.repo.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("tc.repo.BeginTx: %w", err)
	}

	defer tx.Rollback() //nolint:errcheck

	bankAccount, err := tc.repo.GetBankAccountByIBAN(ctx, tx, bt.Account.OrganizationIBAN)
	if err != nil {
		return fmt.Errorf("tc.repo.GetBankAccountByIBAN: %w", err)
	}

	bankAccount.BalanceCents -= bt.TotalCents()
	if bankAccount.BalanceCents < 0 {
		return ErrInsufficientFunds
	}

	if err := tc.repo.UpdateBankAccount(ctx, tx, bankAccount); err != nil {
		return fmt.Errorf("tc.repo.UpdateBankAccount: %w", err)
	}

	if err := tc.repo.SaveCreditTransfers(ctx, tx, bt.CreditTransfers); err != nil {
		return fmt.Errorf("tc.repo.SaveCreditTransfers: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("tx.Commit: %w", err)
	}

	return nil
}

// BulkTransfer represents a bulk transfer and its associated business logic.
type BulkTransfer struct {
	Account         BankAccount
	CreditTransfers []CreditTransfer
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
	ID int64
	// OrganizationName, BIC and IBAN may be populated by
	// incoming HTTP requests.
	OrganizationName string
	OrganizationBIC  string
	OrganizationIBAN string

	// BalanceCents is only available when retrieving a BankAccount from a
	// repository.
	BalanceCents int64
}

// CreditTransfer represents a single credit transfer involved in a bulk
// transfer and its associated business logic.
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
