// Package mock provides mock implementations of transferdomain interfaces.
package mock

import (
	"context"

	"github.com/angusgmorrison/hexagonal/internal/app/transferdomain"
)

// Repository satisfies transferdomain.Repository, and allows the user to set
// the error to be returned from PerformBulkTransfer.
type Repository struct {
	Err error
}

// PerformBulkTransfer is a mock implementation of
// transferdomain.Repository.PerformBulkTransfer. It returns the error on r.
func (r *Repository) PerformBulkTransfer(
	ctx context.Context,
	bt transferdomain.BulkTransfer,
	validate transferdomain.BulkTransferValidator,
) error {
	return r.Err
}
