// Package mock provides mock implementations of controller interfaces.
package mock

import (
	"context"

	"github.com/angusgmorrison/hexagonal/internal/controller"
)

// Repository satisfies controller.Repository, and allows the user to set
// the error to be returned from PerformBulkTransfer.
type Repository struct {
	Err error
}

// PerformBulkTransfer is a mock implementation of
// controller.Repository.PerformBulkTransfer. It returns the error on r.
func (r *Repository) PerformBulkTransfer(
	ctx context.Context,
	bt controller.BulkTransfer,
	validate controller.BulkTransferValidator,
) error {
	return r.Err
}
