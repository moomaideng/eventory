# ADR-09: Object Storage

## Context

Sponsors upload brand logos shown on tournament pages after a successful sponsorship. Logos are binary objects (images), not relational rows. Storing large blobs in PostgreSQL bloats backups and couples file IO to the database. Tournament Service owns sponsorship display data and needs durable object storage so it can store/fetch logos without putting file bytes in Postgres.

Options considered:

- **PostgreSQL bytea / similar**: transactional with row data, but poor fit for large binaries and CDN-style serving.
- **S3-compatible object storage**: standard object API, separates blob storage from service DBs; needs credentials and bucket lifecycle.
- **Only external URLs typed by sponsors**: simplest, but no first-party upload/control and broken links are common.

## Decision

Use **S3-compatible object storage** for sponsor logos. Tournament Service is responsible for store and fetch.

## Status

Accepted

## Consequences

**Positive**

- Binary assets stay out of PostgreSQL.
- Can swap providers (AWS S3, MinIO, Supabase Storage) behind an S3 API.

**Negative**

- Extra infra and credentials to manage in each environment.
- Object storage and the database are not one transaction, so the service must handle partial failure (for example cleanup or retry) when one write succeeds and the other fails.

