// Package bankaccountstable operates on a database bank_accounts table and
// represents its rows. It is database driver-agnostic.
package bankaccountstable

import (
	"context"
	"embed"
	"fmt"

	"github.com/angusgmorrison/hexagonal/service"
)

// row represents a row of a bank_accounts table.
type row struct {
	ID               int64
	OrganizationName string
	IBAN             string
	BIC              string
	BalanceCents     int64
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
	query, err := _queries.ReadFile("count_bank_accounts.sql")
	if err != nil {
		return 0, fmt.Errorf("read count_bank_accounts.sql: %w", err)
	}

	results := make([]int64, 0, 1)

	if err := s.Select(ctx, &results, string(query)); err != nil {
		return 0, fmt.Errorf("count bank_accounts: %w", err)
	}

	return results[0], nil
}

// FindByID returns the first account with the corresponding row ID.
func FindByID(ctx context.Context, s selector, id int64) (service.BankAccount, error) {
	query, err := _queries.ReadFile("find_bank_account_by_id.sql")
	if err != nil {
		return service.BankAccount{}, fmt.Errorf("read find_bank_account_by_id.sql: %w", err)
	}

	results := make([]row, 0, 1)

	if err := s.Select(ctx, &results, string(query), id); err != nil {
		return service.BankAccount{}, fmt.Errorf("FindByID %q: %w", id, err)
	}

	return results[0].toDomain(), nil
}

// FindByIBAN returns the first account with the corresponding IBAN.
func FindByIBAN(ctx context.Context, s selector, iban string) (service.BankAccount, error) {
	query, err := _queries.ReadFile("find_bank_account_by_iban.sql")
	if err != nil {
		return service.BankAccount{}, fmt.Errorf("read find_bank_account_by_iban.sql: %w", err)
	}

	results := make([]row, 0, 1)

	if err := s.Select(ctx, &results, string(query), iban); err != nil {
		return service.BankAccount{}, fmt.Errorf("FindByIBAN %q: %w", iban, err)
	}

	return results[0].toDomain(), nil
}

// Insert saves a new row to the table, returning the inserted account.
func Insert(
	ctx context.Context,
	s selector,
	ba service.BankAccount,
) (service.BankAccount, error) {
	query, err := _queries.ReadFile("insert_bank_accounts.sql")
	if err != nil {
		return service.BankAccount{}, fmt.Errorf("read insert_bank_accounts.sql: %w", err)
	}

	var (
		r    = rowFromDomain(ba)
		args = []any{
			r.OrganizationName, // 1
			r.IBAN,             // 2
			r.BIC,              // 3
			r.BalanceCents,     // 4
		}
		results = make([]row, 0, 1)
	)

	if err := s.Select(ctx, &results, string(query), args...); err != nil {
		return ba, fmt.Errorf("execute insert bank account: %w", err)
	}

	return results[0].toDomain(), nil
}

// Update updates an existing bank account, which is matched using its ID.
func Update(
	ctx context.Context,
	s selector,
	ba service.BankAccount,
) (service.BankAccount, error) {
	query, err := _queries.ReadFile("update_bank_account.sql")
	if err != nil {
		return service.BankAccount{}, fmt.Errorf("read update_bank_account.sql: %w", err)
	}

	var (
		r    = rowFromDomain(ba)
		args = []any{
			r.ID,               // 1
			r.OrganizationName, // 2
			r.IBAN,             // 3
			r.BIC,              // 4
			r.BalanceCents,     // 5
		}
		results = make([]row, 0, 1)
	)

	if err := s.Select(ctx, &results, string(query), args...); err != nil {
		return ba, fmt.Errorf("execute update query: %w", err)
	}

	return results[0].toDomain(), nil
}
