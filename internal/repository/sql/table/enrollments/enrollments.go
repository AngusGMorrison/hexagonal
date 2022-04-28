// Package enrollments operates on a database courses table. It is
// driver-agnostic.
package enrollments

import (
	"context"
	"embed"
	"fmt"

	"github.com/angusgmorrison/hexagonal/internal/repository/sql"
)

// Row represents a row of the enrollments table.
type Row struct {
	ID        int64 `db:"id"`
	CourseID  int64 `db:"course_id"`
	StudentID int64 `db:"student_id"`
}

//go:embed queries
var _queries embed.FS

// Insert inserts the given rows into the enrollments table.
func Insert(ctx context.Context, bq sql.BindQueryer, rows []Row) ([]Row, error) {
	query, err := _queries.ReadFile("queries/insert_enrollments.sql")
	if err != nil {
		return nil, fmt.Errorf("read queries/insert_enrollments.sql: %w", err)
	}

	boundQuery, positionalArgs, err := bq.Bind(string(query), rows)
	if err != nil {
		return nil, fmt.Errorf("bind queries/insert_enrollments.sql: %w", err)
	}

	results := make([]Row, 0, len(rows))

	if err := bq.Query(ctx, &results, boundQuery, positionalArgs...); err != nil {
		return nil, fmt.Errorf("Insert: %w", err)
	}

	return results, nil
}
