INSERT INTO bank_accounts (organization_name, iban, bic, balance_cents)
VALUES ($1, $2, $3, $4)
RETURNING *;
