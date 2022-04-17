// Package mock provides mock implementations of service interfaces.
package mock

import (
	"context"

	"github.com/angusgmorrison/hexagonal/service"
	"github.com/stretchr/testify/mock"
)

// BulkTransactionService satisfies service.BulkTransactionService.
type BulkTransactionService struct {
	mock.Mock
}

// BulkTransaction returns the error passed to the mock.
func (bts *BulkTransactionService) BulkTransaction(ctx context.Context, bt service.BulkTransaction) error {
	args := bts.Called(ctx, bt)

	return args.Error(0)
}

// BulkTransactionRepository satisfies service.BulkTransactionRepository.
type BulkTransactionRepository struct {
	mock.Mock
}

// GetBankAccountByIBAN returns the service.BulkTransaction and error passed to
// the mock.
func (btr *BulkTransactionRepository) GetBankAccountByIBAN(
	ctx context.Context,
	bt service.BulkTransaction,
) (service.BulkTransaction, error) {
	args := btr.Called(ctx, bt)

	bulkTransaction := args.Get(0).(service.BulkTransaction)

	return bulkTransaction, args.Error(1)
}

// Save returns the service.BulkTransaction and error passed to the mock.
func (btr *BulkTransactionRepository) Save(
	ctx context.Context,
	bt service.BulkTransaction,
) (service.BulkTransaction, error) {
	args := btr.Called(ctx, bt)

	bulkTransaction := args.Get(0).(service.BulkTransaction)

	return bulkTransaction, args.Error(1)
}

// Abort returns the error passed to the mock.
func (btr *BulkTransactionRepository) Abort() error {
	args := btr.Called()

	return args.Error(0)
}
