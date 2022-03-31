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

-- For this demo, seeds are included in the migration for simplicity. In a real
-- application, seeds should be loaded from fixtures and migrations should be
-- reserved for database schema changes.

INSERT INTO bank_accounts (
  organization_name, balance_cents, iban, bic
)
VALUES
  ('ACME Corp', 10000000, 'FR10474608000002006107XXXXX', 'OIVUSCLQXXX');

INSERT INTO transactions (
  counterparty_name, counterparty_iban, counterparty_bic, amount_cents, amount_currency, bank_account_id, description
)
VALUES
  ('ACME Corp. Main Account', 'EE382200221020145685', 'CCOPFRPPXXX', 11000000, 'EUR', 1, 'Treasury income'),
  ('Bip Bip', 'EE383680981021245685', 'CRLYFRPPTOU', -1000000, 'EUR', 1, 'Bip Bip Salary');
