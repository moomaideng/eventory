# Backend Architecture & Agent Guidelines

This repository follows Clean Architecture principles adapted for Go with **Huma v2**, **GORM** (PostgreSQL), and **Go Mongo Driver** (MongoDB).

All backend services (`services/tournament`, `services/account`, `services/payment`) follow this structure and strict separation of concerns.

---

## 1. Canonical Layer Definitions

```
       ┌────────────────────────────────────────────────────────┐
       │             Transport: handlers/rest/                  │
       │   • Huma v2 Operation registration                     │
       │   • Request / Response DTOs with validation tags       │
       │   • Context auth extraction & HTTP error mapping       │
       └───────────────────────────┬────────────────────────────┘
                                   │ Calls
                                   ▼
       ┌────────────────────────────────────────────────────────┐
       │             Domain: internal/usecases/                 │
       │   • Core business logic, workflows, state machines     │
       │   • Invariant enforcement & authorization rules        │
       │   • Depends strictly on interfaces (Repositories/Ports)│
       │   • Returns domain models/views & domain errors        │
       └──────────────┬──────────────────────────┬──────────────┘
                      │                          │
        Calls interface                          │ Calls interface
                      ▼                          ▼
┌──────────────────────────────────────┐  ┌─────────────────────────────────────┐
│   Persistence: internal/repositories/│  │     External: internal/ports/       │
│ • Pure database queries & GORM/Mongo │  │ • Cross-service RPC definitions     │
│ • Transactions, row locking, offsets │  │ • Implemented in internal/adapters/ │
│ • Maps DB errors to sentinel errors  │  │   (gRPC client, Payment HTTP client)│
└──────────────────────────────────────┘  └─────────────────────────────────────┘
                      │
                      ▼
       ┌────────────────────────────────────────────────────────┐
       │                    models/                             │
       │   • Database entity schemas with GORM/BSON tags        │
       │   • Enums, constants, status lifecycle definitions     │
       │   • Pure domain helper methods (no external calls)     │
       └────────────────────────────────────────────────────────┘
```

### Layer Responsibilities & Boundaries

| Layer | Path | Allowed Dependencies | Prohibited Dependencies |
|---|---|---|---|
| **Transport** | `handlers/rest/` | Huma v2, DTOs, Use Cases, Middlewares | GORM, SQL, Repositories, DB transactions |
| **Domain** | `internal/usecases/` | Domain Models, Repository Interfaces, Port Interfaces | `net/http`, Huma, GORM, MongoDB, SQL |
| **Persistence**| `internal/repositories/`| GORM, Mongo, DB Drivers, Models | HTTP, Huma, Use Cases, Ports |
| **Models** | `models/` | Standard library (time, uuid) | Any other internal package |

---

## 2. Architecture Rules & Anti-Patterns

- Do not split a single domain entity into multiple repositories or use cases merely because different frontend pages exist.
- Never put business logic, transition validations, or mutating closures into repositories.
- Repositories only perform queries, record updates, lock rows (`SELECT ... FOR UPDATE`), and manage transactions.
- Transition rules and permissions must always be evaluated in the use case before or during transactional execution.
- Keep files focused and readable.
- Pass all required dependencies (Repositories, Ports, Services) through constructor functions (`New<Component>UseCase(...)`).
- Never use `.With<Dependency>(...)` chaining to resolve circular dependencies between use cases. Merge the overlapping concepts into a single domain usecase instead.
- Whenever routes, operations, path parameters, query parameters, request bodies, or response DTOs are added or modified:
  ```bash
  npm --prefix frontend run openapi:generate
  ```
  This regenerates `frontend/lib/api/schema.d.ts` and guarantees type safety between frontend and backend.

## 3. Testing & Verification Checklist

Before opening a PR or finishing a task:
1. **Unit Tests**: Run `go test -v ./services/<service>/...`.
2. **Acceptance Tests**: Ensure test suites compile: `go test -c -o /dev/null ./tests/...`.
3. **Static Analysis**: Run `go vet ./...` from the `backend/` directory. Zero errors or linter warnings are permitted.
4. **OpenAPI Sync**: If any endpoint schemas changed, run `npm --prefix frontend run openapi:generate` and ensure `npm --prefix frontend run lint && npm --prefix frontend run typecheck` passes.

---

## 4. Canonical CRUD Reference Implementation

For a clean, end-to-end CRUD reference adhering strictly to these layer boundaries and conventions, follow the **Tournament Staff** endpoint:

- **Entity Model:** [`services/tournament/models/tournament_staff.go`](services/tournament/models/tournament_staff.go)
  - Pure GORM entity definition, role enums, table name, and domain helper methods.
- **Persistence Repository:** [`services/tournament/internal/repositories/tournament_staff_repository.go`](services/tournament/internal/repositories/tournament_staff_repository.go)
  - Interface definition and GORM implementation for database CRUD operations without business logic.
- **Domain Use Case:** [`services/tournament/internal/usecases/tournament_staff_usecase.go`](services/tournament/internal/usecases/tournament_staff_usecase.go)
  - Business rules, caller authorization checks, view projection, and dependency injection via constructor.
- **Transport Handler:** [`services/tournament/handlers/rest/tournament_staff_handler.go`](services/tournament/handlers/rest/tournament_staff_handler.go)
  - Huma v2 route registrations, request/response DTO schemas with doc/validation tags, auth context extraction, and error translation.
- **Unit Tests:** [`services/tournament/internal/usecases/tournament_staff_usecase_test.go`](services/tournament/internal/usecases/tournament_staff_usecase_test.go)
  - Table-driven unit tests using repository stubs and standard Go testing assertions.

