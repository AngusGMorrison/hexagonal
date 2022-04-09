package controller

import (
	"context"
	"errors"
)

// TransferRepository specifies the methods required to save BulkTransfers to a
// data store.
type TransferRepository interface {
	PerformBulkTransfer(context.Context, BulkTransfer, BulkTransferValidator) error
}

// TransferController provides the fields and methods required to perform bulk
// transfers.
type TransferController struct {
	repo TransferRepository
}

// NewTransferController configures and returns a Service.
func NewTransferController(repo TransferRepository) *TransferController {
	return &TransferController{repo: repo}
}

// PerformBulkTransfer exposes the underlying repo's PerformBulkTransfer method
// as a convenience.
func (tc *TransferController) PerformBulkTransfer(
	ctx context.Context, bt BulkTransfer, validate BulkTransferValidator,
) error {
	return tc.repo.PerformBulkTransfer(ctx, bt, validate)
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

// ErrInsufficientFunds signals that a bank account did not have enough funds to
// complete a bulk transfer.
var ErrInsufficientFunds = errors.New("insufficient funds to settle bulk transfer")

// BulkTransferValidator represents a function that returns an error if the
// BulkTransfer is in an invalid state.
type BulkTransferValidator func(BulkTransfer) error

// ValidateBulkTransfer is a BuklTransferValidator that calls all validations
// related to Bulktransfer
func ValidateBulkTransfer(bt BulkTransfer) error {
	return ValidatePositiveBankBalance(bt)
}

// ValidatePositiveBankBalance asserts that the BulkTransfer's associated
// BankAccount remains in credit after the CreditTransfers have been applied.
func ValidatePositiveBankBalance(bt BulkTransfer) error {
	if bt.Account.BalanceCents < 0 {
		return ErrInsufficientFunds
	}

	return nil
}
