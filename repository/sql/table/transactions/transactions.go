// Package transactions operates on a database transactions table and represents
// its rows. It is driver-agnostic for all database drivers that support
// positional arguments in queries.
package transactions

import (
	"context"
	"embed"
	"fmt"

	"github.com/angusgmorrison/hexagonal/repository/sql"
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
	ID               int64  `db:"id"`
	BankAccountID    int64  `db:"bank_account_id"`
	CounterpartyName string `db:"counterparty_name"`
	CounterpartyIBAN string `db:"counterparty_iban"`
	CounterpartyBIC  string `db:"counterparty_bic"`
	AmountCents      int64  `db:"amount_cents"`
	AmountCurrency   string `db:"amount_currency"`
	Description      string `db:"description"`
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

//go:embed queries
var _queries embed.FS

// Count returns the number of rows in the table.
func Count(ctx context.Context, q sql.Queryer) (int64, error) {
	query, err := _queries.ReadFile("queries/count_transactions.sql")
	if err != nil {
		return 0, fmt.Errorf("read queries/count_transactions.sql: %w", err)
	}

	results := make([]int64, 0, 1)

	if err := q.Query(ctx, &results, string(query)); err != nil {
		return 0, fmt.Errorf("count transactions: %w", err)
	}

	return results[0], nil
}

// SelectByCounterpartyName returns all transactions with the given counterpary
// name.
func SelectByCounterpartyName(
	ctx context.Context,
	q sql.Queryer,
	name string,
) (service.Transactions, error) {
	query, err := _queries.ReadFile("queries/select_transactions_by_counterparty_name.sql")
	if err != nil {
		return nil, fmt.Errorf("read queries/select_transactions_by_counterparty_name.sql: %w", err)
	}

	var results rows

	if err := q.Query(ctx, &results, string(query), name); err != nil {
		return nil, fmt.Errorf("execute select by counterparty name: %w", err)
	}

	return results.toDomain(), nil
}

// Insert bulk saves new rows to the table, returning the inserted rows.
func Insert(
	ctx context.Context,
	bq sql.BindQueryer,
	transaction service.Transaction,
) (service.Transaction, error) {
	args := rows{rowFromDomain(transaction)}

	results, err := insert(ctx, bq, args)
	if err != nil {
		return transaction, nil
	}

	return results[0].toDomain(), nil
}

// BulkInsert saves new rows to the table, returning the inserted rows.
func BulkInsert(
	ctx context.Context,
	bq sql.BindQueryer,
	transactions service.Transactions,
) (service.Transactions, error) {
	args := rowsFromDomain(transactions)

	results, err := insert(ctx, bq, args)
	if err != nil {
		return transactions, err
	}

	return results.toDomain(), nil
}

func insert(
	ctx context.Context,
	bq sql.BindQueryer,
	rs rows,
) (rows, error) {
	query, err := _queries.ReadFile("queries/insert_transactions.sql")
	if err != nil {
		return nil, fmt.Errorf("read queries/insert_transactions.sql: %w", err)
	}

	boundQuery, positionalArgs, err := bq.Bind(string(query), rs)
	if err != nil {
		return nil, fmt.Errorf("bind queries/insert_transactions.sql: %w", err)
	}

	results := make(rows, 0, len(rs))

	if err := bq.Query(ctx, &results, boundQuery, positionalArgs...); err != nil {
		return nil, fmt.Errorf("execute insert transactions: %w", err)
	}

	return results, nil
}
