// Package mock provides mock implementations of rest interfaces.
package mock

import (
	"context"

	"github.com/angusgmorrison/hexagonal/internal/controller"
	"github.com/stretchr/testify/mock"
)

// TransactionController satisfies rest.TransactionController.
type TransactionController struct {
	mock.Mock
}

// BulkTransaction returns the error passed to the mock.
func (tc *TransactionController) BulkTransaction(
	ctx context.Context,
	bt controller.BulkTransaction,
) error {
	args := tc.Called(ctx, bt)

	return args.Error(0)
}
