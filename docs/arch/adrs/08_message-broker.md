# ADR-08: Message Broker (Asynchronous Collaboration)

## Context

After payment succeeds, registration is submitted, sponsorship is confirmed, or a match result is recorded, Eventory should notify interested parties (email and/or in-app). That work is not required to complete the writer's database transaction in the same synchronous call stack. Coupling Tournament or Payment to Notification over sync RPC would make user actions wait on email providers and would amplify Notification outages.

Domain events in scope:

- `payment.succeeded`
- `registration.submitted`
- `match_result.recorded`
- `sponsorship.confirmed`

Notification delivery can be eventually consistent. Retries must not block UC-02 / UC-03 / UC-04 success paths. Publishers and Notification should stay decoupled.

Options considered:

- **Sync call Tournament/Payment → Notification**: simple mentally, but user latency and failure coupling are poor.
- **Message broker**: publishers emit events; Notification consumes and delivers (email / in-app) with retry.
- **No notifications in architecture**: smaller system, misses sponsor/competitor communication needs in the product story.

## Decision

Use a **message broker** for asynchronous collaboration. Payment and Tournament **publish** the domain events above. Notification Service **consumes** them and sends email / in-app notifications. Prefer at-least-once delivery with idempotent consumers so delivery is eventually consistent.

## Status

Accepted

## Consequences

**Positive**

- Registration, sponsorship, and match-result paths stay responsive while notification runs separately.

**Negative**

- Eventual consistency: UI may confirm success before the email arrives.
- Duplicate deliveries are possible; Notification must handle the same event safely more than once.
