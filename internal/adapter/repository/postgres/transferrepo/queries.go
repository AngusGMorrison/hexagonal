package transferrepo

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

var _queryDir = filepath.Join(
	"internal", "adapter", "repository", "postgres", "transferrepo", "sql")

const (
	_getBankAccountByIBANQueryFilename = "get_bank_account_by_iban.sql"
	_updateBankAccountQueryFilename    = "update_bank_account.sql"
	_insertTransactionsQueryFilename   = "insert_transactions.sql"
)

type queries map[string]string

func loadQueries() (queries, error) {
	absoluteDir, err := absQueryDir()
	if err != nil {
		return nil, err
	}

	filenames := queryFilenames()

	qs := make(queries, len(filenames))

	for _, filename := range filenames {
		path := filepath.Join(absoluteDir, filename)

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

func absQueryDir() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("os.Getwd: %w", err)
	}

	return filepath.Join(cwd, _queryDir), nil
}

func queryFilenames() []string {
	return []string{
		_getBankAccountByIBANQueryFilename,
		_updateBankAccountQueryFilename,
		_insertTransactionsQueryFilename,
	}
}
