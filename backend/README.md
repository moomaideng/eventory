# Eventory Backend

This directory contains the Go backend service. It is built utilizing a structured Ports & Adapters architecture to separate core business logic from frameworks, external APIs, and database implementations.

## Technology Stack

*   **Language:** Go 1.26+
*   **API Framework:** Huma v2 (Automated OpenAPI 3.1 documentation)
*   **Router & Middleware:** Chi Router with Logger, Recoverer, and CORS handler
*   **Database & ORM:** PostgreSQL with GORM
*   **Authentication:** Supabase Auth verification via JWKS (ES256 / RS256)
*   **Configuration:** Viper
*   **Live Reload:** Air

## Repository Architecture

```text
.
├── backend/
│   ├── cmd/
│   │   └── api/
│   │       └── main.go         # Application entry point, Chi router, and Huma wiring
│   ├── internal/               # Private application code
│   │   ├── handlers/           # Huma HTTP handlers and DTOs (/me, /onboard, /{id})
│   │   ├── middlewares/        # HTTP middlewares (Supabase JWKS Auth, Dev guard)
│   │   ├── models/             # Domain and Database models (GORM tags)
│   │   ├── repositories/       # Data access layer (GORM queries)
│   │   ├── seeds/              # Seed scripts for development mock records
│   │   ├── services/           # External service adapters
│   │   └── usecases/           # Core business logic and unit tests
│   ├── pkg/                    # Public/Shared utilities
│   │   ├── config/             # Viper configuration loading
│   │   └── database/           # PostgreSQL connection initialization
│   ├── .air.toml               # Air configuration for live reloading
│   ├── Dockerfile              # Container build for production
│   └── README.md               # Backend-specific documentation
```

## Prerequisites

Before running the server, ensure the following dependencies are installed:

* **Go:** Version 1.26 or higher.
* **Docker & Docker Compose:** For running the local PostgreSQL container.
* **Air:** For live reloading (`go install github.com/air-verse/air@v1.67.3`).
* **Make (Optional):** If you prefer running convenience shortcuts from the repository root on macOS/Linux.

> **Windows Note:** Ensure your Go binary installation path (typically `%USERPROFILE%\go\bin`) is added to your user `PATH` environment variable so that you can invoke `air` directly in PowerShell or Command Prompt.

---

## Getting Started (Native Development)

You can run the backend service entirely natively from within this `backend/` directory without using Make or root scripts.

### 1. Configure Environment

Copy the local environment template inside this directory (or duplicate it manually):

```bash
cp .env.example .env
```

> **PostgreSQL Port Note:** If you already run PostgreSQL natively on your machine (e.g. as a Windows service), port `5432` may already be occupied. In that case, modify `POSTGRES_PORT` in `.env` (e.g. `POSTGRES_PORT=5433`) and update the port in `DB_DSN` accordingly.

### 2. Start PostgreSQL

Launch the local PostgreSQL container in the background:

```bash
docker compose -f ../docker-compose.yml --env-file .env up -d --wait postgres
```

### 3. Run Database Migrations

Apply GORM auto-migrations to build or rebuild your database schema:

```bash
# Recommended for dev: wipes & rebuilds schema fresh (guarantees 100% sync with Go models)
go run ./cmd/migrate --reset

# Or omit --reset if you want to preserve existing database data:
# go run ./cmd/migrate
```

### 4. Seed Mock Data

Populate the database with initial development records (accounts, tournaments):

```bash
go run ./cmd/seed
```

### 5. Start the Server

- **With live-reload (Air):**
  ```bash
  air -c .air.toml
  ```
- **Or standard Go compile & run:**
  ```bash
  go run ./cmd/api
  ```

The API will start at `http://localhost:8080`.
- Interactive OpenAPI 3.1 documentation: `http://localhost:8080/docs`
- Raw OpenAPI schema: `http://localhost:8080/openapi.json`
- Health check: `http://localhost:8080/health`

### 6. Run Unit Tests

```bash
go test -v ./...
```

---

## Alternative: Root Make Commands (macOS / Linux)

If you are on macOS, Linux, or WSL and prefer orchestrating from the repository root using Make:

```bash
make db             # Start PostgreSQL
make migrate        # Run migrations
make seed           # Seed data
make backend        # Run backend with Air
make test           # Run backend tests & frontend checks
```

Viper reads `backend/.env` when running locally, while actual process environment variables take precedence. Native commands use the file's `localhost` `DB_DSN`; Compose overrides it with the Docker-local `postgres` hostname. Production continues receiving its Supabase `DB_DSN` from the deployment environment.
