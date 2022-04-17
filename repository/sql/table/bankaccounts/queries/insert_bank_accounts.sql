INSERT INTO bank_accounts (organization_name, iban, bic, balance_cents)
VALUES (:organization_name, :iban, :bic, :balance_cents)
RETURNING *;
