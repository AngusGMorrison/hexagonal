package classservice

import "fmt"

// OversubscribedError is returned when attempting to enroll more students than
// a course has spaces available.
type OversubscribedError struct {
	CourseCode           string
	AvailableSpaces      uint32
	AttemptedEnrollments uint32
}

func (oe OversubscribedError) Error() string {
	return fmt.Sprintf(
		"attmepted to enroll %d students, but course %q has only %d spaces",
		oe.AttemptedEnrollments, oe.CourseCode, oe.AvailableSpaces)
}

type UnregisteredStudentsError struct {
	Students Students
}

func (use UnregisteredStudentsError) Error() string {
	return fmt.Sprintf("attempted to enroll unregistered students: %s", use.Students)
}

// AlreadyEnrolledError is returned when attempting to enroll students who are
// already enrolled in the class.
type AlreadyEnrolledError struct {
	Students Students
}

func (are AlreadyEnrolledError) Error() string {
	return fmt.Sprintf("students %s are already registered", are.Students)
}
