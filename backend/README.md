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

Before running the server, ensure the following dependencies are installed on your local machine:

* **Go:** Version 1.26 or higher.
* **Docker & Docker Compose:** For running the local PostgreSQL instance.
* **Make:** Optional convenience commands from the repository root.
* **Air:** Required for native backend live reload and installed by `make install`. The Docker development image also includes it.

## Getting Started

Backend development is orchestrated from the repository root. The backend owns its environment file, which is used by both native Go commands and Docker Compose.

1. **Configure the backend environment:**
   ```bash
   cd ..
   cp backend/.env.example backend/.env
   ```

2. **Initialize Infrastructure:**
   Start the PostgreSQL database in the background.
   ```bash
   make db
   ```

3. **Run Database Migrations:**
   Execute GORM auto-migrations to build the database schema based on your current models.
   ```bash
   make migrate
   # or make reset (if db schema needs a clean reset)
   ```

4. **Seed the Database:**
   Populate the database with mock records for local development.
   ```bash
   make seed
   ```

   This runs the current Go seed code natively using `backend/.env`. Use `make seed-docker` when you explicitly want the Docker version.

5. **Start the Application:**
   Run the normal hybrid development workflow (native backend and frontend with PostgreSQL in Docker):
   ```bash
   make dev
   ```

   To run only the backend natively after PostgreSQL is ready, use `make backend`. To run it in Docker, use `make backend-docker`.

The API will start at `http://localhost:8080`. Interactive documentation (OpenAPI 3.1) is automatically generated and accessible at `http://localhost:8080/docs` (with raw schema at `http://localhost:8080/openapi.json`).

Viper reads `backend/.env` when running locally, while actual process environment variables take precedence. Native commands use the file's `localhost` `DB_DSN`; Compose overrides it with the Docker-local `postgres` hostname. Production continues receiving its Supabase `DB_DSN` from the deployment environment.
