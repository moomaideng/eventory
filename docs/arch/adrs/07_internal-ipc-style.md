# ADR-07: Internal Synchronous IPC Style

## Context

Some flows need a synchronous answer from another service in the same user action (for example Tournament confirming payment before it finalizes sponsorship or registration). The public API already uses REST (ADR-06).

Options considered:

- **REST/JSON internally**: one protocol shared with the public API; easy to debug; contracts are looser unless the team is strict with OpenAPI.
- **gRPC (Protocol Buffers)**: typed RPC contracts, generated clients/servers, good fit for sync service-to-service request/response; extra protobuf tooling beside OpenAPI.
- **Apache Avro**: strong schema evolution; common for data pipelines and events; weaker everyday request/response microservice tooling than gRPC.

## Decision

Use **gRPC** for **synchronous service-to-service** calls. Protobuf contracts live in the monorepo. Callers set deadlines and avoid calling another service's database directly.

Asynchronous collaboration is covered by ADR-08.

## Status

Accepted

## Consequences

**Positive**

- Clear, typed contracts for synchronous calls between services.

**Negative**

- Team must maintain protobuf tooling in addition to OpenAPI for REST.
- Timeouts and retries must be set carefully or failures can cascade between services.
