INSERT INTO transactions (
	counterparty_name, counterparty_iban, counterparty_bic, amount_cents,
	amount_currency, bank_account_id, description
) VALUES (
	:counterparty_name, :counterparty_iban, :counterparty_bic, :amount_cents,
	:amount_currency, :bank_account_id, :description
) RETURNING *;
