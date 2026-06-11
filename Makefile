APP_NAME=postgresome

PRIMARY_DB_URL=postgres://postgres:postgres@localhost:55432/postgresome_test?sslmode=disable
SECONDARY_DB_URL=postgres://postgres:postgres@localhost:55433/postgresome_secondary?sslmode=disable


.PHONY: run-agent-primary
run-agent-primary:
	@echo "Starting agent against PRIMARY database..."
	DATABASE_URL="$(PRIMARY_DB_URL)" go run ./cmd/agent


.PHONY: run-agent-secondary
run-agent-secondary:
	@echo "Starting agent against SECONDARY database..."
	DATABASE_URL="$(SECONDARY_DB_URL)" go run ./cmd/agent


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


.PHONY: clean
clean:
	rm -rf bin