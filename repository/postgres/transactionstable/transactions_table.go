// Package transactionstable operates on a database transactions table and
// represents its rows. It is database driver-agnostic.
package transactionstable

import (
	"context"
	"embed"
	"fmt"

	"github.com/angusgmorrison/hexagonal/service"
)

type rows []row

func rowsFromDomain(transactions []service.Transaction) rows {
	transactionRows := make(rows, 0, len(transactions))

	for _, transaction := range transactions {
		transactionRows = append(transactionRows, rowFromDomain(transaction))
	}

	return transactionRows
}

func (rs rows) toDomain() service.Transactions {
	transactions := make(service.Transactions, 0, len(rs))

	for _, r := range rs {
		transactions = append(transactions, r.toDomain())
	}

	return transactions
}

type row struct {
	ID               int64
	BankAccountID    int64
	CounterpartyName string
	CounterpartyIBAN string
	CounterpartyBIC  string
	AmountCents      int64
	AmountCurrency   string
	Description      string
}

func rowFromDomain(t service.Transaction) row {
	return row{
		ID:               t.ID,
		BankAccountID:    t.BankAccountID,
		CounterpartyName: t.CounterpartyName,
		CounterpartyBIC:  t.CounterpartyBIC,
		CounterpartyIBAN: t.CounterpartyIBAN,
		AmountCents:      t.AmountCents,
		AmountCurrency:   t.Currency,
		Description:      t.Description,
	}
}

func (r row) toDomain() service.Transaction {
	return service.Transaction{
		ID:               r.ID,
		BankAccountID:    r.BankAccountID,
		CounterpartyName: r.CounterpartyName,
		CounterpartyBIC:  r.CounterpartyBIC,
		CounterpartyIBAN: r.CounterpartyIBAN,
		AmountCents:      r.AmountCents,
		Currency:         r.AmountCurrency,
		Description:      r.Description,
	}
}

// executor executes a query that returns no result.
type executor interface {
	Execute(ctx context.Context, query string, args ...any) error
}

// selector executes a query that populates dest with the returned rows. Since
// the number of rows is unknown, dest must be a pointer to a slice.
type selector interface {
	Select(ctx context.Context, dest any, query string, args ...any) error
}

//go:embed queries
var _queries embed.FS

// Count returns the number of rows in the table.
func Count(ctx context.Context, s selector) (int64, error) {
	query, err := _queries.ReadFile("count_transactions.sql")
	if err != nil {
		return 0, fmt.Errorf("read count_transactions.sql: %w", err)
	}

	results := make([]int64, 0, 1)

	if err := s.Select(ctx, &results, string(query)); err != nil {
		return 0, fmt.Errorf("count transactions: %w", err)
	}

	return results[0], nil
}

// SelectByCounterpartyName returns all transactions with the given counterpary
// name.
func SelectByCounterpartyName(
	ctx context.Context,
	s selector,
	name string,
) (service.Transactions, error) {
	query, err := _queries.ReadFile("select_transactions_by_counterparty_name.sql")
	if err != nil {
		return nil, fmt.Errorf("read select_transactions_by_counterparty_name.sql: %w", err)
	}

	var results rows

	if err := s.Select(ctx, &results, string(query), name); err != nil {
		return nil, fmt.Errorf("execute select by counterparty name: %w", err)
	}

	return results.toDomain(), nil
}

// Insert bulk saves new rows to the table, returning the inserted rows.
func Insert(
	ctx context.Context,
	s selector,
	transaction service.Transaction,
) (service.Transaction, error) {
	args := rows{rowFromDomain(transaction)}

	results, err := insert(ctx, s, args)
	if err != nil {
		return transaction, nil
	}

	return results[0].toDomain(), nil
}

// BulkInsert saves new rows to the table, returning the inserted rows.
func BulkInsert(
	ctx context.Context,
	s selector,
	transactions service.Transactions,
) (service.Transactions, error) {
	args := rowsFromDomain(transactions)

	results, err := insert(ctx, s, args)
	if err != nil {
		return transactions, err
	}

	return results.toDomain(), nil
}

func insert(
	ctx context.Context,
	s selector,
	rs rows,
) (rows, error) {
	query, err := _queries.ReadFile("insert_transactions.sql")
	if err != nil {
		return nil, fmt.Errorf("read insert_transactions.sql: %w", err)
	}

	var (
		args    = make([]any, 0, len(rs))
		results = make(rows, 0, len(rs))
	)

	for _, r := range rs {
		arg := []any{
			r.CounterpartyName, // 1
			r.CounterpartyIBAN, // 2
			r.CounterpartyBIC,  // 3
			r.AmountCents,      // 4
			r.AmountCurrency,   // 5
			r.BankAccountID,    // 6
			r.Description,      // 7
		}
		args = append(args, arg)
	}

	if err := s.Select(ctx, &results, string(query), args...); err != nil {
		return nil, fmt.Errorf("execute insert transactions: %w", err)
	}

	return results, nil
}
