// Package mock provides mock implementations of controller interfaces.
package mock

import (
	"context"

	"github.com/angusgmorrison/hexagonal/internal/controller"
	"github.com/stretchr/testify/mock"
)

// Transactor satisfies controller.Transactor.
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

// AtomicTransferRepository satisfies controller.AtomicTransferRepository.
type AtomicTransferRepository struct {
	mock.Mock
}

// BeginSerializableTx returns the controller.Transactor and error passed to the
// mock.
func (atr *AtomicTransferRepository) BeginSerializableTx(
	ctx context.Context,
) (controller.Transactor, error) {
	args := atr.Called(ctx)

	tx := args.Get(0).(*Transactor)

	return tx, args.Error(1)
}

// GetBankAccountByIBAN returns the controller.BankAccount and error passed to
// the mock.
func (atr *AtomicTransferRepository) GetBankAccountByIBAN(
	ctx context.Context,
	tx controller.Transactor,
	iban string,
) (controller.BankAccount, error) {
	args := atr.Called(ctx, tx, iban)

	bankAccount := args.Get(0).(controller.BankAccount)

	return bankAccount, args.Error(1)
}

// UpdateBankAccount returns the error passed to the mock.
func (atr *AtomicTransferRepository) UpdateBankAccount(
	ctx context.Context,
	tx controller.Transactor,
	ba controller.BankAccount,
) error {
	args := atr.Called(ctx, tx, ba)

	return args.Error(0)
}

// SaveCreditTransfers returns the error passed to the mock.
func (atr *AtomicTransferRepository) SaveCreditTransfers(
	ctx context.Context,
	tx controller.Transactor,
	transfers controller.Transactions,
) error {
	args := atr.Called(ctx, tx, transfers)

	return args.Error(0)
}
