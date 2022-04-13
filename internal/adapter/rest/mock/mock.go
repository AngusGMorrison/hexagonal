// Package mock provides mock implementations of rest interfaces.
package mock

import (
	"context"

	"github.com/angusgmorrison/hexagonal/internal/controller"
	"github.com/stretchr/testify/mock"
)

// TransferController satisfies rest.TransferController.
type TransferController struct {
	mock.Mock
}

// PerformBulkTransfer returns the error passed to the mock.
func (tc *TransferController) PerformBulkTransfer(
	ctx context.Context,
	bt controller.BulkTransfer,
) error {
	args := tc.Called(ctx, bt)

	return args.Error(0)
}
