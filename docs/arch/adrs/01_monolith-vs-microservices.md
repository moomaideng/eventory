# ADR-01: Monolith vs Microservices

## Context

Eventory supports three business modes on one account: Organizer (host and run tournaments, optional crowdfunding), Sponsor (pledge and pay), and Competitor (register solo or as a team). UC-02 and UC-03 depend on payment outcomes. UC-04 updates match results that other parties need to see. Notification after payment or match updates is valuable but must not block the main request path.

ASR / SNDIC drivers that matter here:

- **Failure isolation**: If payment is slow or down, users should still browse tournaments, create events, and record match results.
- **Deployability**: The team of five should be able to release payment or notification updates without redeploying the whole product every time.
- **Data ownership**: Tournament data, payment records, account profiles, and notification logs need different rules for how long they are kept and how strictly they stay consistent.

Options considered:

- **Monolith**: one deployable app. Simpler local setup and transactions across modules, but payment and tournament logic share fate on deploy and failure, and team ownership boundaries blur as features grow.
- **Microservices**: split by business capability (Account, Tournament, Payment, Notification), each with its own data store. Higher operational cost, clearer ownership, independent deploy and failure domains for payment vs tournament.

## Decision

Adopt a **microservices** architecture aligned to business capabilities:

- Account Service
- Tournament Service
- Payment Service
- Notification Service

Keep collaboration explicit via APIs and events (see ADR-05 through ADR-08).

## Status

Accepted

## Consequences

**Positive**

- Payment and tournament can deploy and fail more independently.
- Each service owns its data and business rules.
- Notifications can run asynchronously so registration, sponsorship, and match result updates stay responsive.

**Negative**

- Cross-service flows need IPC, contracts, and careful consistency.
- Local development and day-to-day operations cost more than a single process.
