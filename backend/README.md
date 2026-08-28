# Eventory backend

This directory contains the Go backend service. It is built utilizing a structured architecture to separate core business logic from frameworks, external APIs, and database implementations.

## Technology Stack

*   **Language:** Go 1.26.5
*   **API Framework:** Huma v2 (Automated OpenAPI 3.1 documentation)
*   **Router:** go-chi/chi
*   **Database & ORM:** PostgreSQL with GORM
*   **Configuration:** Viper
*   **Live Reload:** Air

## Repository Architecture

```text
.
├── backend/
│   ├── cmd/
│   │   └── api/
│   │       └── main.go         # Application entry point (starts server & auto-migrates DB)
│   ├── internal/               # Private application code
│   │   ├── handlers/           # Huma HTTP handlers and DTOs
│   │   ├── middlewares/        # HTTP middlewares (JWT authentication, logging)
│   │   ├── models/             # Domain and Database models combined (GORM tags)
│   │   ├── repositories/       # Data access layer (GORM queries)
│   │   ├── server/             # Server wiring, Chi router, and Huma initialization
│   │   └── usecases/           # Core business logic
│   ├── pkg/                    # Public/Shared utilities
│   │   ├── baserepo/           # Generic repository helpers
│   │   ├── config/             # Viper configuration loading
│   │   └── database/           # PostgreSQL connection initialization
│   ├── .air.toml               # Air configuration for live reloading
│   ├── .env.example            # Environment variables template
│   ├── docker-compose.yaml     # Local database provisioning
│   ├── Dockerfile              # Container build for production
│   ├── Makefile                # Development automation commands
│   └── README.md               # Backend-specific documentation
```

## Prerequisites

Before running the server, ensure the following dependencies are installed on your local machine:

* **Go:** Version 1.22 or higher.
* **Docker & Docker Compose:** For running the local PostgreSQL instance.
* **Make:** For executing automation commands.
* **Air:** For automatic server reloading during development. Install via `go install github.com/cosmtrek/air@latest`.

1. **Configure Environment Variables:**
Duplicate the example environment file.

```bash
cp .env.example .env
```

2. **Initialize Infrastructure:**
Start the PostgreSQL database in the background.

```bash
docker compose up -d
```

3. **Run Database Migrations:**
Execute GORM auto-migrations to build the database schema based on your current models.

```bash
make migrate
# or make reset (if db schema need reset)
```

4. **Seed the Database:**
Populate the newly migrated tables with mock data (Accounts, Profiles, Roles) for local development.

```bash
make seed
```

5. **Start the Application:**
Run the server using Air to enable live reloading upon file saves.

```bash
make dev
# or make run (if air is not installed)
```

The API will start at `http://localhost:8080`. Interactive documentation (OpenAPI 3.1) is automatically generated and accessible at `http://localhost:8080/docs`.
