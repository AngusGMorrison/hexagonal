// Package mock provides mock implementations of rest interfaces.
package mock

import (
	"context"

	"github.com/angusgmorrison/hexagonal/internal/service"
	"github.com/stretchr/testify/mock"
)

// TransactionService satisfies rest.TransactionService.
type TransactionService struct {
	mock.Mock
}

// BulkTransaction returns the error passed to the mock.
func (tc *TransactionService) BulkTransaction(
	ctx context.Context,
	bt service.BulkTransaction,
) error {
	args := tc.Called(ctx, bt)

	return args.Error(0)
}
