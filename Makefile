APP_NAME=postgresome

PRIMARY_DB_URL=postgres://postgres:postgres@localhost:55432/postgresome_test?sslmode=disable
SECONDARY_DB_URL=postgres://postgres:postgres@localhost:55433/postgresome_secondary?sslmode=disable
POSTGRESOME_DB_URL=postgres://postgres:postgres@localhost:55434/postgresome_app?sslmode=disable
POSTGRESOME_API_URL=http://localhost:9090
POSTGRESOME_SECRET_KEY=postgresome-dev-secret-key

.PHONY: run-api
run-api:
	@echo "Starting API server..."
	cd backend && POSTGRESOME_DATABASE_URL="$(POSTGRESOME_DB_URL)" POSTGRESOME_SECRET_KEY="$(POSTGRESOME_SECRET_KEY)" go run ./api


.PHONY: run-frontend
run-frontend:
	@echo "Starting frontend dev server..."
	cd frontend && npm run dev


.PHONY: migrate
migrate:
	@echo "Running database migrations..."
	cd backend && POSTGRESOME_DATABASE_URL="$(POSTGRESOME_DB_URL)" go run ./migrate


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


.PHONY: docker-frontend-reset
docker-frontend-reset:
	docker compose -f docker-compose.app.yml rm -sf frontend
	docker compose -f docker-compose.app.yml up -d --build frontend


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
	cd backend && go mod tidy


.PHONY: test
test:
	cd backend && go test ./...


.PHONY: build-api
build-api:
	cd backend && go build -o bin/api ./api


.PHONY: clean
clean:
	rm -rf backend/bin
