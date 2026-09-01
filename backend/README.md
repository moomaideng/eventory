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
* **Make:** For executing automation commands.
* **Air (Optional on the host):** The Docker development image already includes Air for automatic server reloading.

## Getting Started

Backend development is orchestrated from the repository root so it shares the same local environment as the frontend and Docker Compose.

1. **Configure the root environment:**
   ```bash
   cd ..
   cp .env.example .env
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

5. **Start the Application:**
   Run the backend and its dependencies in Docker. Air automatically rebuilds and restarts the API when Go source files change:
   ```bash
   make backend
   ```

The API will start at `http://localhost:8080`. Interactive documentation (OpenAPI 3.1) is automatically generated and accessible at `http://localhost:8080/docs` (with raw schema at `http://localhost:8080/openapi.json`).

The backend reads `DB_DSN` directly. Local Compose receives the Docker-local DSN from the root `.env`; production receives the Supabase DSN from GitHub Actions secrets.
