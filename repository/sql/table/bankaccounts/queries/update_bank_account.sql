UPDATE bank_accounts
SET
	organization_name = :organization_name,
	iban = :iban,
	bic = :bic,
	balance_cents = :balance_cents
WHERE id = :id
RETURNING *;
