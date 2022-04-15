SELECT id, organization_name, balance_cents, iban, bic
FROM bank_accounts
WHERE iban = $1;
