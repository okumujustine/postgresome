APP_NAME=postgresome

PRIMARY_DB_URL=postgres://postgres:postgres@localhost:55432/postgresome_test?sslmode=disable
SECONDARY_DB_URL=postgres://postgres:postgres@localhost:55433/postgresome_secondary?sslmode=disable
POSTGRESOME_DB_URL=postgres://postgres:postgres@localhost:55434/postgresome_app?sslmode=disable
POSTGRESOME_API_URL=http://localhost:9090

.PHONY: run-api
run-api:
	@echo "Starting API server..."
	POSTGRESOME_DATABASE_URL="$(POSTGRESOME_DB_URL)" go run ./cloud/api


.PHONY: run-frontend
run-frontend:
	@echo "Starting frontend dev server..."
	cd frontend && npm run dev


.PHONY: migrate
migrate:
	@echo "Running database migrations..."
	POSTGRESOME_DATABASE_URL="$(POSTGRESOME_DB_URL)" go run ./cloud/migrate


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


.PHONY: docker-build
docker-build:
	docker compose -f docker-compose.app.yml build


.PHONY: docker-up
docker-up:
	docker compose -f docker-compose.app.yml up -d --build


.PHONY: dev-reset
dev-reset:
	docker compose -f docker-compose.app.yml down -v
	docker compose -f docker-compose.app.yml up -d --build


.PHONY: docker-down
docker-down:
	docker compose -f docker-compose.app.yml down


.PHONY: docker-logs
docker-logs:
	docker compose -f docker-compose.app.yml logs -f


.PHONY: tidy
tidy:
	go mod tidy


.PHONY: test
test:
	go test ./...


.PHONY: build-api
build-api:
	go build -o bin/api ./cloud/api


.PHONY: clean
clean:
	rm -rf bin
