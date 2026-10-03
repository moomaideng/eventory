# ADR-04: Database Technology

## Context

Most Eventory domain data is relational and needs ACID behavior: accounts and role profiles, tournaments and lifecycle, team lobbies and membership, payments and ledger rows, notification delivery logs. Foreign keys and transactions fit PostgreSQL well, and the team already runs Postgres in local Docker.

Custom registration forms are different: each tournament defines its own fields (text, dropdown, file metadata). That shape is document-like and varies per tournament, so a rigid relational schema per field is a poor fit. MongoDB fits flexible form schema and submitted answers while Tournament still owns that data (database-per-service: separate Mongo instance/database for Tournament forms, not a shared DB).

Options considered:

- **PostgreSQL only**: strong for relational/ACID data; awkward for per-tournament form documents.
- **MongoDB only**: flexible documents; weak fit for relational tournament/payment invariants the product relies on.
- **PostgreSQL per service + MongoDB for Tournament forms**: relational default where ACID and relations matter; document store only for the flexible form capability.

PostgreSQL gives mature transactions, constraints, and team familiarity. MongoDB gives schema flexibility exactly where registration forms need it. Together they match the data shapes without forcing one store to do both jobs poorly.

## Decision

- Use **PostgreSQL** as the default datastore, **one database (or schema-isolated instance) per service**: Account, Tournament, Payment, Notification.
- Use **MongoDB** only in **Tournament Service** for custom registration form schemas and submitted answers.

No shared database across services. Other services access Tournament or Payment data through APIs/events, not by reading foreign tables.

## Status

Accepted

## Consequences

**Positive**

- Each service can backup, migrate, and scale its store without coordinating a single shared schema.
- New registration form fields can ship without relational migrations across the whole platform.

**Negative**

- Two datastore technologies to operate (Postgres + Mongo) for Tournament.
- Cross-service consistency (e.g. payment then registration) needs application-level flows, not cross-DB transactions.
- Team must keep Mongo scoped to forms so it does not become a second general-purpose store.
