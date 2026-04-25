.PHONY: codegen sqlc-gen build run lint test docker-up docker-down

codegen:
	oapi-codegen --config=./apis/todo/config.yaml ./apis/todo/todo.yaml

sqlc-gen:
	sqlc generate

build:
	go build -o bin/todo-server ./cmd/app/...

run: build
	./bin/todo-server

lint:
	golangci-lint run ./...

test:
	go test ./... -v -race

docker-up:
	docker run --name todo-postgres \
		-e POSTGRES_USER=postgres \
		-e POSTGRES_PASSWORD=postgres \
		-e POSTGRES_DB=todos \
		-p 5432:5432 -d postgres:16

docker-down:
	docker stop todo-postgres && docker rm todo-postgres
