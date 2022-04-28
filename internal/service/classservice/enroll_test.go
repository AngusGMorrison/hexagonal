//go:build unit

package classservice

import (
	"context"
	"errors"
	"log"
	"os"
	testing "testing"

	"github.com/angusgmorrison/hexagonal/internal/primitive"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/require"
)

func TestEnroll(t *testing.T) {
	t.Parallel()

	t.Run("validates EnrollmentRequest", func(t *testing.T) {
		t.Parallel()

		var (
			logger        = log.New(os.Stdout, "validates EnrollmentRequest ", log.LstdFlags)
			validate      = validator.New()
			scribeFactory = NewMockAtomicClassScribeFactory(t).Execute
			service       = New(logger, validate, scribeFactory)
		)

		testCases := []struct {
			name string
			req  EnrollmentRequest
		}{
			{
				name: "missing course code",
				req: EnrollmentRequest{
					CourseCode: "",
					Students:   Students{defaultStudent(t)},
				},
			},
			{
				name: "empty Students",
				req: EnrollmentRequest{
					CourseCode: "SICP",
					Students:   Students{},
				},
			},
			{
				name: "nil Students",
				req: EnrollmentRequest{
					CourseCode: "SICP",
					Students:   nil,
				},
			},
		}

		for _, tc := range testCases {
			tc := tc

			t.Run(tc.name, func(t *testing.T) {
				err := service.Enroll(context.Background(), tc.req)
				require.Error(t, err)
			})
		}
	})

	t.Run("validates course exists", func(t *testing.T) {
		t.Parallel()

		var (
			logger        = log.New(os.Stdout, "validates course exists ", log.LstdFlags)
			validate      = validator.New()
			scribeFactory = NewMockAtomicClassScribeFactory(t)
			scribe        = NewMockAtomicClassScribe(t)
			service       = New(logger, validate, scribeFactory.Execute)
			ctx           = context.Background()
			req           = defaultEnrollmentRequest(t)
			wantErr       = errors.New("course not found")
		)

		scribeFactory.On("Execute").Return(scribe)
		scribe.On("Begin", ctx).Return(nil)
		scribe.On("Rollback").Return(nil)
		scribe.On(
			"GetClassByCourseCode",
			ctx,
			req.CourseCode,
		).Return(Class{}, wantErr)

		err := service.Enroll(ctx, req)
		require.ErrorIs(t, err, wantErr)
	})

	t.Run("validates students are registered", func(t *testing.T) {
		t.Parallel()

		var (
			logger        = log.New(os.Stdout, "validates students are registered ", log.LstdFlags)
			validate      = validator.New()
			scribeFactory = NewMockAtomicClassScribeFactory(t)
			scribe        = NewMockAtomicClassScribe(t)
			service       = New(logger, validate, scribeFactory.Execute)
			ctx           = context.Background()
			req           = defaultEnrollmentRequest(t)
			wantErr       = UnregisteredStudentsError{Students: req.Students}
		)

		scribeFactory.On("Execute").Return(scribe)
		scribe.On("Begin", ctx).Return(nil)
		scribe.On("Rollback").Return(nil)
		scribe.On("GetClassByCourseCode", ctx, req.CourseCode).Return(defaultClass(t), nil)
		scribe.On(
			"GetStudentsByEmail",
			ctx,
			req.Students.EmailAddresses(),
		).Return(Students{}, nil)

		err := service.Enroll(ctx, req)

		var gotErr UnregisteredStudentsError
		require.ErrorAs(t, err, &gotErr)
		require.Equal(t, wantErr, gotErr, "unequal UnregisteredStudentsErrors")
	})

	t.Run("validates that students aren't already enrolled", func(t *testing.T) {
		t.Parallel()

		var (
			logger        = log.New(os.Stdout, "validates students are registered ", log.LstdFlags)
			validate      = validator.New()
			scribeFactory = NewMockAtomicClassScribeFactory(t)
			scribe        = NewMockAtomicClassScribe(t)
			service       = New(logger, validate, scribeFactory.Execute)
			ctx           = context.Background()
			req           = defaultEnrollmentRequest(t)
		)

		scribeFactory.On("Execute").Return(scribe)
		scribe.On("Begin", ctx).Return(nil)
		scribe.On("Rollback").Return(nil)
		scribe.On("GetClassByCourseCode", ctx, req.CourseCode).Return(defaultClass(t), nil)

		registeredStudent := defaultStudent(t)
		registeredStudent.ID = 1
		registeredStudents := Students{registeredStudent}
		wantErr := AlreadyEnrolledError{Students: registeredStudents}

		scribe.On(
			"GetStudentsByEmail",
			ctx,
			req.Students.EmailAddresses(),
		).Return(registeredStudents, nil)

		err := service.Enroll(ctx, req)

		var gotErr AlreadyEnrolledError
		require.ErrorAs(t, err, &gotErr)
		require.Equal(t, wantErr, gotErr, "unequal AlreadyEnrolledErrors")
	})

	t.Run("validates that class has capacity for enrolling students", func(t *testing.T) {
		t.Parallel()

		var (
			logger        = log.New(os.Stdout, "validates students are registered ", log.LstdFlags)
			validate      = validator.New()
			scribeFactory = NewMockAtomicClassScribeFactory(t)
			scribe        = NewMockAtomicClassScribe(t)
			service       = New(logger, validate, scribeFactory.Execute)
			ctx           = context.Background()
			req           = defaultEnrollmentRequest(t)
		)

		class := Class{
			Course: Course{
				Code:     "SICP",
				Capacity: 0,
			},
		}

		wantErr := OversubscribedError{
			CourseCode:           class.Code,
			AvailableSpaces:      class.availableSpaces(),
			AttemptedEnrollments: uint32(len(req.Students)),
		}

		scribeFactory.On("Execute").Return(scribe)
		scribe.On("Begin", ctx).Return(nil)
		scribe.On("Rollback").Return(nil)
		scribe.On("GetClassByCourseCode", ctx, req.CourseCode).Return(class, nil)

		registeredStudent := defaultStudent(t)
		registeredStudent.ID = 1
		registeredStudents := Students{registeredStudent}

		scribe.On(
			"GetStudentsByEmail",
			ctx,
			req.Students.EmailAddresses(),
		).Return(registeredStudents, nil)

		err := service.Enroll(ctx, req)

		var gotErr OversubscribedError
		require.ErrorAs(t, err, &gotErr)
		require.Equal(t, wantErr, gotErr, "unequal OversubscribedErrors")
	})

	t.Run("enrolls students", func(t *testing.T) {
		t.Parallel()

		var (
			logger        = log.New(os.Stdout, "validates students are registered ", log.LstdFlags)
			validate      = validator.New()
			scribeFactory = NewMockAtomicClassScribeFactory(t)
			scribe        = NewMockAtomicClassScribe(t)
			service       = New(logger, validate, scribeFactory.Execute)
			ctx           = context.Background()
			req           = defaultEnrollmentRequest(t)
		)

		class := Class{
			Course: Course{
				Code:     "SICP",
				Capacity: 1,
			},
		}

		scribeFactory.On("Execute").Return(scribe)
		scribe.On("Begin", ctx).Return(nil)
		scribe.On("Rollback").Return(nil)
		scribe.On("GetClassByCourseCode", ctx, req.CourseCode).Return(class, nil)

		registeredStudent := defaultStudent(t)
		registeredStudent.ID = 1
		registeredStudents := Students{registeredStudent}

		scribe.On(
			"GetStudentsByEmail",
			ctx,
			req.Students.EmailAddresses(),
		).Return(registeredStudents, nil)
		scribe.On(
			"EnrollStudents",
			ctx,
			class.Course,
			registeredStudents,
		).Return(class, nil)
		scribe.On("Commit").Return(nil)

		err := service.Enroll(ctx, req)
		require.NoError(t, err)
	})
}

func defaultEnrollmentRequest(t *testing.T) EnrollmentRequest {
	t.Helper()

	return EnrollmentRequest{
		CourseCode: "SICP",
		Students: Students{
			defaultStudent(t),
		},
	}
}

func defaultClass(t *testing.T) Class {
	t.Helper()

	return Class{
		Course:   defaultCourse(),
		Students: Students{defaultStudent(t)},
	}
}

func defaultCourse() Course {
	return Course{
		Code:     "SICP",
		Capacity: 2,
	}
}

func defaultStudent(t *testing.T) Student {
	t.Helper()

	birthdate, err := primitive.ParseBirthdate("1990-03-04")
	require.NoError(t, err, "parse birthdate")

	return Student{
		Name:      "Ramdas Tifft",
		Birthdate: birthdate,
		Email:     "r.tifft@gmail.com",
	}
}
