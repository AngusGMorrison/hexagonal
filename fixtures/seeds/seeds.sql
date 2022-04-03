INSERT INTO bank_accounts (organization_name, balance_cents, iban, bic)
VALUES
  ('ACME Corp', 10000000, 'FR10474608000002006107XXXXX', 'OIVUSCLQXXX');

INSERT INTO transactions (
  counterparty_name, counterparty_iban, counterparty_bic, amount_cents,
  amount_currency, bank_account_id, description
)
VALUES
  ('ACME Corp. Main Account', 'EE382200221020145685', 'CCOPFRPPXXX', 11000000, 'EUR', 1, 'Treasury income'),
  ('Bip Bip', 'EE383680981021245685', 'CRLYFRPPTOU', -1000000, 'EUR', 1, 'Bip Bip Salary');
