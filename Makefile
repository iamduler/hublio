include .env
export

MIGRATION_PATH=./internal/db/migrations
DATABASE_URL=postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)

ENV_FILE=.env
PROD_COMPOSE_FILE=docker-compose.prod.yml
NOAPP_COMPOSE_FILE=docker-compose.noapp.yml
DEV_COMPOSE_FILE=docker-compose.dev.yml

import_db:
	docker exec -it postgres psql -U $(DB_USER) -d $(DB_NAME) < ./db/schema.sql

export_db:
	docker exec -it postgres pg_dump -U $(DB_USER) -d $(DB_NAME) > ./db/schema.sql

server:
	go run cmd/api/main.go

migrate_create:
	migrate create -ext sql -dir $(MIGRATION_PATH) -seq $(name)

migrate_up:
	migrate -path $(MIGRATION_PATH) -database $(DATABASE_URL) up

migrate_down:
	migrate -path $(MIGRATION_PATH) -database $(DATABASE_URL) down

migrate_status:
	migrate -path $(MIGRATION_PATH) -database $(DATABASE_URL) status

migrate_force:
	migrate -path $(MIGRATION_PATH) -database $(DATABASE_URL) force $(version)

migrate_version:
	migrate -path $(MIGRATION_PATH) -database $(DATABASE_URL) version

migrate_reset:
	migrate -path $(MIGRATION_PATH) -database $(DATABASE_URL) reset

migrate_drop:
	migrate -path $(MIGRATION_PATH) -database $(DATABASE_URL) drop

migrate_goto:
	migrate -path $(MIGRATION_PATH) -database $(DATABASE_URL) goto $(version)
	# example: make migrate_goto version=1

sqlc:
	sqlc generate

build:
	go build -o bin/app.exe ./cmd/api

run:
	bin/app.exe

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
 
.PHONY: server migrate_create migrate_up migrate_down migrate_status migrate_force migrate_version migrate_reset 
.PHONY: migrate_drop migrate_goto sqlc build run prod stop_prod logs_prod dev noapp stop_noapp bash