//go:build integration || unit

package transactions

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/angusgmorrison/hexagonal/repository/sql"
)

// Truncate truncates the table.
func Truncate(ctx context.Context, exec sql.Execer) error {
	query, err := _queries.ReadFile("queries/truncate_transactions.sql")
	if err != nil {
		return fmt.Errorf("read queries/truncate_transactions.sql: %w", err)
	}

	if err := exec.Execute(ctx, string(query)); err != nil {
		return fmt.Errorf("truncate transactions: %w", err)
	}

	return nil
}
