include .env
export

# Ensure tools installed via `go install` are on PATH (migrate, sqlc, …).
export PATH := $(PATH):$(shell go env GOPATH)/bin

MIGRATION_PATH=./migrations
DATABASE_URL=postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)

MIGRATE ?= $(shell command -v migrate 2>/dev/null || echo "$(shell go env GOPATH)/bin/migrate")

ENV_FILE=.env
PROD_COMPOSE_FILE=docker-compose.prod.yml
NOAPP_COMPOSE_FILE=docker-compose.noapp.yml
DEV_COMPOSE_FILE=docker-compose.dev.yml

server:
	go run ./cmd/api

worker:
	go run ./cmd/worker

migrate_create:
	$(MIGRATE) create -ext sql -dir $(MIGRATION_PATH) -seq $(name)

migrate_up:
	$(MIGRATE) -path $(MIGRATION_PATH) -database "$(DATABASE_URL)" up

migrate_down:
	$(MIGRATE) -path $(MIGRATION_PATH) -database "$(DATABASE_URL)" down

migrate_status:
	$(MIGRATE) -path $(MIGRATION_PATH) -database "$(DATABASE_URL)" version

migrate_force:
	$(MIGRATE) -path $(MIGRATION_PATH) -database "$(DATABASE_URL)" force $(version)

migrate_version:
	$(MIGRATE) -path $(MIGRATION_PATH) -database "$(DATABASE_URL)" version

migrate_goto:
	$(MIGRATE) -path $(MIGRATION_PATH) -database "$(DATABASE_URL)" goto $(version)

install_tools:
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

db_create:
	bash ./scripts/create-db.sh

db_setup: db_create migrate_up

sqlc:
	sqlc generate

test:
	go test ./...

vet:
	go vet ./...

check: vet test build

build:
	go build -o bin/api ./cmd/api
	go build -o bin/worker ./cmd/worker

enqueue_health:
	@curl -sS -X POST -H "X-API-KEY: $${API_KEY}" http://localhost:$${SERVER_PORT:-8080}/api/v1/platform/queue/health

# Production
prod:
	docker compose -f $(PROD_COMPOSE_FILE) down
	docker compose -f $(PROD_COMPOSE_FILE) --env-file $(ENV_FILE) up -d --build

stop_prod:
	docker compose -f $(PROD_COMPOSE_FILE) down

logs_prod:
	docker compose -f $(PROD_COMPOSE_FILE) logs -f --tail 100

# Dev
dev:
	docker compose -f $(DEV_COMPOSE_FILE) down
	docker compose -f $(DEV_COMPOSE_FILE) --env-file $(ENV_FILE) up --build

# No app
noapp:
	docker compose -f $(NOAPP_COMPOSE_FILE) down
	docker compose -f $(NOAPP_COMPOSE_FILE) --env-file $(ENV_FILE) up -d --build

stop_noapp:
	docker compose -f $(NOAPP_COMPOSE_FILE) down

bash:
	docker exec -it go-api /bin/sh

.PHONY: server worker migrate_create migrate_up migrate_down migrate_status migrate_force migrate_version
.PHONY: migrate_goto install_tools db_create db_setup sqlc test vet check build enqueue_health
.PHONY: prod stop_prod logs_prod dev noapp stop_noapp bash
