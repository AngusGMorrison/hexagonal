//go:build !integration

package handler

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/angusgmorrison/hexagonal/internal/app/transferdomain"
	"github.com/angusgmorrison/hexagonal/internal/app/transferdomain/mock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTransferHandler(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	logger := log.New(os.Stdout, "hexagonal_test ", log.LstdFlags)
	service := transferdomain.NewService(&mock.Repository{})
	handler := NewTransferHandler(logger, service)

	assert.Equal(logger, handler.logger, "logger")
	assert.Equal(service, handler.service, "service")
}

func TestBulkTransfer(t *testing.T) {
	t.Parallel()
	require := require.New(t)

	logger := log.New(os.Stdout, "hexagonal_test ", log.LstdFlags)

	t.Run("malformed request", func(t *testing.T) {
		var (
			repo    = mock.Repository{}
			service = transferdomain.NewService(&repo)
			handler = NewTransferHandler(logger, service)
			router  = newRouter(handler)
		)

		fixturePath := filepath.Join(fixtureDir(), "401_bad_request.json")
		fixtureBytes, err := ioutil.ReadFile(fixturePath)
		require.NoError(err)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/bulk_transfer", bytes.NewReader(fixtureBytes))
		r.Header.Set("content-type", "application/json")

		router.ServeHTTP(w, r)

		body, err := ioutil.ReadAll(w.Body)
		require.NoError(err)

		require.Equal(http.StatusBadRequest, w.Code, "response status")
		require.Empty(body, "response body")
	})

	t.Run("transfer created", func(t *testing.T) {
		var (
			repo    = mock.Repository{}
			service = transferdomain.NewService(&repo)
			handler = NewTransferHandler(logger, service)
			router  = newRouter(handler)
		)

		fixturePath := filepath.Join(fixtureDir(), "201_created.json")
		fixtureBytes, err := ioutil.ReadFile(fixturePath)
		require.NoError(err)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/bulk_transfer", bytes.NewReader(fixtureBytes))
		r.Header.Set("content-type", "application/json")

		router.ServeHTTP(w, r)

		body, err := ioutil.ReadAll(w.Body)
		require.NoError(err)

		require.Equal(http.StatusCreated, w.Code, "response status")
		require.Empty(body, "response body")
	})

	t.Run("transfer creation error", func(t *testing.T) {
		var (
			repo    = mock.Repository{Err: transferdomain.ErrInsufficientFunds}
			service = transferdomain.NewService(&repo)
			handler = NewTransferHandler(logger, service)
			router  = newRouter(handler)
		)

		fixturePath := filepath.Join(fixtureDir(), "422_insufficient_funds.json")
		fixtureBytes, err := ioutil.ReadFile(fixturePath)
		require.NoError(err)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/bulk_transfer", bytes.NewReader(fixtureBytes))
		r.Header.Set("content-type", "application/json")

		router.ServeHTTP(w, r)

		body, err := ioutil.ReadAll(w.Body)
		require.NoError(err)

		require.Equal(http.StatusUnprocessableEntity, w.Code, "response status")
		require.Empty(body, "response body")
	})
}

func fixtureDir() string {
	return filepath.Join("..", "..", "..", "..", "..", "fixtures", "requests")
}

func newRouter(h *TransferHandler) *gin.Engine {
	r := gin.Default()
	r.POST("/bulk_transfer", h.BulkTransfer)

	return r
}
