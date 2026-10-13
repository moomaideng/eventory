# Backend

The hidden server behind Eventory. It checks the rules and owns the database. Written in Go.

Setup and run instructions live in **[Getting Started](../docs/getting-started.md)**. This page is about the code.

---

## Run it

From the repo root, pick one:

```bash
# Postgres, MongoDB, Traefik, Account, Tournament, Payment, frontend on the host
task dev

# or the same stack without the frontend
task dev:backend
```

Both run `task migrate` first: Account, Tournament, and Payment schemas. `task migrate` alone starts Postgres and applies those schemas. It does not start the APIs.

```bash
# sample rows, after the apps are healthy
task seed
```

To run a service on the host instead of in Compose, use `task migrate`, then:

```bash
go run ./services/account/cmd/api
go run ./services/tournament/cmd/api
go run ./services/payment/cmd/api
```

`8081`, `9091`, `8082`, and `8083` must be free, so do not leave `task dev:backend` running. Traefik on `:8080` forwards to the containers, not to these host processes. Call `:8081`, `:8082`, and `:8083` directly. MongoDB runs on port `27017` as a secondary database for the Tournament Service.

| URL | What |
| --- | --- |
| http://localhost:8081/docs | Account endpoints |
| http://localhost:8082/docs | Tournament endpoints |
| http://localhost:8083/docs | Payment endpoints |
| http://localhost:8080/health | Gateway health check |
| localhost:27017 | MongoDB (database: `tournament_db`) |

---

## Where things go

A request travels through four layers within each microservice (`account`, `tournament`, `payment`). Each layer has a single job.

```mermaid
flowchart LR
    R["Request"] --> H["handlers/<br/>read & reply"]
    H --> U["usecases/<br/>the rules"]
    U --> Rep["repositories/<br/>database queries"]
    Rep --> DB[("PostgreSQL / MongoDB")]
```

| Folder | Job | Example question it answers |
| --- | --- | --- |
| `services/<name>/handlers/` | Read the request, send the reply | "What did the browser send?" |
| `services/<name>/internal/usecases/` | The actual rules | "Is this user allowed to close registration?" |
| `services/<name>/internal/repositories/` | Database reads and writes | "Fetch tournament 123" |
| `services/<name>/models/` | Shape of each table or collection document | "What columns does a tournament have?" |
| `services/<name>/cmd/` | One service each: `cmd/api`, `cmd/migrate`, and `cmd/seed` | — |
| `internal/` | Shared across services: `config`, `database`, `middlewares` | "Is this person logged in?" |

**Why split it up:** each microservice under `services/<name>` owns its domain logic and databases. Inside each service, rules in `usecases/` stay readable and testable because they contain no web code and no database queries.

---

## Adding an endpoint

1. Add or update a table shape in `services/<name>/models/`.
2. Add the query in `services/<name>/internal/repositories/`.
3. Put the rules in `services/<name>/internal/usecases/`, and write a test next to it.
4. Register the endpoint in `services/<name>/handlers/` with `huma.Register(...)`.
5. Check it appears at http://localhost:8081/docs, http://localhost:8082/docs, or http://localhost:8083/docs.
6. Tell the frontend about it. The apps must be up (`task dev:backend`):

```bash
# from backend/
npm --prefix ../frontend run openapi:generate
```

Step 6 is required. The frontend gets its types from the running backend, so skipping it leaves the frontend blind to your new endpoint.

---

## Commands

| Command | Does |
| --- | --- |
| `go run ./services/account/cmd/migrate` | Create or update Account tables. `-reset` drops `public` first |
| `go run ./services/tournament/cmd/migrate` | Create or update Tournament tables, including search indexes. `-reset` drops `public` first |
| `go run ./services/payment/cmd/migrate` | Create or update Payment tables. `-reset` drops `public` first |
| `go run ./services/account/cmd/api` | Account on the host. Reads `services/account/.env.default` |
| `go run ./services/tournament/cmd/api` | Tournament on the host. Reads `services/tournament/.env.default` |
| `go run ./services/payment/cmd/api` | Payment on the host. Reads `services/payment/.env.default` |
| `go run ./services/account/cmd/seed` | Sample accounts |
| `go run ./services/tournament/cmd/seed` | Sample tournaments |
| `go run ./services/payment/cmd/seed` | Sample mock payments |
| `go test ./...` | Run the tests |
| `go build ./services/account/cmd/api` | Check Account still compiles |
| `go build ./services/tournament/cmd/api` | Check Tournament still compiles |
| `go build ./services/payment/cmd/api` | Check Payment still compiles |
| `task gen-proto` | From the repo root. Lint and regenerate gRPC code. Needs [Buf](https://buf.build/docs/installation/) |

Run `go test ./...` before opening a pull request.

---

## Settings

Each service loads `services/<name>/.env.default`, then an optional `.env` in that same directory, then the process environment.

| Setting | Meaning |
| --- | --- |
| `HTTP_PORT` | Account listens on 8081. Tournament listens on 8082. Payment listens on 8083 |
| `GRPC_PORT` | Account gRPC port. Default 9091 |
| `DB_DSN` | Database address. Account uses `account_db`. Tournament uses `tournament_db`. Payment uses `payment_db` |
| `ACCOUNT_GRPC_ADDR` | Where Tournament calls Account. `127.0.0.1:9091` on the host, `account:9091` in Compose |
| `MONGO_URI` | MongoDB connection URI for Tournament Service dynamic forms |
| `MONGO_DB_NAME` | MongoDB database name for Tournament Service (`tournament_db`) |
| `PAYMENT_SERVICE_KEY` | Server-only secret for internal Payment endpoints (`X-Payment-Service-Key`) |
| `SUPABASE_URL` | Used to verify login tickets |
| `CORS_ALLOWED_ORIGINS` | Which websites may call this API |

Compose replaces `DB_DSN` so the containers talk to the `postgres` hostname. Host `go run` keeps the `localhost` value from `.env.default`.

---

## Login, briefly

```mermaid
flowchart LR
    U["User"] -->|"signs in with Google"| S["Supabase"]
    S -->|"signed ticket"| B["Browser"]
    B -->|"sends ticket with each request"| A["Backend"]
    A -->|"checks signature"| S
```

The backend never sees a password. It only checks that the ticket's signature is genuine.

---

Something broken? → **[Troubleshooting](../docs/troubleshooting.md)**
