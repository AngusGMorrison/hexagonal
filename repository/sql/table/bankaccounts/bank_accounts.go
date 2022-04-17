// Package bankaccounts operates on a database bank_accounts table and
// represents its rows. It is driver-agnostic for all database drivers that
// support positional arguments in queries.
package bankaccounts

import (
	"context"
	"embed"
	"fmt"

	"github.com/angusgmorrison/hexagonal/repository/sql"
	"github.com/angusgmorrison/hexagonal/service"
)

type row struct {
	ID               int64  `db:"id"`
	OrganizationName string `db:"organization_name"`
	IBAN             string `db:"iban"`
	BIC              string `db:"bic"`
	BalanceCents     int64  `db:"balance_cents"`
}

func rowFromDomain(ba service.BankAccount) row {
	return row{
		ID:               ba.ID,
		OrganizationName: ba.OrganizationName,
		IBAN:             ba.OrganizationIBAN,
		BIC:              ba.OrganizationBIC,
		BalanceCents:     ba.BalanceCents,
	}
}

func (r row) toDomain() service.BankAccount {
	return service.BankAccount{
		ID:               r.ID,
		OrganizationName: r.OrganizationName,
		OrganizationIBAN: r.IBAN,
		OrganizationBIC:  r.BIC,
		BalanceCents:     r.BalanceCents,
	}
}

//go:embed queries
var _queries embed.FS

// Count returns the number of rows in the table.
func Count(ctx context.Context, s sql.Queryer) (int64, error) {
	query, err := _queries.ReadFile("queries/count_bank_accounts.sql")
	if err != nil {
		return 0, fmt.Errorf("read queries/count_bank_accounts.sql: %w", err)
	}

	results := make([]int64, 0, 1)

	if err := s.Query(ctx, &results, string(query)); err != nil {
		return 0, fmt.Errorf("count bank_accounts: %w", err)
	}

	return results[0], nil
}

// FindByID returns the first account with the corresponding row ID.
func FindByID(ctx context.Context, q sql.Queryer, id int64) (service.BankAccount, error) {
	query, err := _queries.ReadFile("queries/find_bank_account_by_id.sql")
	if err != nil {
		return service.BankAccount{}, fmt.Errorf("read queries/find_bank_account_by_id.sql: %w", err)
	}

	results := make([]row, 0, 1)

	if err := q.Query(ctx, &results, string(query), id); err != nil {
		return service.BankAccount{}, fmt.Errorf("FindByID %q: %w", id, err)
	}

	return results[0].toDomain(), nil
}

// FindByIBAN returns the first account with the corresponding IBAN.
func FindByIBAN(ctx context.Context, q sql.Queryer, iban string) (service.BankAccount, error) {
	query, err := _queries.ReadFile("queries/find_bank_account_by_iban.sql")
	if err != nil {
		return service.BankAccount{}, fmt.Errorf("read queries/find_bank_account_by_iban.sql: %w", err)
	}

	results := make([]row, 0, 1)

	if err := q.Query(ctx, &results, string(query), iban); err != nil {
		return service.BankAccount{}, fmt.Errorf("FindByIBAN %q: %w", iban, err)
	}

	return results[0].toDomain(), nil
}

// Insert saves a new row to the table, returning the inserted account.
func Insert(
	ctx context.Context,
	qb sql.BindQueryer,
	ba service.BankAccount,
) (service.BankAccount, error) {
	query, err := _queries.ReadFile("queries/insert_bank_accounts.sql")
	if err != nil {
		return service.BankAccount{}, fmt.Errorf("read queries/insert_bank_accounts.sql: %w", err)
	}

	boundQuery, positionalArgs, err := qb.Bind(string(query), rowFromDomain(ba))
	if err != nil {
		return service.BankAccount{}, fmt.Errorf("bind queries/insert_bank_accounts.sql: %w", err)
	}

	results := make([]row, 0, 1)

	if err := qb.Query(ctx, &results, boundQuery, positionalArgs...); err != nil {
		return ba, fmt.Errorf("execute insert bank account: %w", err)
	}

	return results[0].toDomain(), nil
}

// Update updates an existing bank account, which is matched using its ID.
func Update(
	ctx context.Context,
	bq sql.BindQueryer,
	ba service.BankAccount,
) (service.BankAccount, error) {
	query, err := _queries.ReadFile("queries/update_bank_account.sql")
	if err != nil {
		return service.BankAccount{}, fmt.Errorf("read queries/update_bank_account.sql: %w", err)
	}

	boundQuery, positionalArgs, err := bq.Bind(string(query), rowFromDomain(ba))
	if err != nil {
		return service.BankAccount{}, fmt.Errorf("bind queries/update_bank_account.sql: %w", err)
	}

	results := make([]row, 0, 1)

	if err := bq.Query(ctx, &results, boundQuery, positionalArgs...); err != nil {
		return ba, fmt.Errorf("execute update query: %w", err)
	}

	return results[0].toDomain(), nil
}
