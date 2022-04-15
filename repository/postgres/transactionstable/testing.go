//go:build integration || unit

package transactionstable

import (
	"context"
	_ "embed"
	"fmt"
)

func Truncate(ctx context.Context, exec executor) error {
	query, err := _queries.ReadFile("truncate_transactions.sql")
	if err != nil {
		return fmt.Errorf("read truncate_transactions.sql: %w", err)
	}

	if err := exec.Execute(ctx, string(query)); err != nil {
		return fmt.Errorf("truncate transactions: %w", err)
	}

	return nil
}
