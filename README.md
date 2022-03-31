# Hexagonal Architecture Demo

A demonstration of the hexagonal architecture pattern in Go, which developed as part of a series of training workshops I created for Qonto, Europe's leading finance solution for freelancers and SMEs.

This demo provides an HTTP server with one endpoint: `/bulk_transfers`, which receives requests to apply multiple transfers to a single bank account identified by its IBAN. The request must only succeed if the bank account has sufficient funds to settle all the transfers in the bulk transfer. If so, the server persists the balance change and all the transfers to a Postgres database and responds 201 Created. Otherwise, the server commits nothing and responds 422 Unprocessable Entity.

## Running the demo

This project uses docker-compose to run both the `hexagonal` application and a PostgreSQL server.

```bash
docker-compose up hexagonal
```

Requests can then be made to
```bash
POST localhost:3000/bulk_transfer
```

Sample payloads and a Postman collection are provided in `/fixtures`.

## Database

This demo uses the `hexagonal_development` database running locally on the PostgreSQL instance specified by docker-compose.yml.

The effect of bulk transfer requests on the database can be monitored using `psql`:
```bash
docker-compose exec postgres psql -U postgres hexagonal_development
```

### Migrations

After building the application, run migrations with `make migrate`.

For the sake of simplicity, seed data is included in the up migration. To restore the database to its original state, roll back to truncate the tables and then migrate up again.

Naturally this is not recommended for production applications.

### Schema
**bank_accounts**
* id INTEGER
* organization_name TEXT
* balance_cents INTEGER
* iban TEXT
* bic TEXT

**transactions**
* id INTEGER
* counterparty_name TEXT
* counterparty_iban TEXT
* counterparty_bic TEXT
* amount_cents INTEGER
* amount_currency TEXT
* bank_account_id INTEGER FOREIGN KEY
* description TEXT
