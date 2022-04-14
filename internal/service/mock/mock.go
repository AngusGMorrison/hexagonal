// Package mock provides mock implementations of service interfaces.
package mock

import (
	"context"

	"github.com/angusgmorrison/hexagonal/internal/service"
	"github.com/stretchr/testify/mock"
)

// Transactor satisfies service.Transactor.
type Transactor struct {
	mock.Mock
}

// Commit returns the error passed to the mock.
func (t *Transactor) Commit() error {
	args := t.Called()

	return args.Error(0)
}

// Rollback returns the errors passed to the mock.
func (t *Transactor) Rollback() error {
	args := t.Called()

	return args.Error(0)
}

// AtomicTransferRepository satisfies service.AtomicTransferRepository.
type AtomicTransferRepository struct {
	mock.Mock
}

// BeginSerializableTx returns the service.Transactor and error passed to the
// mock.
func (atr *AtomicTransferRepository) BeginSerializableTx(
	ctx context.Context,
) (service.Transactor, error) {
	args := atr.Called(ctx)

	tx := args.Get(0).(*Transactor)

	return tx, args.Error(1)
}

// GetBankAccountByIBAN returns the service.BankAccount and error passed to
// the mock.
func (atr *AtomicTransferRepository) GetBankAccountByIBAN(
	ctx context.Context,
	tx service.Transactor,
	iban string,
) (service.BankAccount, error) {
	args := atr.Called(ctx, tx, iban)

	bankAccount := args.Get(0).(service.BankAccount)

	return bankAccount, args.Error(1)
}

// UpdateBankAccount returns the error passed to the mock.
func (atr *AtomicTransferRepository) UpdateBankAccount(
	ctx context.Context,
	tx service.Transactor,
	ba service.BankAccount,
) error {
	args := atr.Called(ctx, tx, ba)

	return args.Error(0)
}

// SaveCreditTransfers returns the error passed to the mock.
func (atr *AtomicTransferRepository) SaveCreditTransfers(
	ctx context.Context,
	tx service.Transactor,
	transfers service.Transactions,
) error {
	args := atr.Called(ctx, tx, transfers)

	return args.Error(0)
}
