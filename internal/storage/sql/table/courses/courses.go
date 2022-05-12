// Package courses operates on a database courses table and represents its rows.
// It is driver-agnostic.
package courses

import (
	"context"
	"embed"
	"fmt"

	"github.com/angusgmorrison/hexagonal/internal/storage/sql"
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

	if len(results) == 0 {
		return Row{}, CourseNotFoundError{Code: code}
	}

	return results[0], nil
}

// Insert inserts the given courses into the table.
func Insert(ctx context.Context, bq sql.BindQueryer, courses []Row) ([]Row, error) {
	query, err := _queries.ReadFile("queries/insert_courses.sql")
	if err != nil {
		return nil, fmt.Errorf("read queries/insert_courses.sql: %w", err)
	}

	boundQuery, positionalArgs, err := bq.Bind(string(query), courses)
	if err != nil {
		return nil, fmt.Errorf("bind queries/insert_courses.sql: %w", err)
	}

	results := make([]Row, 0, len(courses))

	if err := bq.Query(ctx, &results, boundQuery, positionalArgs...); err != nil {
		return nil, fmt.Errorf("Insert: %w", err)
	}

	return results, nil
}

// CourseNotFoundError is returned when searching for a course by code returns
// no results.
type CourseNotFoundError struct {
	Code string
}

func (cnfe CourseNotFoundError) Error() string {
	return fmt.Sprintf("no course with code %q", cnfe.Code)
}
