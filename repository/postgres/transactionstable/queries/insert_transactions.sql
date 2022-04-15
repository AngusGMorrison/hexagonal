INSERT INTO transactions (
	counterparty_name, counterparty_iban, counterparty_bic, amount_cents,
	amount_currency, bank_account_id, description
) VALUES (
	$1, $2, $3, $4, $5, $6, $7
)
RETURNING *;
