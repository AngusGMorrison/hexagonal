# Advanced Hexagonal Architecture Demo

> ### Looking for a more recent and realistic example of Hexagonal Architecture in Go? Check out [my implementation of the RealWorld spec](https://github.com/AngusGMorrison/realworld-go).

## What is this?

An experiment in "extreme decoupling" that satisifies the following criteria:
* Business logic is fully and rigorously decoupled from transport and persistence layers;
* Business logic can perform atomic operations without directly manipulating a database or transaction object;
* Within the persistence layer, database tables are not only independent of domain models, but of the database itself. Tables have no concept of transactions or drivers<sup>1</sup>.

This work was inspired by a series of training workshops I created for Qonto, Europe's leading finance solution for freelancers and SMEs. It addresses the problem of how to cleanly separate domains in a mono- or macrolithic project where the database tables required by different domains may overlap and atomicity is essential.

This demo provides an HTTP server with one endpoint: `/enroll`, which receives requests to enroll students in a course identified by a unique code. The request must only succeed if the following criteria are met:
* The course exists in the database;
* At least one student is being enrolled;
* All of the students attempting to enroll in the course exist in the database;
* None of the students are already enrolled in the course;
* The course has sufficient capacity for all of the enrolling students.

If any of these conditions are violated, the server responds 422 Unprocessable Entity.

If the request is syntactically invalid, the server responds 400 Bad Request.

Otherwise, the students are enrolled in the course and the server responds 201 Created.

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

Run the migrations and seed the development database:
```bash
docker-compose run --rm hexagonal bash
make migrate
make migate_test
make seed
```

Run the server:
```bash
docker-compose up hexagonal
```

Requests can then be made to
```bash
POST localhost:3000/enroll
```

A Postman collection containing sample requests is provided in `Hexagonal.postman_collection.json`.

## Database

This demo uses the `hexagonal_development` database running locally on the PostgreSQL instance specified by docker-compose.yml.

The effect of bulk transfer requests on the database can be monitored using `psql`:
```bash
docker-compose exec postgres psql -U postgres hexagonal_development
```

### Migrations

After building the application, run migrations with `make migrate`. Use `make migrate_test` to migrate the test database.

Alternatively, the database to be migrated is given by the `DB_NAME` environment variable, which defaults to `hexagonal_development`. To migrate the test database using this method, run `DB_NAME=hexagonal_test make migrate`.

To seed the database, run `make seed`. The seeds to be loaded are found under `internal/storage/sql/seeds`.

### Schema
`courses` and `students` are joined in a many-to-many relationship by the `enrollments` table.

**courses**
* id BIGSERIAL PRIMARY KEY
* title VARCHAR
* code VARCHAR
* description TEXT
* capacity INT

**students**
* id BIGSERIAL PRIMARY KEY
* name VARCHAR
* birthdate DATE
* email VARCHAR

**enrollments**
* id INTEGER
* course_id BIGINT REFERENCES courses
* student_id BIGINT REFERENCES students

## Domain

Courses and students are aggregated under the `class` domain, which represents an association of one course with zero or more students.

Note that this business domain is entirely independent of its representation in the database. The business logic has no understanding of join tables or even of relational databases.

## Tests
Before running tests, create and migrate the test database:
```bash
docker-compose exec postgres psql -U postgres
```
```sql
CREATE DATABASE hexagonal_test;
```
```bash
make migrate_test
```

A full integration test suite can be found under `integration_test`. Run integration tests with `make integration_test`.

Example unit tests can be found for the `handler` package in `internal/handler/rest/enrollments_test.go`, and for the `classservice` package in `internal/service/classservice/enroll_test.go`. Run these using `make unit_test`.

## Notes
1. Although table code has no explicit dependency on any database or driver package, in practice I take advantage of PostgreSQL's ability to return the rows modified by a query without making a second query. This could be made truly driver-agnostic, but in typical business scenarios there is little need to. The key advantage of the proposed architecture is the separation of the table representation from the manner in which transactions are executed, i.e. with or without a transaction.