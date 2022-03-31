package transferrepo

import "github.com/angusgmorrison/hexagonal/internal/app/transferdomain"

// bankAccountRow represents a row of the bank_accounts table.
type bankAccountRow struct {
	ID               int64  `db:"id"`
	OrganizationName string `db:"organization_name"`
	BalanceCents     int64  `db:"balance_cents"`
	IBAN             string `db:"iban"`
	BIC              string `db:"bic"`
}

func bankAccountRowFromDomain(domainAccount transferdomain.BankAccount) bankAccountRow {
	return bankAccountRow{
		ID:               domainAccount.ID,
		OrganizationName: domainAccount.OrganizationName,
		IBAN:             domainAccount.OrganizationIBAN,
		BIC:              domainAccount.OrganizationBIC,
		BalanceCents:     domainAccount.BalanceCents,
	}
}

func (ba bankAccountRow) toDomain() transferdomain.BankAccount {
	return transferdomain.BankAccount{
		ID:               ba.ID,
		OrganizationName: ba.OrganizationName,
		OrganizationIBAN: ba.IBAN,
		OrganizationBIC:  ba.BIC,
		BalanceCents:     ba.BalanceCents,
	}
}

// transactionRows is a convenience wrapper around one or more instances of
// transactionRow.
type transactionRows []transactionRow

func transactionRowsFromDomain(domainTransfers []transferdomain.CreditTransfer) transactionRows {
	rows := make(transactionRows, len(domainTransfers))

	for i, dt := range domainTransfers {
		rows[i] = transactionRowFromDomain(dt)
	}

	return rows
}

func (trs transactionRows) toDomain() []transferdomain.CreditTransfer {
	creditTransfers := make([]transferdomain.CreditTransfer, len(trs))

	for i, tr := range trs {
		creditTransfers[i] = tr.toDomain()
	}

	return creditTransfers
}

// transactionRow represents a row of the transactions table.
type transactionRow struct {
	ID               int64  `db:"id"`
	BankAccountID    int64  `db:"bank_account_id"`
	CounterpartyName string `db:"counterparty_name"`
	CounterpartyIBAN string `db:"counterparty_iban"`
	CounterpartyBIC  string `db:"counterparty_bic"`
	AmountCents      int64  `db:"amount_cents"`
	AmountCurrency   string `db:"amount_currency"`
	Description      string `db:"description"`
}

func transactionRowFromDomain(domainTransfer transferdomain.CreditTransfer) transactionRow {
	return transactionRow{
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

func (tr transactionRow) toDomain() transferdomain.CreditTransfer {
	return transferdomain.CreditTransfer{
		ID:               tr.ID,
		BankAccountID:    tr.BankAccountID,
		CounterpartyName: tr.CounterpartyName,
		CounterpartyBIC:  tr.CounterpartyBIC,
		CounterpartyIBAN: tr.CounterpartyIBAN,
		AmountCents:      tr.AmountCents,
		Currency:         tr.AmountCurrency,
		Description:      tr.Description,
	}
}
