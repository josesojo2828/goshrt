.PHONY: build test migrate-up migrate-down docker-build docker-up lint clean

# ─── Build ──────────────────────────────────────────────────────────────────
build:
	go build -o bin/goshrt ./cmd/goshrt

# ─── Test ───────────────────────────────────────────────────────────────────
test:
	go test ./... -v -race -count=1

test-coverage:
	go test ./... -coverprofile=coverage.out -count=1
	go tool cover -html=coverage.out -o coverage.html

# ─── Migrations ─────────────────────────────────────────────────────────────
MIGRATIONS_PATH = internal/store/postgres/migrations
POSTGRES_DSN ?= postgres://goshrt:goshrt@localhost:5432/goshrt?sslmode=disable

migrate-up:
	migrate -path $(MIGRATIONS_PATH) -database "$(POSTGRES_DSN)" up

migrate-down:
	migrate -path $(MIGRATIONS_PATH) -database "$(POSTGRES_DSN)" down

migrate-create:
	@read -p "Migration name: " name; \
	migrate create -ext sql -dir $(MIGRATIONS_PATH) -seq $$name

# ─── Docker ─────────────────────────────────────────────────────────────────
docker-build:
	docker compose build

docker-up:
	docker compose up --build -d

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f

# ─── Lint ───────────────────────────────────────────────────────────────────
lint:
	golangci-lint run ./...

# ─── Clean ──────────────────────────────────────────────────────────────────
clean:
	rm -rf bin/ coverage.out coverage.html
