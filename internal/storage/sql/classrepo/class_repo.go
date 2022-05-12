// Package classrepo provides implementations of
// classservice.AtomicRepository and classservice.Repository for use with an SQL
// database.
package classrepo

import (
	"context"
	"fmt"

	"github.com/angusgmorrison/hexagonal/internal/primitive"
	"github.com/angusgmorrison/hexagonal/internal/service/classservice"
	"github.com/angusgmorrison/hexagonal/internal/storage/sql"
	"github.com/angusgmorrison/hexagonal/internal/storage/sql/table/courses"
	"github.com/angusgmorrison/hexagonal/internal/storage/sql/table/enrollments"
	"github.com/angusgmorrison/hexagonal/internal/storage/sql/table/students"
)

// AtomicRepository satisfies classservice.AtomicRepository.
type AtomicRepository struct {
	db sql.Database
}

var _ classservice.AtomicRepository = (*AtomicRepository)(nil)

// NewAtomic instantiates a new AtomicRepository using the database provided.
func NewAtomic(db sql.Database) *AtomicRepository {
	return &AtomicRepository{db: db}
}

// Execute decorates the given AtomicOperation with a transaction. If the
// AtomicOperation returns an error, the transaction is rolled back. Otherwise,
// the transaction is committed.
func (ar *AtomicRepository) Execute(
	ctx context.Context,
	op classservice.AtomicOperation,
) error {
	tx, err := ar.db.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() { _ = tx.Rollback() }()

	classRepoWithTransaction := Repository{operator: tx}

	if err := op(ctx, &classRepoWithTransaction); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

// Repository satisfies classservice.Repository. It is agnostic as to whether
// its sql.TableOperator is a database or transaction.
type Repository struct {
	operator sql.TableOperator
}

var _ classservice.Repository = (*Repository)(nil)

// GetClass returns a course and all its enrolled students from the course code
// provided.
func (r *Repository) GetClassByCourseCode(
	ctx context.Context,
	courseCode string,
) (classservice.Class, error) {
	courseRow, err := courses.FindByCode(ctx, r.operator, courseCode)
	if err != nil {
		return classservice.Class{}, fmt.Errorf("GetClassByCourseCode(%q): %w", courseCode, err)
	}

	studentRows, err := students.OnCourse(ctx, r.operator, courseRow.ID)
	if err != nil {
		return classservice.Class{}, fmt.Errorf("GetClassByCourseCode(%q): %w", courseCode, err)
	}

	return classFromRows(courseRow, studentRows), nil
}

// GetStudentsByEmail returns all the students whose email addresses are
// contained in the slice provided.
func (r *Repository) GetStudentsByEmail(
	ctx context.Context,
	emails []primitive.EmailAddress,
) (classservice.Students, error) {
	studentRows, err := students.SelectByEmail(ctx, r.operator, emails)
	if err != nil {
		return nil, fmt.Errorf("GetStudentsByEmail(%v): %w", emails, err)
	}

	return studentsFromRows(studentRows), nil
}

// EnrollStudents enrolls the given students in a course and returns the latest
// state of the class. Each student's ID field must be populated.
func (r *Repository) EnrollStudents(
	ctx context.Context,
	course classservice.Course,
	stu classservice.Students,
) (classservice.Class, error) {
	rows := enrollmentRowsFromCouseAndStudents(course, stu)

	rows, err := enrollments.Insert(ctx, r.operator, rows)
	if err != nil {
		return classservice.Class{}, fmt.Errorf("EnrollStudents: %w", err)
	}

	class, err := r.GetClassByCourseCode(ctx, course.Code)
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
			Birthdate: primitive.Birthdate(s.Birthdate),
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
