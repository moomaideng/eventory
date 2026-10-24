# ADR-12: Service Discovery and Container Orchestration

## Context

Services must find each other and the API Gateway (ADR-05) must find every public service. Container IPs change on every deploy, so addresses cannot be hardcoded. Production is a single VM deployed by CI over SSH, and a five-person team must be able to operate it, so the orchestrator should be cheap to run and to learn.

Options considered:

- **Docker Compose on one host**: Docker's embedded DNS resolves service names, and Traefik's Docker provider builds routes from container labels. No extra component, and the same files run locally and in production. No scheduling across hosts, rolling updates, or failover.
- **Docker Swarm**: a virtual IP per service with built-in balancing across replicas, plus rolling updates. Needs the Compose files converted to stack files and Traefik in Swarm mode. Its benefits need more than one node, which we do not have.
- **Kubernetes**: self-healing, rolling updates, autoscaling, but too heavy to run and learn for one VM and this team.

## Decision

**Docker Compose on the single production host**.

- **Internal discovery**: services call each other by Compose service name (for example `account:9091`), resolved by Docker's embedded DNS.
- **Public discovery**: Traefik's Docker provider reads router labels from running containers; a new service is exposed by adding labels (ADR-05).

Swarm and Kubernetes are deferred until we need more than one host or automatic failover.

## Status

Accepted

## Consequences

**Positive**

- No extra infrastructure; names stay stable across redeploys.
- A service can run several replicas (`docker compose up --scale`); Traefik adds each labeled container to the same service and balances HTTP across them.

**Negative**

- The host is a single point of failure; health checks and restart policies recover containers, not the machine.
- A redeploy restarts containers and causes a short gap; retries (NFR7) cover it for read calls only.
- Scaling a service balances HTTP traffic only: gRPC calls to Account stay on one replica, because a gRPC client keeps a single connection by default.
