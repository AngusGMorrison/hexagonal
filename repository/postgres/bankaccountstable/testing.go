//go:build integration || unit

package bankaccountstable

import (
	"context"
	_ "embed"
	"fmt"
)

func Truncate(ctx context.Context, exec executor) error {
	query, err := _queries.ReadFile("truncate_bank_accounts.sql")
	if err != nil {
		return fmt.Errorf("read truncate_bank_accounts.sql: %w", err)
	}

	if err := exec.Execute(ctx, string(query)); err != nil {
		return fmt.Errorf("truncate bank_accounts: %w", err)
	}

	return nil
}
