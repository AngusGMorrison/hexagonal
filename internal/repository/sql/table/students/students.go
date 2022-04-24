// Package students operates on a database students table and represents its
// rows. It is driver-agnostic.
package students

import (
	"context"
	"embed"
	"fmt"
	"time"

	"github.com/angusgmorrison/hexagonal/internal/primitive"
	"github.com/angusgmorrison/hexagonal/internal/repository/sql"
)

//go:embed queries
var _queries embed.FS

// Row represents a row of the students table.
type Row struct {
	ID        int64                  `db:"id"`
	Name      string                 `db:"name"`
	Birthdate time.Time              `db:"birthdate"`
	Email     primitive.EmailAddress `db:"email"`
}

// OnCourse returns the rows of all students enrolled in the course with the
// given ID.
func OnCourse(ctx context.Context, q sql.Queryer, courseID int64) ([]Row, error) {
	query, err := _queries.ReadFile("queries/select_students_on_course.sql")
	if err != nil {
		return nil, fmt.Errorf("read queries/select_students_on_course.sql: %w", err)
	}

	var results []Row

	if err := q.Query(ctx, &results, string(query), courseID); err != nil {
		return nil, fmt.Errorf("OnCourse(%d): %w", courseID, err)
	}

	return results, nil
}

// SelectByEmail returns all students whose email addresses are present in the
// given slice.
func SelectByEmail(
	ctx context.Context,
	q sql.Queryer,
	emails []primitive.EmailAddress,
) ([]Row, error) {
	query, err := _queries.ReadFile("queries/select_students_by_email.sql")
	if err != nil {
		return nil, fmt.Errorf("read queries/select_students_on_course.sql: %w", err)
	}

	results := make([]Row, 0, len(emails))

	inArg, err := sql.SliceToInArg(emails)
	if err != nil {
		return nil, fmt.Errorf("SelectByEmail: %w", err)
	}

	if err := q.Query(ctx, &results, string(query), inArg); err != nil {
		return nil, fmt.Errorf("SelectByEmail(%q): %w", inArg, err)
	}

	return results, nil
}
