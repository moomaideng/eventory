# Architecture Decision Records

This directory contains the Architecture Decision Records (ADRs). Each ADR documents one significant architectural decision using the Nygard template: **Context**, **Decision**, **Status**, **Consequences**.

ADRs are numbered in the order they are decided. They are never renumbered. They are not rewritten after acceptance. A reversed decision gets a new ADR that supersedes the old one. Mark the old one's Status as `Superseded by ADR-NN`.

## Index

| # | Title | Status |
|---|-------|--------|
| [01](./01_monolith-vs-microservices.md) | Monolith vs Microservices | Accepted; partially superseded by ADR-12, ADR-13, ADR-14 |
| [02](./02_monorepo-vs-polyrepo.md) | Monorepo vs Polyrepo | Accepted |
| [03](./03_auth-pattern.md) | Authentication Pattern | Accepted |
| [04](./04_database-choice.md) | Database Technology | Superseded by ADR-12 |
| [05](./05_client-delivery.md) | Client Delivery | Accepted |
| [06](./06_frontend-language-framework.md) | Frontend Language & Framework | Accepted |
| [07](./07_object-storage.md) | Object Storage | Accepted |
| [08](./08_backend-language.md) | Backend Language | Accepted |
| [09](./09_backend-http-api-stack.md) | Backend HTTP & API Stack | Accepted |
| [10](./10_backend-coding-style.md) | Backend Coding Style | Superseded by ADR-11 |
| [11](./11_backend-coding-style-v2.md) | Backend Coding Style v2 | Accepted |
| [12](./12_database-technology-v2.md) | Database Technology v2 | Accepted |
| [13](./13_internal-ipc-style.md) | Internal IPC Style | Accepted |
| [14](./14_message-broker.md) | Message Broker | Accepted |
| [15](./15_api-gateway.md) | API Gateway and Reverse Proxy | Accepted |
| [16](./16_authentication-service-boundary.md) | Authentication Service Boundary and Session Management | Accepted |
| [17](./17_runtime-environments-service-discovery.md) | Runtime Environments and Service Discovery | Accepted |
| [18](./18_observability-stack.md) | Observability and Telemetry Stack | Accepted |
| [19](./19_dashboard-read-model-cqrs.md) | Dashboard Read Model and CQRS | Accepted |
