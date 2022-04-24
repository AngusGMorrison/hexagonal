package classservice

import (
	"context"
	"fmt"

	"github.com/angusgmorrison/hexagonal/pkg/slice"
)

func (svc *classService) Enroll(ctx context.Context, req EnrollmentRequest) error {
	if err := svc.validate.Struct(req); err != nil {
		return fmt.Errorf("Enroll: %w", err)
	}

	scribe := svc.newAtomicScribe()

	if err := scribe.Begin(ctx); err != nil {
		return fmt.Errorf("Enroll: %w", err)
	}

	defer func() { _ = scribe.Rollback() }()

	class, err := scribe.GetClassByCourseCode(ctx, req.CourseCode)
	if err != nil {
		return fmt.Errorf("Enroll: %w", err)
	}

	if err := verifyStudentsAreRegistered(ctx, scribe, req.Students); err != nil {
		return err
	}

	if err := verifyStudentsNotAlreadyEnrolled(class, req.Students); err != nil {
		return err
	}

	if !class.hasCapacityFor(req.Students) {
		return OversubscribedError{
			CourseCode:           class.Code,
			AvailableSpaces:      class.remainingSpaces(),
			AttemptedEnrollments: uint32(len(req.Students)),
		}
	}

	_, err = scribe.EnrollStudents(ctx, class.Course, req.Students)
	if err != nil {
		return fmt.Errorf("Enroll: %w", err)
	}

	if err := scribe.Commit(); err != nil {
		return fmt.Errorf("Enroll: %w", err)
	}

	return nil
}

func verifyStudentsAreRegistered(
	ctx context.Context,
	scribe AtomicClassScribe,
	students Students,
) error {
	// Check that all students attempting to enroll are registered.
	registeredStudents, err := scribe.GetStudentsByEmail(ctx, students.EmailAddresses())
	if err != nil {
		return fmt.Errorf("ensureStudentsAreRegistered: %w", err)
	}

	if len(registeredStudents) < len(students) {
		return UnregisteredStudentsError{
			Students: slice.Difference(registeredStudents, students),
		}
	}

	return nil
}

func verifyStudentsNotAlreadyEnrolled(
	class Class,
	students Students,
) error {
	alreadyEnrolledEmails := slice.Intersection(
		class.Students.EmailAddresses(), students.EmailAddresses())

	if len(alreadyEnrolledEmails) > 0 {
		alreadyEnrolledEmailSet := slice.ToSet(alreadyEnrolledEmails)
		alreadyEnrolledStudents := slice.Filter(students, func(student Student) bool {
			return alreadyEnrolledEmailSet[student.Email]
		})

		return AlreadyEnrolledError{Students: alreadyEnrolledStudents}
	}

	return nil
}
