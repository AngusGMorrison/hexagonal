// Package students operates on a database students table and represents its
// rows. It is driver-agnostic.
package students

import (
	"context"
	"embed"
	"fmt"
	"time"

	"github.com/angusgmorrison/hexagonal/internal/primitive"
	"github.com/angusgmorrison/hexagonal/internal/storage/sql"
	"github.com/jmoiron/sqlx"
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
	rq sql.RebindQueryer,
	emails []primitive.EmailAddress,
) ([]Row, error) {
	query, err := _queries.ReadFile("queries/select_students_by_email.sql")
	if err != nil {
		return nil, fmt.Errorf("read queries/select_students_by_email.sql: %w", err)
	}

	inQuery, positionalArgs, err := sqlx.In(string(query), emails)
	if err != nil {
		return nil, fmt.Errorf("generate IN query with emails: %w", err)
	}

	boundQuery := rq.Rebind(inQuery)

	results := make([]Row, 0, len(emails))

	if err := rq.Query(ctx, &results, boundQuery, positionalArgs...); err != nil {
		return nil, fmt.Errorf("SelectByEmail(%+v): %w", emails, err)
	}

	return results, nil
}

// Insert inserts the given students into the table.
func Insert(ctx context.Context, bq sql.BindQueryer, students []Row) ([]Row, error) {
	query, err := _queries.ReadFile("queries/insert_students.sql")
	if err != nil {
		return nil, fmt.Errorf("read queries/insert_students.sql: %w", err)
	}

	boundQuery, positionalArgs, err := bq.Bind(string(query), students)
	if err != nil {
		return nil, fmt.Errorf("bind queries/insert_students.sql: %w", err)
	}

	results := make([]Row, 0, len(students))

	if err := bq.Query(ctx, &results, boundQuery, positionalArgs...); err != nil {
		return nil, fmt.Errorf("Insert: %w", err)
	}

	return results, nil
}
