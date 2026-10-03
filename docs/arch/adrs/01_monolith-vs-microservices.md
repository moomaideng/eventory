# ADR-01: Monolith vs Microservices

## Context

Eventory supports three business modes on one account: Organizer (host and run tournaments, optional crowdfunding), Sponsor (pledge and pay), and Competitor (register solo or as a team). UC-02 and UC-03 depend on payment outcomes. UC-04 updates match results that other parties need to see. Notification after payment or match updates is valuable but must not block the main request path.

Architecturally significant requirements (ASRs) and quality drivers (SNDIC) that matter here:

- **Independent change rate**: Payment rules, gateway adapters, and ledger handling change on a different cadence than tournament/bracket logic.
- **Failure isolation**: A slow or failing payment provider must not take down browse/create tournament or record-result paths.
- **Deployability**: The team of five needs to ship payment and notification increments without redeploying the whole product every time.
- **Data ownership**: Tournament state, payment ledger, account profiles, and notification delivery logs have different consistency and retention needs.

Options considered:

- **Monolith**: one deployable app. Simpler local setup and transactions across modules, but payment and tournament logic share fate on deploy and failure, and team ownership boundaries blur as features grow.
- **Microservices (coarse-grained)**: split by business capability (Account, Tournament, Payment, Notification), each with its own data store. Higher operational cost, clearer ownership, independent deploy and failure domains for payment vs tournament.

## Decision

Adopt a **coarse-grained microservices** architecture aligned to business capabilities:

- Account Service
- Tournament Service
- Payment Service
- Notification Service

Do not split further into tiny technical services. Keep collaboration explicit via APIs and events (see ADR-05, ADR-06, ADR-07).

## Status

Accepted

## Consequences

**Positive**

- Payment and tournament can deploy and fail more independently.
- Each service owns its data and business rules, which matches the Service–Operations–Collaborators design.
- Async notification work can sit off the critical path of register/sponsor/record-result.

**Negative**

- Cross-service flows need IPC, contracts, and careful consistency (payment then registration/sponsorship).
- Local development and ops cost more than a single process.
- The team must keep service count small so the split stays maintainable for five people.
