package integration_test

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBulkTransfer(t *testing.T) {
	require := require.New(t)
	env := defaultEnvConfig()

	infra, err := newInfrastructure(env)
	require.NoError(err, "newInfrastructure")

	t.Cleanup(infra.mustCleanup)

	t.Run("bank account has insufficient funds", func(t *testing.T) {
		bankAccount, err := infra.repo.insertBankAccount(defaultBankAccount())
		require.NoError(err, "insertBankAccount")

		t.Log(bankAccount.ID)

		defer func() {
			err := infra.repo.truncateBankAccounts()
			require.NoError(err, "truncateBankAccounts")
		}()
	})
}
