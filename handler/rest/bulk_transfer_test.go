//go:build unit

package rest

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/angusgmorrison/hexagonal/envconfig"
	restmock "github.com/angusgmorrison/hexagonal/handler/rest/mock"
	"github.com/angusgmorrison/hexagonal/service"
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

func TestHandleBulkTransfer_BadRequest(t *testing.T) {
	t.Parallel()
	require := require.New(t)

	logger := log.New(os.Stdout, "TestHandleBulkTransfer_BadRequest ", log.LstdFlags)

	testCases := []struct {
		name            string
		fixtureFilename string
	}{
		{
			"401 no organization_name",
			"401_no_name.json",
		},
		{
			"401 no organization_bic",
			"401_no_bic.json",
		},
		{
			"401 no organization_iban",
			"401_no_iban.json",
		},
		{
			"401 no credit transfers",
			"401_no_transfers.json",
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var (
				transactionService = restmock.TransactionService{}
				server             = NewServer(logger, defaultConfig(), &transactionService)
			)

			fixturePath := filepath.Join("testdata", tc.fixtureFilename)
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
	}
}

func TestHandleBulkTransfer_ValidRequest(t *testing.T) {
	t.Parallel()
	require := require.New(t)

	logger := log.New(os.Stdout, "TestHandleBulkTransfer_ValidRequest ", log.LstdFlags)

	testCases := []struct {
		name            string
		fixtureFilename string
		serviceErr      error
		statusCode      int
	}{
		{
			name:            "201 created",
			fixtureFilename: "bulk_transfer_request.json",
			serviceErr:      nil,
			statusCode:      http.StatusCreated,
		},
		{
			name:            "422 unprocessable entity",
			fixtureFilename: "bulk_transfer_request.json",
			serviceErr:      service.ErrInsufficientFunds,
			statusCode:      http.StatusUnprocessableEntity,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var (
				transactionService = restmock.TransactionService{}
				server             = NewServer(logger, defaultConfig(), &transactionService)
			)

			fixturePath := filepath.Join("testdata", tc.fixtureFilename)
			fixtureBytes, err := ioutil.ReadFile(fixturePath)
			require.NoError(err)

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/bulk_transfer", bytes.NewReader(fixtureBytes))
			r.Header.Set("content-type", "application/json")

			expectedBTR := bulkTransferRequest{
				OrganizationName: "ACME Corp",
				OrganizationBIC:  "OIVUSCLQXXX",
				OrganizationIBAN: "FR10474608000002006107XXXXX",
				CreditTransfers: creditTransfers{
					{
						Amount:           json.Number("23.17"),
						Currency:         "EUR",
						CounterpartyName: "Bip Bip",
						CounterpartyBIC:  "CRLYFRPPTOU",
						CounterpartyIBAN: "EE383680981021245685",
						Description:      "Neverland/6318",
					},
				},
			}

			transactionService.On(
				"BulkTransaction",
				mock.AnythingOfType("*gin.Context"),
				expectedBTR.toDomain(),
			).Return(tc.serviceErr)

			server.ServeHTTP(w, r)

			transactionService.AssertExpectations(t)

			body, err := ioutil.ReadAll(w.Body)
			require.NoError(err)

			require.Equal(tc.statusCode, w.Code, "response status")
			require.Empty(body, "response body")
		})
	}
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
