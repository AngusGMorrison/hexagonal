.PHONY: build_migrate migrate build_server run test rollback

build_migrate:
	CGO_ENABLED=0 go build -o ./bin/migrate ./cmd/migrate

migrate: build_migrate
	bin/migrate

build_server:
	CGO_ENABLED=0 go build -o ./bin/server ./cmd/server

run: build_server migrate 
	bin/server

test:
	go test -race ./...

rollback:
	bin/migrate -rollback
