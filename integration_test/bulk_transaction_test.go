//go:build integration

package integration_test

import (
	"bytes"
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/angusgmorrison/hexagonal/repository/sql"
	"github.com/angusgmorrison/hexagonal/repository/sql/table/bankaccounts"
	"github.com/angusgmorrison/hexagonal/repository/sql/table/transactions"
	"github.com/angusgmorrison/hexagonal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBulkTransfer(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	env := defaultEnvConfig()
	logger := log.New(os.Stdout, "TestBulkTransfer ", log.LstdFlags)

	infra, err := newInfrastructure(env, logger)
	require.NoError(err, "newInfrastructure")

	t.Cleanup(infra.cleanup)

	t.Run("bank account has insufficient funds", func(t *testing.T) {
		defer truncateBulkTransactionTables(t, infra.db)

		expectedBankAccount, err := bankaccounts.Insert(
			context.Background(), infra.db, defaultBankAccount())
		require.NoError(err, "insertBankAccount")

		// TODO: Embed fixtures
		fixturePath := filepath.Join("..", "fixtures", "requests", "422_insufficient_funds.json")
		fixtureBytes, err := ioutil.ReadFile(fixturePath)
		require.NoError(err, "read fixture file")

		req, err := http.NewRequest(http.MethodPost, bulkTransferURL(), bytes.NewReader(fixtureBytes))
		require.NoError(err, "create HTTP request")

		req.Header.Add("content-type", "application/json")

		res, err := infra.client.Do(req)
		require.NoError(err, "perform client request")

		defer func() {
			err := res.Body.Close()
			require.NoError(err, "close response body")
		}()

		assert.Equal(http.StatusUnprocessableEntity, res.StatusCode, "incorrect response status code")

		body, err := ioutil.ReadAll(res.Body)
		require.NoError(err, "read response body")

		assert.Len(body, 0, "incorrect response body length")

		gotBankAccount, err := bankaccounts.FindByID(
			context.Background(), infra.db, expectedBankAccount.ID)
		require.NoError(err, "retrieve bank account")

		assert.Equal(expectedBankAccount, gotBankAccount, "bank account changed")

		transactionCount, err := transactions.Count(context.Background(), infra.db)
		require.NoError(err, "count transactions")

		assert.Zero(transactionCount, "transactions created")
	})

	t.Run("bank account has enough funds", func(t *testing.T) {
		defer truncateBulkTransactionTables(t, infra.db)

		expectedBankAccount, err := bankaccounts.Insert(
			context.Background(), infra.db, defaultBankAccount())
		require.NoError(err, "insertBankAccount")

		totalTransferCents := int64(6225150)

		expectedBankAccount.BalanceCents -= totalTransferCents

		// TODO: Embed fixtures
		fixturePath := filepath.Join("..", "fixtures", "requests", "201_created.json")
		fixtureBytes, err := ioutil.ReadFile(fixturePath)
		require.NoError(err, "read fixture file")

		req, err := http.NewRequest(http.MethodPost, bulkTransferURL(), bytes.NewReader(fixtureBytes))
		require.NoError(err, "create HTTP request")

		req.Header.Add("content-type", "application/json")

		res, err := infra.client.Do(req)
		require.NoError(err, "perform client request")

		defer func() {
			err := res.Body.Close()
			require.NoError(err, "close response body")
		}()

		assert.Equal(http.StatusCreated, res.StatusCode, "incorrect response status code")

		body, err := ioutil.ReadAll(res.Body)
		require.NoError(err, "read response body")

		assert.Len(body, 0, "incorrect response body length")

		gotBankAccount, err := bankaccounts.FindByID(
			context.Background(), infra.db, expectedBankAccount.ID)
		require.NoError(err, "retrieve bank account")

		assert.Equal(expectedBankAccount, gotBankAccount, "bank account changed")

		transactionCount, err := transactions.Count(context.Background(), infra.db)
		require.NoError(err, "count transactions")

		assert.EqualValues(3, transactionCount, "transactions created")

		expectedTransaction := service.Transaction{
			BankAccountID:    expectedBankAccount.ID,
			CounterpartyName: "Bip Bip",
			CounterpartyIBAN: "EE383680981021245685",
			CounterpartyBIC:  "CRLYFRPPTOU",
			Currency:         "EUR",
			AmountCents:      1450,
			Description:      "Wonderland/4410",
		}

		gotTransactions, err := transactions.SelectByCounterpartyName(
			context.Background(), infra.db, expectedTransaction.CounterpartyName)
		require.NoError(err, "select transactions")

		assert.Len(gotTransactions, 1,
			"count of transactions with counterparty_name %q", expectedTransaction.CounterpartyName)

		gotTransaction := gotTransactions[0]
		expectedTransaction.ID = gotTransaction.ID

		assert.Equal(expectedTransaction, gotTransaction, "unequal transactions")
	})
}

func defaultBankAccount() service.BankAccount {
	return service.BankAccount{
		OrganizationName: "ACME Corp",
		OrganizationBIC:  "OIVUSCLQXXX",
		OrganizationIBAN: "FR10474608000002006107XXXXX",
		BalanceCents:     10000000,
	}
}

func truncateBulkTransactionTables(t *testing.T, exec sql.Execer) {
	err := bankaccounts.Truncate(context.Background(), exec)
	assert.NoError(t, err, "truncateBankAccounts")

	err = transactions.Truncate(context.Background(), exec)
	assert.NoError(t, err, "truncateTransactions")
}
