.PHONY: install dev dev-windows dev-docker down db backend frontend backend-native frontend-native backend-docker frontend-docker frontend-deps migrate reset seed migrate-docker reset-docker seed-docker test build

# Compose uses each application's local environment file. Later files only
# affect Compose interpolation; each service also declares its own env_file.
COMPOSE = docker compose --env-file backend/.env --env-file frontend/.env.local

ifeq ($(OS),Windows_NT)
DEV_RUNNER = powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/dev.ps1
else
DEV_RUNNER = bash ./scripts/dev.sh
endif

# Install host-side project dependencies and the pinned live-reload tool.
install:
	npm --prefix frontend ci
	go -C backend mod download
	go install github.com/air-verse/air@v1.67.3

# Fast local development: PostgreSQL in Docker, application processes on host.
dev:
	$(DEV_RUNNER)

# Equivalent native workflow for GNU Make users running Windows PowerShell.
dev-windows:
	powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/dev.ps1

# Complete containerized stack for parity checks and Docker-only development.
dev-docker:
	$(COMPOSE) up --build

down:
	$(COMPOSE) down

db:
	$(COMPOSE) up -d --wait postgres

backend:
	cd backend && air -c .air.toml

frontend:
	npm --prefix frontend run dev

backend-native: backend

frontend-native: frontend

backend-docker:
	$(COMPOSE) up --build backend

frontend-docker:
	$(COMPOSE) up --build frontend

frontend-deps:
	$(COMPOSE) stop frontend
	$(COMPOSE) run --rm --no-deps frontend npm ci
	$(COMPOSE) start frontend

migrate:
	go -C backend run ./cmd/migrate

reset:
	go -C backend run ./cmd/migrate -reset

seed:
	go -C backend run ./cmd/seed

migrate-docker:
	$(COMPOSE) run --rm --build migrate

reset-docker:
	$(COMPOSE) run --rm --build migrate ./migrate -reset

seed-docker:
	$(COMPOSE) run --rm --build migrate ./seed

test:
	go -C backend test -v ./...
	npm --prefix frontend run lint
	npm --prefix frontend run typecheck

build:
	$(COMPOSE) build
