package transferrepo

import "github.com/angusgmorrison/hexagonal/internal/controller"

// BankAccountRow represents a row of the bank_accounts table.
type BankAccountRow struct {
	ID               int64  `db:"id"`
	OrganizationName string `db:"organization_name"`
	BalanceCents     int64  `db:"balance_cents"`
	IBAN             string `db:"iban"`
	BIC              string `db:"bic"`
}

func bankAccountRowFromDomain(domainAccount controller.BankAccount) BankAccountRow {
	return BankAccountRow{
		ID:               domainAccount.ID,
		OrganizationName: domainAccount.OrganizationName,
		IBAN:             domainAccount.OrganizationIBAN,
		BIC:              domainAccount.OrganizationBIC,
		BalanceCents:     domainAccount.BalanceCents,
	}
}

func (ba BankAccountRow) toDomain() controller.BankAccount {
	return controller.BankAccount{
		ID:               ba.ID,
		OrganizationName: ba.OrganizationName,
		OrganizationIBAN: ba.IBAN,
		OrganizationBIC:  ba.BIC,
		BalanceCents:     ba.BalanceCents,
	}
}

// TransactionRows is a convenience wrapper around one or more instances of
// transactionRow.
type TransactionRows []TransactionRow

func transactionRowsFromDomain(domainTransfers []controller.CreditTransfer) TransactionRows {
	rows := make(TransactionRows, len(domainTransfers))

	for i, dt := range domainTransfers {
		rows[i] = transactionRowFromDomain(dt)
	}

	return rows
}

// TransactionRow represents a row of the transactions table.
type TransactionRow struct {
	ID               int64  `db:"id"`
	BankAccountID    int64  `db:"bank_account_id"`
	CounterpartyName string `db:"counterparty_name"`
	CounterpartyIBAN string `db:"counterparty_iban"`
	CounterpartyBIC  string `db:"counterparty_bic"`
	AmountCents      int64  `db:"amount_cents"`
	AmountCurrency   string `db:"amount_currency"`
	Description      string `db:"description"`
}

func transactionRowFromDomain(domainTransfer controller.CreditTransfer) TransactionRow {
	return TransactionRow{
		ID:               domainTransfer.ID,
		BankAccountID:    domainTransfer.BankAccountID,
		CounterpartyName: domainTransfer.CounterpartyName,
		CounterpartyBIC:  domainTransfer.CounterpartyBIC,
		CounterpartyIBAN: domainTransfer.CounterpartyIBAN,
		AmountCents:      domainTransfer.AmountCents,
		AmountCurrency:   domainTransfer.Currency,
		Description:      domainTransfer.Description,
	}
}
