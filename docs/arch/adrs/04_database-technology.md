# ADR-04: Database Technology

## Context

Most domain data is relational and needs strong consistency: accounts and role profiles, tournaments and lifecycle, team lobbies and membership, payments and ledger rows, notification delivery logs. These records reference each other (for example a lobby belongs to a tournament; a payment belongs to a registration or sponsorship). The system needs foreign keys, joins, and ACID transactions.

Custom registration forms are different. Each tournament defines its own fields (text, dropdown, file metadata). The structure changes per tournament, so a fixed table-per-field design is awkward. That part needs a flexible document-style store. Form data still belongs to Tournament Service (database-per-service: not a shared database other services read directly).

Options considered:

- **SQL only (e.g. PostgreSQL)**: strong for relational data, constraints, and transactions; awkward for per-tournament form documents that change shape often.
- **NoSQL document store only (e.g. MongoDB)**: flexible for form schemas and answers; weak fit for relational tournament and payment rules that need transactions and clear relationships.
- **SQL per service + document store for Tournament forms (PostgreSQL + MongoDB)**: use a relational database where ACID and relations matter; use a document store only for flexible registration forms.

## Decision

- Use **PostgreSQL** as the default datastore, **one database per service**: Account, Tournament, Payment, Notification.
- Use **MongoDB** only in **Tournament Service** for custom registration form schemas and submitted answers.

No shared database across services. Other services access Tournament or Payment data through APIs/events, not by reading foreign tables.

## Status

Accepted

## Consequences

**Positive**

- Each service can backup, migrate, and scale its store without coordinating a single shared schema.
- New registration form fields can ship without relational migrations across the whole platform.

**Negative**

- Two datastore technologies to operate (PostgreSQL + MongoDB) for Tournament.
- Cross-service consistency (e.g. payment then registration) needs application-level flows, not cross-database transactions.
