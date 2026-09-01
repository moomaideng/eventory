SHELL := /bin/sh
ENV_FILE ?= .env

.PHONY: dev down db backend backend-dev frontend migrate reset seed test build require-env

require-env:
	@test -f "$(ENV_FILE)" || { \
		echo "Missing $(ENV_FILE). Copy .env.example to .env first."; \
		exit 1; \
	}

define run_backend
	@set -a; . "./$(ENV_FILE)"; set +a; \
	export DB_DSN="host=localhost user=$$POSTGRES_USER password=$$POSTGRES_PASSWORD dbname=$$POSTGRES_DB port=$${POSTGRES_PORT:-5432} sslmode=disable"; \
	export PORT="$${BACKEND_PORT:-8080}"; \
	export ENVIRONMENT=development; \
	cd backend && $(1)
endef

define run_frontend
	@set -a; . "./$(ENV_FILE)"; set +a; \
	export NEXT_PUBLIC_SUPABASE_URL="$$SUPABASE_URL"; \
	export NEXT_PUBLIC_SUPABASE_PUBLISHABLE_KEY="$$SUPABASE_PUBLISHABLE_KEY"; \
	export NEXT_PUBLIC_API_URL="$${API_URL:-http://localhost:8080}"; \
	export INTERNAL_API_URL="$${API_URL:-http://localhost:8080}"; \
	cd frontend && $(1)
endef

# Run the complete local stack in containers.
dev: require-env
	docker compose up --build

down:
	docker compose down

# Run only PostgreSQL in Docker for host-based backend development.
db: require-env
	docker compose up -d postgres

# Run backend commands on the host with configuration from the root .env.
backend: require-env
	$(call run_backend,go run ./cmd/api)

backend-dev: require-env
	$(call run_backend,air)

migrate: require-env
	$(call run_backend,go run ./cmd/migrate)

reset: require-env
	$(call run_backend,go run ./cmd/migrate -reset)

seed: require-env
	$(call run_backend,go run ./cmd/seed)

# Next.js requires browser-visible variables to use the NEXT_PUBLIC_ prefix.
frontend: require-env
	$(call run_frontend,npm run dev)

test:
	cd backend && go test -v ./...
	cd frontend && npm run lint

build: require-env
	cd backend && go build -o bin/api ./cmd/api
	$(call run_frontend,npm run build)
