//go:build integration || unit

package bankaccounts

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/angusgmorrison/hexagonal/repository/sql"
)

// Truncate truncates the table.
func Truncate(ctx context.Context, exec sql.Execer) error {
	query, err := _queries.ReadFile("queries/truncate_bank_accounts.sql")
	if err != nil {
		return fmt.Errorf("read queries/truncate_bank_accounts.sql: %w", err)
	}

	if err := exec.Execute(ctx, string(query)); err != nil {
		return fmt.Errorf("truncate bank_accounts: %w", err)
	}

	return nil
}
