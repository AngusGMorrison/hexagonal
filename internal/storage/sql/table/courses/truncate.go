//go:build integration || unit

package courses

import (
	"context"
	"fmt"

	"github.com/angusgmorrison/hexagonal/internal/storage/sql"
)

func Truncate(ctx context.Context, exec sql.Execer) error {
	query, err := _queries.ReadFile("queries/truncate_courses.sql")
	if err != nil {
		return fmt.Errorf("Truncate: %w", err)
	}

	if err := exec.Execute(ctx, string(query)); err != nil {
		return fmt.Errorf("Truncate: %w", err)
	}

	return nil
}
