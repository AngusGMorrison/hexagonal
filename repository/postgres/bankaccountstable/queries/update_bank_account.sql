UPDATE bank_accounts
SET
	organization_name = $2,
	iban = $3,
	bic = $4,
	balance_cents = $5,
WHERE id = $1
RETURNING *;
