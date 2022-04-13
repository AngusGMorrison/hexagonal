UPDATE bank_accounts
SET
	balance_cents = :balance_cents, organization_name = :organization_name,
	iban = :iban, bic = :bic
WHERE id = :id
RETURNING *;
