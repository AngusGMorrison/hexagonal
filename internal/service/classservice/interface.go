package classservice

import (
	"context"

	"github.com/angusgmorrison/hexagonal/internal/primitive"
	"github.com/go-playground/validator/v10"
)

// Interface specifies the business operations of the service.
//
// Exporting an interface for which this package also provides an implementation
// makes it simple for dependant packages to mock the service while ensuring
// that the service package is authoritative.
type Interface interface {
	Enroll(ctx context.Context, er EnrollmentRequest) error
}

// New configures and returns an Interface implementation.
func New(
	logger logger,
	validate *validator.Validate,
	repo AtomicRepository,
) Interface {
	return &classService{
		logger:   logger,
		validate: validate,
		repo:     repo,
	}
}

// classService implements classservice.Interface.
type classService struct {
	logger   logger
	validate *validator.Validate
	repo     AtomicRepository
}

type AtomicOperation func(context.Context, Repository) error

type AtomicRepository interface {
	Execute(context.Context, AtomicOperation) error
}

type Repository interface {
	// GetClassByCourseCode loads a course and its students based on a course
	// code.
	GetClassByCourseCode(ctx context.Context, courseCode string) (Class, error)

	// GetStudentsByEmail loads all the students corresponding to the email
	// addresses provided.
	GetStudentsByEmail(ctx context.Context, emails []primitive.EmailAddress) (Students, error)

	// Enroll writes the enrollment of students in a class to a repository.
	EnrollStudents(ctx context.Context, c Course, s Students) (Class, error)
}

type logger interface {
	Printf(format string, args ...any)
}
