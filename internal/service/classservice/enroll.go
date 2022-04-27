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

	registeredStudents, err := scribe.GetStudentsByEmail(ctx, req.Students.EmailAddresses())
	if err != nil {
		return fmt.Errorf("ensureStudentsAreRegistered: %w", err)
	}

	if len(registeredStudents) < len(req.Students) {
		return UnregisteredStudentsError{
			Students: slice.Difference(registeredStudents, req.Students),
		}
	}

	if err := verifyStudentsNotAlreadyEnrolled(class, registeredStudents); err != nil {
		return err
	}

	if !class.hasCapacityFor(registeredStudents) {
		return OversubscribedError{
			CourseCode:           class.Code,
			AvailableSpaces:      class.availableSpaces(),
			AttemptedEnrollments: uint32(len(registeredStudents)),
		}
	}

	_, err = scribe.EnrollStudents(ctx, class.Course, registeredStudents)
	if err != nil {
		return fmt.Errorf("Enroll: %w", err)
	}

	if err := scribe.Commit(); err != nil {
		return fmt.Errorf("Enroll: %w", err)
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
