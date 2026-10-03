# ADR-08: Object Storage

## Context

Sponsors upload brand logos shown on tournament pages after a successful sponsorship. Logos are binary objects (images), not relational rows. Storing large blobs in PostgreSQL bloats backups and couples file IO to the database. Tournament Service owns sponsorship display data and needs a durable place to store and fetch logo bytes or URLs.

ASRs: durable object storage for sponsor logos; Tournament (or its adapter) can store/fetch without putting file bytes in Postgres.

Options considered:

- **PostgreSQL bytea / similar**: transactional with row data, but poor fit for large binaries and CDN-style serving.
- **S3-compatible object storage**: standard object API, separates blob storage from service DBs; needs credentials and bucket lifecycle.
- **Only external URLs typed by sponsors**: simplest, but no first-party upload/control and broken links are common.

## Decision

Use **S3-compatible object storage** for sponsor logos. Tournament Service stores/fetches objects (or persists object keys/URLs next to sponsorship records in its own database). Account profile avatars may use the same approach later; v1 focus is sponsor logos on tournaments.

## Status

Accepted

## Consequences

**Positive**

- Binary assets stay out of PostgreSQL.
- Fits the architecture diagram collaborator (Object Storage) for Tournament.
- Can swap providers (AWS S3, MinIO, Supabase Storage) behind an S3 API.

**Negative**

- Extra infra and credentials to manage in each environment.
- Need rules for public vs private objects and cleanup when sponsorships change.
- Latency and failure modes differ from DB writes; upload UX must handle errors.
