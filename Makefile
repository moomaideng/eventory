.PHONY: dev down db backend frontend migrate reset seed test build

# Cross-platform aliases. Docker Compose reads the root .env directly.
dev:
	docker compose up --build

down:
	docker compose down

db:
	docker compose up -d postgres

backend:
	docker compose up --build backend

frontend:
	docker compose up --build frontend

migrate:
	docker compose run --rm migrate

reset:
	docker compose run --rm migrate ./migrate -reset

seed:
	docker compose run --rm backend ./seed

test:
	go -C backend test -v ./...
	npm --prefix frontend run lint

build:
	docker compose build
