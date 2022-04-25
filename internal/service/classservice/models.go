package classservice

import (
	"strings"
	"time"

	"github.com/angusgmorrison/hexagonal/internal/primitive"
)

// Course represents the service's understanding of a course. Note that although
// there are more columns in the course database table, these don't feature in
// the service's model because the service has no need to know about them.
type Course struct {
	ID       int64
	Code     string
	Capacity uint32
}

// Students is a convenience wrapper.
type Students []Student

func (s Students) IDs() []int64 {
	ids := make([]int64, 0, len(s))

	for _, student := range s {
		ids = append(ids, student.ID)
	}

	return ids
}

func (s Students) EmailAddresses() []primitive.EmailAddress {
	emails := make([]primitive.EmailAddress, 0, len(s))

	for _, student := range s {
		emails = append(emails, student.Email)
	}

	return emails
}

// String returns a comma-separated list of student email addresses. Satisfies
// fmt.Stringer.
func (s Students) String() string {
	var builder strings.Builder

	for i, student := range s {
		if i != 0 {
			builder.WriteString(", ")
		}

		builder.WriteString(string(student.Email))
	}

	return builder.String()
}

// Student represents a student.
type Student struct {
	ID        int64
	Name      string
	Birthdate time.Time
	Email     primitive.EmailAddress
}

// Class represents a course and its enrolled students. Note that the existence
// of a database join table between classes and students is invisible. Their
// relationship is described entirely by their colocation in the Class struct.
type Class struct {
	Course
	Students
}

func (c Class) hasCapacityFor(s Students) bool {
	return c.availableSpaces() >= uint32(len(s))
}

func (c Class) availableSpaces() uint32 {
	return c.Course.Capacity - uint32(len(c.Students))
}

// EnrollmentRequest represents a batch of students to be enrolled in a course.
type EnrollmentRequest struct {
	CourseCode string   `validate:"required"`
	Students   Students `validate:"min=1"`
}
