package postgres

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/angusgmorrison/hexagonal/internal/adapter/envconfig"
	"github.com/angusgmorrison/hexagonal/internal/controller"
	"github.com/jmoiron/sqlx"
)

// Transactions is a convenience wrapper around one or more instances of
// Transaction.
type Transactions []Transaction

func transactionsFromDomain(ct []controller.Transaction) Transactions {
	transactions := make(Transactions, 0, len(ct))

	for _, t := range ct {
		transactions = append(transactions, transactionFromDomain(t))
	}

	return transactions
}

func (ts Transactions) toDomain() controller.Transactions {
	cts := make(controller.Transactions, 0, len(ts))

	for _, t := range ts {
		cts = append(cts, t.toDomain())
	}

	return cts
}

// Transaction represents a row of the transactions table.
type Transaction struct {
	ID               int64  `db:"id"`
	BankAccountID    int64  `db:"bank_account_id"`
	CounterpartyName string `db:"counterparty_name"`
	CounterpartyIBAN string `db:"counterparty_iban"`
	CounterpartyBIC  string `db:"counterparty_bic"`
	AmountCents      int64  `db:"amount_cents"`
	AmountCurrency   string `db:"amount_currency"`
	Description      string `db:"description"`
}

func transactionFromDomain(ct controller.Transaction) Transaction {
	return Transaction{
		ID:               ct.ID,
		BankAccountID:    ct.BankAccountID,
		CounterpartyName: ct.CounterpartyName,
		CounterpartyBIC:  ct.CounterpartyBIC,
		CounterpartyIBAN: ct.CounterpartyIBAN,
		AmountCents:      ct.AmountCents,
		AmountCurrency:   ct.Currency,
		Description:      ct.Description,
	}
}

func (t Transaction) toDomain() controller.Transaction {
	return controller.Transaction{
		ID:               t.ID,
		BankAccountID:    t.BankAccountID,
		CounterpartyName: t.CounterpartyName,
		CounterpartyBIC:  t.CounterpartyBIC,
		CounterpartyIBAN: t.CounterpartyIBAN,
		AmountCents:      t.AmountCents,
		Currency:         t.AmountCurrency,
		Description:      t.Description,
	}
}

// TransactionRepository provides the methods necessary to perform transfers.
// Satisfies controller.AtomicTransferRepository.
type TransactionRepository struct {
	db        *DB
	appConfig envconfig.App
	queries   Queries
}

// Statically verify controller interface compliance.
var _ controller.AtomicTransactionRepository = (*TransactionRepository)(nil)

const (
	_countTransactions                    QueryFilename = "count_transactions.sql"
	_insertTransactions                   QueryFilename = "insert_transactions.sql"
	_selectTransactionsByCounterpartyName QueryFilename = "select_transactions_by_counterparty_name.sql"
	_truncateTransactions                 QueryFilename = "truncate_transactions.sql"
)

// NewTransactionRepository returns a new BankAccountRepository with its SQL
// queries preloaded.
func NewTransactionRepository(db *DB, appConfig envconfig.App) (*TransactionRepository, error) {
	repo := TransactionRepository{
		db:        db,
		appConfig: appConfig,
	}

	if err := repo.loadQueries(); err != nil {
		return nil, err
	}

	return &repo, nil
}

// BulkInsertTx bulk inserts transactions atomically.
func (tr *TransactionRepository) BulkInsertTx(
	ctx context.Context,
	transactor controller.Transactor,
	transactions controller.Transactions,
) error {
	tx, ok := transactor.(*sqlx.Tx)
	if !ok {
		return TxTypeError{tx: tx}
	}

	rows := transactionsFromDomain(transactions)

	if _, err := tx.NamedExecContext(
		ctx,
		tr.queries[_insertTransactions],
		rows,
	); err != nil {
		return fmt.Errorf("insert transactions: %w", err)
	}

	return nil
}

// Count returns the number of rows in the transactions table.
func (tr *TransactionRepository) Count(ctx context.Context) (int64, error) {
	var count int64

	if err := tr.db.Get(ctx, &count, tr.queries[_countTransactions]); err != nil {
		return 0, fmt.Errorf("count transactions: %w", err)
	}

	return count, nil
}

// SelectByCounterpartyName returns all transactions with the given counterpary
// name.
func (tr *TransactionRepository) SelectByCounterpartyName(
	ctx context.Context,
	name string,
) (controller.Transactions, error) {
	var rows Transactions

	if err := tr.db.Select(ctx, &rows, tr.queries[_selectTransactionsByCounterpartyName], name); err != nil {
		return nil, fmt.Errorf("select transactions by counterparty name: %w", err)
	}

	return rows.toDomain(), nil
}

// Truncate truncates the transactions table.
func (tr *TransactionRepository) Truncate(ctx context.Context) error {
	if !truncationPermitted(tr.appConfig.Env) {
		return UnpermittedTruncationError{tr.appConfig.Env}
	}

	if _, err := tr.db.Exec(ctx, tr.queries[_truncateTransactions]); err != nil {
		return fmt.Errorf("truncateTransactions: %w", err)
	}

	return nil
}

func (tr *TransactionRepository) loadQueries() error {
	queryDir := filepath.Join(tr.appConfig.Root, RelativeQueryDir(), "transaction")
	queryFilenames := []QueryFilename{
		_countTransactions,
		_insertTransactions,
		_selectTransactionsByCounterpartyName,
		_truncateTransactions,
	}

	if tr.queries == nil {
		tr.queries = make(Queries, len(queryFilenames))
	}

	if err := tr.queries.Load(queryDir, queryFilenames); err != nil {
		return fmt.Errorf("load transaction queries: %w", err)
	}

	return nil
}
