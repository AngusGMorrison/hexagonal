package classservice

import (
	"context"
	"fmt"

	"github.com/angusgmorrison/hexagonal/pkg/slice"
)

// Enroll enrolls the students contained in the given EnrollmentRequest in the
// course matching the request's CourseCode.
//
// If the course does not exist, any of the students do not exist, any of the
// students are already enrolled in the course, or enrolling the students in the
// course would cause the course to be oversubscribed, an error is returned.
func (svc *classService) Enroll(ctx context.Context, req EnrollmentRequest) error {
	if err := svc.validate.Struct(req); err != nil {
		return fmt.Errorf("Enroll: %w", err)
	}

	enroll := func(ctx context.Context, repo Repository) error {
		class, err := repo.GetClassByCourseCode(ctx, req.CourseCode)
		if err != nil {
			return fmt.Errorf("Enroll: %w", err)
		}

		registeredStudents, err := repo.GetStudentsByEmail(ctx, req.Students.EmailAddresses())
		if err != nil {
			return fmt.Errorf("Enroll: %w", err)
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

		class, err = repo.EnrollStudents(ctx, class.Course, registeredStudents)
		if err != nil {
			return fmt.Errorf("Enroll: %w", err)
		}

		return nil
	}

	if err := svc.repo.Execute(ctx, enroll); err != nil {
		return err
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
