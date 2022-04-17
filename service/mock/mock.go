// Package mock provides mock implementations of service interfaces.
package mock

import (
	"context"

	"github.com/angusgmorrison/hexagonal/service"
	"github.com/stretchr/testify/mock"
)

// BulkTransactionService satisfies service.BulkTransactionService.
type BulkTransactionService struct {
	mock mock.Mock
}

// GetBankAccountByIBAN returns the service.BulkTransaction and error passed to
// the mock.
func (bts *BulkTransactionService) GetBankAccountByIBAN(
	ctx context.Context,
	bt service.BulkTransaction,
) (service.BulkTransaction, error) {
	args := bts.mock.Called(ctx, bt)

	bulkTransaction := args.Get(0).(service.BulkTransaction)

	return bulkTransaction, args.Error(1)
}

// Save returns the service.BulkTransaction and error passed to the mock.
func (bts *BulkTransactionService) Save(
	ctx context.Context,
	bt service.BulkTransaction,
) (service.BulkTransaction, error) {
	args := bts.mock.Called(ctx, bt)

	bulkTransaction := args.Get(0).(service.BulkTransaction)

	return bulkTransaction, args.Error(1)
}

// Abort returns the error passed to the mock.
func (bts *BulkTransactionService) Abort() error {
	args := bts.mock.Called()

	return args.Error(0)
}
