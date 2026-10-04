# ADR-06: Public API Style

## Context

The web client needs a stable way to talk to Eventory through the API Gateway (ADR-05). Public client APIs should be easy to call from a browser and easy to document.

Options considered:

- **REST with JSON for the public API**: familiar for web clients and tools (OpenAPI, browser, curl); resource-oriented APIs fit Account and Tournament CRUD.
- **GraphQL for the public API**: flexible client queries; extra schema and gateway work the team does not need yet.
- **gRPC to the browser**: typed and efficient; poor fit for a normal web app.

## Decision

Use **REST with JSON** for all **browser-facing** (public) APIs through the API Gateway. Document those APIs with OpenAPI where the stack supports it.

## Status

Accepted

## Consequences

**Positive**

- Easy to call and debug from the browser and standard HTTP tools.

**Negative**

- One screen may need several REST calls.
- Public REST contracts must be versioned carefully once the UI depends on them.
