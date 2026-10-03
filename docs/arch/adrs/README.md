# Architecture Decision Records

Each ADR documents one significant architectural decision using the Nygard template: **Context**, **Decision**, **Status**, **Consequences**.

ADRs are numbered in the order they are decided. They are never renumbered. They are not rewritten after acceptance. A reversed decision gets a new ADR that supersedes the old one. Mark the old one's Status as `Superseded by ADR-NN`.

Justify decisions with architecturally significant requirements (ASRs) and quality drivers (SNDIC), not with "the assignment requires it." Put technology-fit arguments in **Context**; keep **Consequences** for what follows after the decision.

## Index

| ADR | Status |
|---|--------|
| [01 Monolith vs Microservices](./01_monolith-vs-microservices.md) | Accepted |
| [02 Monorepo vs Polyrepo](./02_monorepo-vs-polyrepo.md) | Accepted |
| [03 Authentication Pattern](./03_authentication-pattern.md) | Accepted |
| [04 Database Technology](./04_database-technology.md) | Accepted |
| [05 API Gateway and Edge REST](./05_api-gateway-and-edge-rest.md) | Accepted |
| [06 Internal Synchronous IPC (gRPC)](./06_internal-ipc-grpc.md) | Accepted |
| [07 Message Broker (Asynchronous Collaboration)](./07_message-broker.md) | Accepted |
| [08 Object Storage](./08_object-storage.md) | Accepted |
