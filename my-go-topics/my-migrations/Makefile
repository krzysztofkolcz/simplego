DB_URL=postgres://mymigrationsuser:pass@localhost:5432/db?sslmode=disable

migrate-public:
	migrate -path db/migrations/public -database $(DB_URL) up

migrate-tenant:
	migrate -path db/migrations/tenant -database "$(DB_URL)&search_path=$(SCHEMA)" up

integration-test:
	go test -tags=integration ./tests/integration/...

# go test -v ./tests/integration/...        # verbose (print test names)
# go test -count=1 ./tests/integration/...  # bez cache (ważne przy integracyjnych!)
# go test -run TestUser ./tests/integration/...  # tylko wybrane testy