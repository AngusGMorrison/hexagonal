package transferrepo

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
)

const (
	_getBankAccountByIBANQueryFilename = "get_bank_account_by_iban.sql"
	_updateBankAccountQueryFilename    = "update_bank_account.sql"
	_insertTransactionsQueryFilename   = "insert_transactions.sql"
)

type queries map[string]string

func loadQueries(appRoot string) (queries, error) {
	absDir := absQueryDir(appRoot)

	filenames := queryFilenames()

	qs := make(queries, len(filenames))

	for _, filename := range filenames {
		path := filepath.Join(absDir, filename)

		queryBytes, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", path, err)
		}

		qs[filename] = string(queryBytes)
	}

	return qs, nil
}

func (q queries) getBankAccountByIBAN() string {
	return q[_getBankAccountByIBANQueryFilename]
}

func (q queries) updateBankAccount() string {
	return q[_updateBankAccountQueryFilename]
}

func (q queries) insertTransactions() string {
	return q[_insertTransactionsQueryFilename]
}

func absQueryDir(appRoot string) string {
	return filepath.Join(
		appRoot, "internal", "adapter", "repository", "postgres", "transferrepo", "sql")
}

func queryFilenames() []string {
	return []string{
		_getBankAccountByIBANQueryFilename,
		_updateBankAccountQueryFilename,
		_insertTransactionsQueryFilename,
	}
}
