package postgres

import (
	"context"
	"fmt"

	"github.com/angusgmorrison/hexagonal/repository/postgres/bankaccountstable"
	"github.com/angusgmorrison/hexagonal/service"
)

type Serializer interface {
	BeginSerializableTx(ctx context.Context) (ExecSelector, error)
}

// executor executes a query that returns no result.
type Executor interface {
	Execute(ctx context.Context, query string, args ...any) error
}

// selector executes a query that populates dest with the returned rows. Since
// the number of rows is unknown, dest must be a pointer to a slice.
type Selector interface {
	Select(ctx context.Context, dest any, query string, args ...any) error
}

type ExecSelector interface {
	Executor
	Selector
}

type database interface {
	Serializer
	ExecSelector
}

type BulkTransactionRepository struct {
	db database
}

func NewBulkTransferRepository(db database) *BulkTransactionRepository {
	return &BulkTransactionRepository{db: db}
}

func (btr *BulkTransactionRepository) BeginSerializableTx(
	ctx context.Context,
) (service.Transactor, error) {
	return btr.db.BeginSerializableTx(ctx)
}

func (btr *BulkTransactionRepository) LoadBankAccountFromIBAN(
	ctx context.Context,
	transactor service.Transactor,
	bt service.BulkTransaction,
) (service.BulkTransaction, error) {
	tx, ok := transactor.(ExecSelector)
	if !ok {
		return bt, TxTypeError{tx: transactor}
	}

	bankAccount, err := bankaccountstable.FindByIBAN(ctx, tx, bt.Account.OrganizationIBAN)
	if err != nil {
		return bt, fmt.Errorf("LoadBankAccountFromIBAN: %w", err)
	}

	bt.Account = bankAccount

	return bt, nil
}
