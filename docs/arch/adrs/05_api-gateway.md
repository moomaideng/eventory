# ADR-05: API Gateway

## Context

The web client calls several backend capabilities (Account, Tournament, and related APIs). Exposing every service port to the browser couples the client to internal topology and multiplies CORS, TLS, and entry-point setup on each service. The browser needs one public entry so internal hostnames and ports stay hidden.

Options considered:

- **No gateway (client calls each service directly)**: simple early on; leaks service ports and multiplies auth/CORS setup.
- **Traefik as API Gateway**: reverse proxy with dynamic routing (e.g. labels/config); fits container deploys; team must learn Traefik config.
- **nginx as API Gateway**: very common and stable; routing is mostly static config, less dynamic than Traefik for many small services.
- **Kong (or similar) as API Gateway**: rich plugin ecosystem (auth, rate limit); heavier to run than Traefik for a small team.

## Decision

Use an **API Gateway**. Choose **Traefik** as that gateway: one public entry URL, path-based routing to Account, Tournament, and other public service APIs.

## Status

Accepted

## Consequences

**Positive**

- One public entry point for the web client.
- Internal service hostnames and ports stay behind the gateway.
- Traefik works well with container label-based routing as services are added.

**Negative**

- Gateway config and routing become critical path for all UI calls.
- Misrouted paths fail at the edge; team must keep Traefik routes in sync with services.
