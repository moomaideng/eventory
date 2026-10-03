# ADR-01: Monolith vs Microservices

## Context

This is a Software Architecture course term project. The syllabus and graded deliverables require the team to design and build a microservices system. Those deliverables include the Service–Operations–Collaborators table, per-service design and APIs, and DDD-based decomposition.

The product is a personal expense tracker for college students. v1 is small: auth, bank-hook receipt ingestion, income/expense/category CRUD, and a dashboard. If we only looked at the product, a monolith would be simpler to build and run for a 5-person team in one semester.

The team has 5 people, which is small for many independently deployable services. Build, deploy, and observability work grow with each extra service.

Options considered:

- **Monolith**: one app to build, run, and deploy that is simpler for a 5-person semester project, but it does not meet the course's service-decomposition deliverable.
- **Microservices**: one service per DDD subdomain, each with its own data, which meets the course requirements but adds extra deploy and communication work.

## Decision

Adopt a microservices architecture. Split the system by DDD subdomain. The main reason is the course's architecture practice requirements.

Keep the split coarse-grained. Do not create more services than needed. Pair this with a monorepo (ADR-02) and a simple v1 stack: one REST style, one database technology, and no message broker. The course allows this for a first architecture iteration.

## Status

Accepted. The v1 constraints of one REST style, one database technology, and no message broker are partially superseded by ADR-12, ADR-13, and ADR-14. The decision to use microservices remains.

## Consequences

**Positive**

- Meets the graded deliverable for service decomposition, service APIs, and collaboration diagrams.
- Makes each service own its business capability and its data. That is good practice at any product size.

**Negative**

- Extra operational work: inter-service calls, per-service deploys, and data owned by different services. This is more than v1 of the product needs.
- Local development is slower than a single monolith app.
- Some of this complexity waits until service boundaries are set in the Microservice Design step. That includes reverse proxy, dev/prod environment, internal IPC style, message broker, and observability stack.
