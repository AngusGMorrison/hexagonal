.PHONY: build_migrate migrate rollback build_seed seed build_server run migrate_test unit_test integration_test

build_migrate:
	CGO_ENABLED=0 go build -o ./bin/migrate ./cmd/migrate

migrate: build_migrate
	bin/migrate

rollback: build_migrate
	bin/migrate -rollback

build_seed:
	CGO_ENABLED=0 go build -o ./bin/seed ./cmd/seed

seed: build_seed
	bin/seed

build_server:
	CGO_ENABLED=0 go build -o ./bin/server ./cmd/server

run: build_server migrate 
	bin/server

migrate_test: build_migrate
	DB_NAME=hexagonal_test bin/migrate

unit_test:
	go test -race ./...

integration_test:
	go test -race -tags=integration ./...
