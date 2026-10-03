# ADR-06: Internal Synchronous IPC (gRPC)

## Context

Tournament Service must start and check payment as part of sponsor and registration flows (`CreateCheckout`, `GetPaymentStatus`) before it can confirm sponsorship or registration. That collaboration is synchronous: Tournament needs a result in the same user action. The browser already uses REST at the edge (ADR-05). Using REST again internally is possible but gives looser contracts for service-to-service calls the team owns.

ASRs: typed internal contracts for payment collaboration; deadlines on sync calls; no browser dependency on internal RPC.

Options considered:

- **REST internally everywhere**: one protocol to learn; weaker typed contracts and more ad hoc HTTP between services.
- **gRPC internally for sync calls**: protobuf contracts, generated clients/servers, explicit RPC APIs for Tournament → Payment.
- **gRPC to the browser**: not chosen; web edge stays REST (ADR-05).

## Decision

Use **gRPC** for **synchronous service-to-service** calls, starting with **Tournament Service → Payment Service** (`CreateCheckout`, `GetPaymentStatus`, and related payment CRUD used internally). Protobuf contracts live in the monorepo. Callers set deadlines and avoid calling another service's database directly.

Asynchronous collaboration is covered by ADR-07.

## Status

Accepted

## Consequences

**Positive**

- Clear, typed internal API between Tournament and Payment.
- Matches the microservice design: payment orchestration is not done by the browser talking to Payment for core flows.
- Easier to evolve Payment internals behind a stable RPC contract.

**Negative**

- Team maintains protobuf tooling beside OpenAPI for REST.
- Local debugging of gRPC is heavier than plain HTTP JSON.
- Timeouts and retries must be configured carefully to avoid cascading load.
