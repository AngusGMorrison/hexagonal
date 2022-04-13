SELECT id, organization_name, iban, bic, balance_cents
FROM bank_accounts
WHERE id = $1;
