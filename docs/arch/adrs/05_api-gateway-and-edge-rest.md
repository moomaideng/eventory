# ADR-05: API Gateway and Edge REST

## Context

The web client calls several backend capabilities (Account, Tournament, and later Payment status as needed). Exposing every service port to the browser couples the client to internal topology and complicates CORS, TLS, and auth at each service. The proposal selects Traefik with dynamic routing and Docker Swarm DNS for service discovery in production-like deploys.

ASRs: single public entry for the browser; REST/JSON for the client; identity checked at the edge where practical; internal topology hidden from the UI.

Options considered:

- **Client calls each service directly**: simple early on, leaks service ports and multiplies auth/CORS setup.
- **API Gateway (Traefik) + REST to services**: one edge URL, path-based routing, central place for TLS and request entry.
- **GraphQL gateway**: flexible client queries, extra gateway and schema work the team does not need yet.

## Decision

Put **Traefik** in front as the **API Gateway**. The web client uses **REST** over the gateway to reach public service APIs. Service discovery uses platform DNS (e.g. Docker Swarm). Browser traffic does not use gRPC.

## Status

Accepted

## Consequences

**Positive**

- One public entry point for the web client.
- Internal service hostnames and ports stay behind the gateway.
- REST stays easy to demo and debug from the browser and OpenAPI tools.

**Negative**

- Gateway config and routing become critical path for all UI calls.
- Misrouted paths fail at the edge; team must keep Traefik labels/routes in sync with services.
- Coarse auth at the gateway does not replace business authorization inside each service.
