//go:build integration

package integration_test

import (
	"bytes"
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/angusgmorrison/hexagonal/internal/primitive"
	"github.com/angusgmorrison/hexagonal/internal/storage/sql"
	"github.com/angusgmorrison/hexagonal/internal/storage/sql/table/courses"
	"github.com/angusgmorrison/hexagonal/internal/storage/sql/table/enrollments"
	"github.com/angusgmorrison/hexagonal/internal/storage/sql/table/students"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnrollment(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	env := defaultEnvConfig()
	logger := log.New(os.Stdout, "TestBulkTransfer ", log.LstdFlags)

	infra, err := newInfrastructure(env, logger)
	require.NoError(err, "newInfrastructure")

	t.Cleanup(infra.cleanup)

	t.Run("success", func(t *testing.T) {
		defer truncateTables(t, infra.db)

		// Seed the database with one course, one enrolled student and one
		// unenrolled student.
		courseRows, err := courses.Insert(
			context.Background(),
			infra.db,
			[]courses.Row{defaultCourseRow()},
		)
		require.NoError(err, "insert default course")

		course := courseRows[0]

		studentRows, err := students.Insert(
			context.Background(),
			infra.db,
			[]students.Row{berthe(t), kassandra(t)},
		)
		require.NoError(err, "insert students")

		enrolledStudent := studentRows[0]
		studentToEnroll := studentRows[1]

		_, err = enrollments.Insert(
			context.Background(),
			infra.db,
			[]enrollments.Row{
				{CourseID: course.ID, StudentID: enrolledStudent.ID},
			},
		)
		require.NoError(err, "insert enrollment")

		// Perform the enrollment request for the unenrolled student.
		bodyBytes, err := ioutil.ReadFile("testdata/201_created.json")
		require.NoError(err, "read request fixture")

		req, err := http.NewRequest(http.MethodPost, enrollmentURL(), bytes.NewReader(bodyBytes))
		require.NoError(err, "create request")

		req.Header.Set("Content-Type", "application/json")

		res, err := infra.client.Do(req)
		require.NoError(err, "perform request")

		defer func() { _ = res.Body.Close() }()

		// Assert that the enrollment request was performed as expected.
		assert.Equal(http.StatusCreated, res.StatusCode, "unexpected status code")

		resBodyBytes, err := ioutil.ReadAll(res.Body)
		require.NoError(err, "read response body")

		assert.Len(resBodyBytes, 0, "unexpected response body")

		gotStudents, err := students.OnCourse(context.Background(), infra.db, course.ID)
		require.NoError(err, "get students on course")

		assert.Len(gotStudents, 2)

		emails := make([]primitive.EmailAddress, 0, len(gotStudents))
		for _, student := range gotStudents {
			emails = append(emails, student.Email)
		}

		assert.Contains(emails, enrolledStudent.Email, "enrolledStudent no longer enrolled")
		assert.Contains(emails, studentToEnroll.Email, "studentToEnroll was not enrolled")
	})

	t.Run("course does not exist", func(t *testing.T) {
		defer truncateTables(t, infra.db)
	})

	t.Run("student not registered", func(t *testing.T) {
		defer truncateTables(t, infra.db)
	})

	t.Run("student already enrolled in class", func(t *testing.T) {
		defer truncateTables(t, infra.db)
	})

	t.Run("class oversubscribed", func(t *testing.T) {
		defer truncateTables(t, infra.db)
	})
}

func truncateTables(t *testing.T, exec sql.Execer) {
	err := courses.Truncate(context.Background(), exec)
	assert.NoError(t, err, "truncate courses")

	err = students.Truncate(context.Background(), exec)
	assert.NoError(t, err, "truncate students")

	err = enrollments.Truncate(context.Background(), exec)
	assert.NoError(t, err, "truncate enrollments")
}

func defaultCourseRow() courses.Row {
	return courses.Row{
		Code:        "SICP",
		Title:       "Structure and Interpretation of Computer Programs",
		Capacity:    2,
		Description: "The classic introduction to computer programming.",
	}
}

func berthe(t *testing.T) students.Row {
	t.Helper()

	birthdate, err := primitive.ParseBirthdate("1987-09-03")
	require.NoError(t, err)

	return students.Row{
		Name:      "Berthe Archibald",
		Birthdate: time.Time(birthdate),
		Email:     "berthe@archibaldindustries.com",
	}
}

func kassandra(t *testing.T) students.Row {
	t.Helper()

	birthdate, err := primitive.ParseBirthdate("1996-07-07")
	require.NoError(t, err)

	return students.Row{
		Name:      "Kassandra Madhukar",
		Birthdate: time.Time(birthdate),
		Email:     "km1996@gmail.com",
	}
}

func blandinus(t *testing.T) students.Row {
	t.Helper()

	birthdate, err := primitive.ParseBirthdate("1991-09-18")
	require.NoError(t, err)

	return students.Row{
		Name:      "Blandinus Branislava",
		Birthdate: time.Time(birthdate),
		Email:     "blandinus@gmail.com",
	}
}
