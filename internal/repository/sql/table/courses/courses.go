// Package courses operates on a database courses table and represents its rows.
// It is driver-agnostic.
package courses

import (
	"context"
	"embed"
	"fmt"

	"github.com/angusgmorrison/hexagonal/internal/repository/sql"
)

//go:embed queries
var _queries embed.FS

// Row represents a row of the courses table.
type Row struct {
	ID          int64  `db:"id"`
	Code        string `db:"code"`
	Title       string `db:"title"`
	Capacity    uint32 `db:"capacity"`
	Description string `db:"description"`
}

// FindByCode returns a row based on its course code.
func FindByCode(ctx context.Context, q sql.Queryer, code string) (Row, error) {
	query, err := _queries.ReadFile("queries/find_course_by_code.sql")
	if err != nil {
		return Row{}, fmt.Errorf("read queries/find_course_by_code.sql: %w", err)
	}

	results := make([]Row, 0, 1)

	if err := q.Query(ctx, &results, string(query), code); err != nil {
		return Row{}, fmt.Errorf("FindByCode(%q): %w", code, err)
	}

	return results[0], nil
}
