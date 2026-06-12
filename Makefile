APP_NAME=postgresome

PRIMARY_DB_URL=postgres://postgres:postgres@localhost:55432/postgresome_test?sslmode=disable
SECONDARY_DB_URL=postgres://postgres:postgres@localhost:55433/postgresome_secondary?sslmode=disable
POSTGRESOME_DB_URL=postgres://postgres:postgres@localhost:55432/postgresome_app?sslmode=disable
POSTGRESOME_API_URL=http://localhost:9090


.PHONY: run-agent-primary
run-agent-primary:
	@echo "Starting agent against PRIMARY database..."
	DATABASE_URL="$(PRIMARY_DB_URL)" \
	AGENT_ID="agent-primary" \
	AGENT_NAME="Primary Agent" \
	AGENT_ENVIRONMENT="development" \
	POSTGRESOME_API_URL="$(POSTGRESOME_API_URL)" \
	DATABASE_INSTANCE_ID="db-primary" \
	go run ./cmd/agent


.PHONY: run-agent-secondary
run-agent-secondary:
	@echo "Starting agent against SECONDARY database..."
	DATABASE_URL="$(SECONDARY_DB_URL)" \
	AGENT_ID="agent-secondary" \
	AGENT_NAME="Secondary Agent" \
	AGENT_ENVIRONMENT="development" \
	POSTGRESOME_API_URL="$(POSTGRESOME_API_URL)" \
	DATABASE_INSTANCE_ID="db-secondary" \
	go run ./cmd/agent


.PHONY: run-api
run-api:
	@echo "Starting API server..."
	POSTGRESOME_DATABASE_URL="$(POSTGRESOME_DB_URL)" go run ./cmd/api


.PHONY: migrate
migrate:
	@echo "Running database migrations..."
	POSTGRESOME_DATABASE_URL="$(POSTGRESOME_DB_URL)" go run ./cmd/migrate


.PHONY: postgres-up
postgres-up:
	docker compose up -d


.PHONY: postgres-down
postgres-down:
	docker compose down


.PHONY: postgres-reset
postgres-reset:
	docker compose down -v
	docker compose up -d


.PHONY: tidy
tidy:
	go mod tidy


.PHONY: test
test:
	go test ./...


.PHONY: build-agent
build-agent:
	go build -o bin/agent ./cmd/agent


.PHONY: build-api
build-api:
	go build -o bin/api ./cmd/api


.PHONY: clean
clean:
	rm -rf bin