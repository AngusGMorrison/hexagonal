.PHONY: build test run migrate

build:
	CGO_ENABLED=0 go build -o ./bin/migrate ./cmd/migrate
	CGO_ENABLED=0 go build -o ./bin/server ./cmd/server

test:
	go test -race ./...

run:
	bin/server

migrate:
	bin/migrate

rollback:
	bin/migrate -rollback
