.PHONY: install dev dev-windows dev-docker down db backend frontend backend-native frontend-native frontend-deps migrate reset seed test build

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
	$(COMPOSE) up --build backend

frontend:
	$(COMPOSE) up --build frontend

backend-native:
	cd backend && air -c .air.toml

frontend-native:
	npm --prefix frontend run dev

frontend-deps:
	$(COMPOSE) stop frontend
	$(COMPOSE) run --rm --no-deps frontend npm ci
	$(COMPOSE) start frontend

migrate:
	$(COMPOSE) run --rm migrate

reset:
	$(COMPOSE) run --rm migrate ./migrate -reset

seed:
	$(COMPOSE) run --rm migrate ./seed

test:
	go -C backend test -v ./...
	npm --prefix frontend run lint
	npm --prefix frontend run typecheck

build:
	$(COMPOSE) build
