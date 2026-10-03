# ADR-02: Monorepo vs Polyrepo

## Context

Eventory has a web client, multiple backend services (ADR-01), infra config (API gateway, Docker Compose), and architecture docs (ADRs, use cases, microservice design). Changes often span layers: a registration field can touch Tournament API, Payment contract, frontend forms, and docs in one iteration.

Options considered:

- **Monorepo**: one repository for frontend, services, infra, and docs. One PR can land a vertical slice; CI may later need path filters.
- **Polyrepo**: one repo per service. Stronger isolation for large orgs, but versioning, cross-repo PRs, and release coordination cost too much for a five-person semester team.

## Decision

Use **one monorepo** for all services, the frontend, infra config, and architecture docs.

## Status

Accepted

## Consequences

**Positive**

- One PR can change API, client, and docs together.
- One place for CI and for reviewers to see the full system.
- Architecture docs stay versioned next to the code they describe.

**Negative**

- The repo grows with every service; CI should later use path-based filtering.
- Repo layout alone does not enforce service boundaries; review and CI must protect them.
