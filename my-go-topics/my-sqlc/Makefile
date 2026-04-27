run:
	go run ./cmd/api

sqlc:
	sqlc generate

migrate-up:
	migrate -path db/migrations/public -database $(DB_URL) up

test:
	go test ./...