# Hexagonal Architecture Demo

A demonstration of the hexagonal architecture pattern in Go, forming part of a series of training workshops I created for Qonto, Europe's leading finance solution for freelancers and SMEs.

This demo provides an HTTP server with one endpoint: `/bulk_transfers`, which receives requests to apply multiple transfers to a single bank account identified by its IBAN. The request must only succeed if the bank account has sufficient funds to settle all the transfers in the bulk transfer. If so, the server persists all the transfers and the updated bank balance to a Postgres database and responds 201 Created. Otherwise, the server commits nothing and responds 422 Unprocessable Entity.

## Running the demo

This project uses docker-compose to run both the `hexagonal` application and a PostgreSQL server.

Before running the application, create the databases `hexagonal_development` and `hexagonal_test` using `psql`:
```bash
docker-compose up -d postgres
docker-compose exec postgres psql -U postgres
```
```sql
CREATE DATABASE hexagonal_development;
CREATE DATABASE hexagonal_test;
```

```bash
docker-compose up hexagonal
```

Requests can then be made to
```bash
POST localhost:3000/bulk_transfer
```

Sample payloads and a Postman collection are provided in `/fixtures/requests`.

## Database

This demo uses the `hexagonal_development` database running locally on the PostgreSQL instance specified by docker-compose.yml.

The effect of bulk transfer requests on the database can be monitored using `psql`:
```bash
docker-compose exec postgres psql -U postgres hexagonal_development
```

### Migrations

After building the application, run migrations with `make migrate`.

The database to be migrated is named by the `DB_NAME` environment variable, which defaults to `hexagonal_development`. To migrate the test database, run `DB_NAME=hexagonal_test make migrate`.

To seed the database, run `make seed`. The seeds to be loaded are found under `fixtures/seeds`.

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
