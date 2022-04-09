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

	"github.com/angusgmorrison/hexagonal/internal/adapter/envconfig"
	"github.com/angusgmorrison/hexagonal/internal/controller"
	"github.com/angusgmorrison/hexagonal/internal/controller/mock"
	"github.com/stretchr/testify/require"
)

func TestHandleBulkTransfer(t *testing.T) {
	t.Parallel()
	require := require.New(t)

	logger := log.New(os.Stdout, "hexagonal_test ", log.LstdFlags)

	t.Run("malformed request", func(t *testing.T) {
		var (
			repo    = mock.Repository{}
			service = controller.NewTransferController(&repo)
			server  = NewServer(logger, defaultConfig(), service)
		)

		fixturePath := filepath.Join(fixtureDir(), "401_bad_request.json")
		fixtureBytes, err := ioutil.ReadFile(fixturePath)
		require.NoError(err)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/bulk_transfer", bytes.NewReader(fixtureBytes))
		r.Header.Set("content-type", "application/json")

		server.ServeHTTP(w, r)

		body, err := ioutil.ReadAll(w.Body)
		require.NoError(err)

		require.Equal(http.StatusBadRequest, w.Code, "response status")
		require.Empty(body, "response body")
	})

	t.Run("transfer created", func(t *testing.T) {
		var (
			repo    = mock.Repository{}
			service = controller.NewTransferController(&repo)
			server  = NewServer(logger, defaultConfig(), service)
		)

		fixturePath := filepath.Join(fixtureDir(), "201_created.json")
		fixtureBytes, err := ioutil.ReadFile(fixturePath)
		require.NoError(err)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/bulk_transfer", bytes.NewReader(fixtureBytes))
		r.Header.Set("content-type", "application/json")

		server.ServeHTTP(w, r)

		body, err := ioutil.ReadAll(w.Body)
		require.NoError(err)

		require.Equal(http.StatusCreated, w.Code, "response status")
		require.Empty(body, "response body")
	})

	t.Run("transfer creation error", func(t *testing.T) {
		var (
			repo    = mock.Repository{Err: controller.ErrInsufficientFunds}
			service = controller.NewTransferController(&repo)
			server  = NewServer(logger, defaultConfig(), service)
		)

		fixturePath := filepath.Join(fixtureDir(), "422_insufficient_funds.json")
		fixtureBytes, err := ioutil.ReadFile(fixturePath)
		require.NoError(err)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/bulk_transfer", bytes.NewReader(fixtureBytes))
		r.Header.Set("content-type", "application/json")

		server.ServeHTTP(w, r)

		body, err := ioutil.ReadAll(w.Body)
		require.NoError(err)

		require.Equal(http.StatusUnprocessableEntity, w.Code, "response status")
		require.Empty(body, "response body")
	})
}

func fixtureDir() string {
	return filepath.Join("..", "..", "..", "fixtures", "requests")
}

func defaultConfig() envconfig.EnvConfig {
	return envconfig.EnvConfig{
		App: envconfig.App{
			Name: "hexagonal",
			Env:  "test",
			Root: filepath.Join(string(filepath.Separator), "usr", "src", "app"),
		},
	}
}
