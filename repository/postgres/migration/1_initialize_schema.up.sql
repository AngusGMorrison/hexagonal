DROP TABLE IF EXISTS bank_accounts CASCADE;
DROP TABLE IF EXISTS transactions;

CREATE TABLE bank_accounts (
  id BIGSERIAL PRIMARY KEY,
  organization_name VARCHAR(255) NOT NULL,
  balance_cents BIGINT NOT NULL,
  iban VARCHAR(255) NOT NULL,
  bic VARCHAR(255) NOT NULL
);

CREATE UNIQUE INDEX iban_idx
ON bank_accounts (iban);

CREATE TABLE transactions (
  id BIGSERIAL PRIMARY KEY,
  bank_account_id BIGINT NOT NULL REFERENCES bank_accounts,
  counterparty_name VARCHAR(255) NOT NULL,
  counterparty_iban VARCHAR(255) NOT NULL,
  counterparty_bic VARCHAR(255) NOT NULL,
  amount_cents BIGINT NOT NULL,
  amount_currency VARCHAR(255) NOT NULL,
  description VARCHAR(255) NOT NULL
);
