package scribe

import (
	"context"
	"fmt"

	"github.com/angusgmorrison/hexagonal/internal/primitive"
	"github.com/angusgmorrison/hexagonal/internal/repository/sql"
	"github.com/angusgmorrison/hexagonal/internal/repository/sql/table/courses"
	"github.com/angusgmorrison/hexagonal/internal/repository/sql/table/enrollments"
	"github.com/angusgmorrison/hexagonal/internal/repository/sql/table/students"
	"github.com/angusgmorrison/hexagonal/internal/service/classservice"
)

// AtomicClassScribe implements classservice.AtomicClassScribe, providing a
// translation layer between database tables and business domain that performs
// its operations atomically.
type AtomicClassScribe struct {
	*atomicScribe
}

// NewAtomicClassScribeFactory returns a scribe factory function that has
// captured a reference to a database. This makes it trivial for business logic
// to instantiate a new scribe for a single database transaction,
// avoiding the need for complex thread safety measures.
//
// Scribes are single-use and short-lived by design. Each scribe returned by the
// factory contains an active transaction against which the scribe's database
// operations are run, so long-lived scribes will leak database connections.
//
// Attempts to reuse the scribe after the transaction is committed or rolled
// back return errors.
func NewAtomicClassScribeFactory(db sql.Database) classservice.AtomicClassScribeFactory {
	return func() classservice.AtomicClassScribe {
		return &AtomicClassScribe{
			atomicScribe: &atomicScribe{db: db},
		}
	}
}

// GetClass returns a course and all its enrolled students from the class code
// provided.
func (acs *AtomicClassScribe) GetClassByCourseCode(
	ctx context.Context,
	courseCode string,
) (classservice.Class, error) {
	tx, err := acs.getTx()
	if err != nil {
		return classservice.Class{}, fmt.Errorf("GetClassByCourseCode(%q): %w", courseCode, err)
	}

	courseRow, err := courses.FindByCode(ctx, tx, courseCode)
	if err != nil {
		return classservice.Class{}, fmt.Errorf("GetClassByCourseCode(%q): %w", courseCode, err)
	}

	studentRows, err := students.OnCourse(ctx, tx, courseRow.ID)
	if err != nil {
		return classservice.Class{}, fmt.Errorf("GetClassByCourseCode(%q): %w", courseCode, err)
	}

	return classFromRows(courseRow, studentRows), nil
}

func (acs *AtomicClassScribe) GetStudentsByEmail(
	ctx context.Context,
	emails []primitive.EmailAddress,
) (classservice.Students, error) {
	tx, err := acs.getTx()
	if err != nil {
		return nil, fmt.Errorf("GetStudentsByEmail(%v): %w", emails, err)
	}

	studentRows, err := students.SelectByEmail(ctx, tx, emails)
	if err != nil {
		return nil, fmt.Errorf("GetStudentsByEmail(%v): %w", emails, err)
	}

	return studentsFromRows(studentRows), nil
}

// EnrollStudents enrolls the given students in a course and returns the latest
// state of the class. Each student's ID field must be populated.
func (acs *AtomicClassScribe) EnrollStudents(
	ctx context.Context,
	course classservice.Course,
	stu classservice.Students,
) (classservice.Class, error) {
	tx, err := acs.getTx()
	if err != nil {
		return classservice.Class{}, fmt.Errorf("EnrollStudents: %w", err)
	}

	rows := enrollmentRowsFromCouseAndStudents(course, stu)

	rows, err = enrollments.Insert(ctx, tx, rows)
	if err != nil {
		return classservice.Class{}, fmt.Errorf("EnrollStudents: %w", err)
	}

	class, err := acs.GetClassByCourseCode(ctx, course.Code)
	if err != nil {
		return classservice.Class{}, fmt.Errorf("EnrollStudents: %w", err)
	}

	return class, nil
}

func classFromRows(cRow courses.Row, sRows []students.Row) classservice.Class {
	return classservice.Class{
		Course:   courseFromRow(cRow),
		Students: studentsFromRows(sRows),
	}
}

func courseFromRow(cRow courses.Row) classservice.Course {
	return classservice.Course{
		ID:       cRow.ID,
		Code:     cRow.Code,
		Capacity: cRow.Capacity,
	}
}

func studentsFromRows(sRows []students.Row) classservice.Students {
	classStudents := make(classservice.Students, 0, len(sRows))

	for _, s := range sRows {
		classStudent := classservice.Student{
			ID:        s.ID,
			Name:      s.Name,
			Birthdate: s.Birthdate,
			Email:     s.Email,
		}
		classStudents = append(classStudents, classStudent)
	}

	return classStudents
}

func enrollmentRowsFromCouseAndStudents(
	c classservice.Course,
	s classservice.Students,
) []enrollments.Row {
	enrollmentRows := make([]enrollments.Row, 0, len(s))

	for _, stu := range s {
		row := enrollments.Row{
			CourseID:  c.ID,
			StudentID: stu.ID,
		}
		enrollmentRows = append(enrollmentRows, row)
	}

	return enrollmentRows
}
