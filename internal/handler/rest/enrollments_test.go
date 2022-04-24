//go:build unit

package rest

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/angusgmorrison/hexagonal/internal/envconfig"
	"github.com/angusgmorrison/hexagonal/internal/service/classservice"
	"github.com/angusgmorrison/hexagonal/internal/service/classservice/mocks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	originalMode := gin.Mode()

	defer gin.SetMode(originalMode)

	gin.SetMode(gin.TestMode)

	code := m.Run()

	os.Exit(code)
}

func TestHandleCreateEnrollments(t *testing.T) {
	t.Parallel()
	require := require.New(t)

	const endpoint = "/enroll"

	t.Run("responds 415 Unsupported Media Type to non-JSON requests", func(t *testing.T) {
		t.Parallel()

		var (
			logger       = log.New(os.Stdout, "TestHandleCreateEnrollments ", log.LstdFlags)
			classService = mocks.NewInterface(t)
			server       = NewServer(logger, defaultConfig(), classService)
			r            = httptest.NewRequest(http.MethodPost, endpoint, nil)
			w            = httptest.NewRecorder()
		)

		server.ServeHTTP(w, r)

		require.Equal(http.StatusUnsupportedMediaType, w.Code, "unexpected status code")
	})

	t.Run("handles syntactically valid requests", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			name       string
			serviceErr error
			wantStatus int
		}{
			{
				name:       "created",
				serviceErr: nil,
				wantStatus: http.StatusCreated,
			},
			{
				name:       "class oversubscribed",
				serviceErr: classservice.OversubscribedError{},
				wantStatus: http.StatusUnprocessableEntity,
			},
			{
				name:       "unregistered students",
				serviceErr: classservice.UnregisteredStudentsError{},
				wantStatus: http.StatusUnprocessableEntity,
			},
			{
				name:       "students already enrolled",
				serviceErr: classservice.AlreadyEnrolledError{},
				wantStatus: http.StatusUnprocessableEntity,
			},
		}

		for _, tc := range testCases {
			tc := tc

			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				fixturePath := filepath.Join("testdata", "enrollment_request.json")
				fixtureBytes, err := ioutil.ReadFile(fixturePath)
				require.NoError(err)

				var (
					logger       = log.New(os.Stdout, "TestHandleCreateEnrollments ", log.LstdFlags)
					classService = mocks.NewInterface(t)
					server       = NewServer(logger, defaultConfig(), classService)
					r            = httptest.NewRequest(http.MethodPost, endpoint, bytes.NewReader(fixtureBytes))
					w            = httptest.NewRecorder()
				)

				r.Header.Set("content-type", string(applicationJSON))

				expectedBirthdate, err := time.Parse(birthdateLayout, "1991-10-03")
				require.NoError(err)

				expectedEnrollmentRequest := classservice.EnrollmentRequest{
					CourseCode: "SICP",
					Students: classservice.Students{
						{
							Name:      "Ramdas Tifft",
							Birthdate: expectedBirthdate,
							Email:     "r.tifft@gmail.com",
						},
					},
				}

				classService.On(
					"Enroll",
					mock.AnythingOfType("*gin.Context"),
					expectedEnrollmentRequest,
				).Return(tc.serviceErr)

				server.ServeHTTP(w, r)

				require.Equal(tc.wantStatus, w.Code, "unexpected status code")
			})
		}
	})
}

func defaultConfig() envconfig.EnvConfig {
	return envconfig.EnvConfig{
		App: envconfig.App{
			Name:    "hexagonal",
			Env:     "test",
			Root:    filepath.Join(string(filepath.Separator), "usr", "src", "app"),
			GinMode: gin.TestMode,
		},
	}
}
