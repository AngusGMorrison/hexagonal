SELECT id, bank_account_id, counterparty_name, counterparty_iban,
  counterparty_bic, amount_cents, amount_currency, description
FROM transactions
WHERE counterparty_name = $1;