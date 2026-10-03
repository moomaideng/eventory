# ADR-07: Message Broker (Asynchronous Collaboration)

## Context

After payment succeeds, registration is submitted, sponsorship is confirmed, or a match result is recorded, Eventory should notify interested parties (email and/or in-app). That work is not required to complete the writer's database transaction in the same synchronous call stack. Coupling Tournament or Payment to Notification over sync RPC would make user actions wait on email providers and would amplify Notification outages.

Domain events already identified in the microservice design:

- `payment.succeeded`
- `registration.submitted`
- `match_result.recorded`
- `sponsorship.confirmed`

ASRs: eventually consistent notification delivery; retry without blocking UC-02/UC-03/UC-04 success paths; services stay decoupled.

Options considered:

- **Sync call Tournament/Payment → Notification**: simple mentally, but user latency and failure coupling are poor.
- **Message broker**: publishers emit events; Notification consumes and delivers (email / in-app) with retry.
- **No notifications in architecture**: smaller system, misses sponsor/competitor communication needs in the product story.

## Decision

Use a **message broker** for asynchronous collaboration. Payment and Tournament **publish** the domain events above. Notification Service **consumes** them and sends email / in-app notifications. Prefer at-least-once delivery with idempotent consumers (and outbox if needed later) so delivery is eventually consistent.

## Status

Accepted

## Consequences

**Positive**

- Register, sponsor, and record-result paths stay responsive.
- Notification can retry and scale separately from Tournament/Payment.
- New consumers can subscribe to the same events later without changing publishers' sync APIs.

**Negative**

- Eventual consistency: UI may confirm success before the email arrives.
- Team must operate broker infra and define event contracts.
- Duplicate deliveries require idempotent notification handling.
