# ADR-11: SMTP for Notification Email Delivery

## Context

After receiving `registration.submitted` through RabbitMQ (ADR-08 and ADR-10),
Notification must submit email without involving Tournament. Local messages
should be inspectable, while production can use an external provider.

Options considered:

- **SMTP relay**: works with local Mailpit and different production providers,
  but needs credentials and TLS. See [RFC 6409](https://www.rfc-editor.org/rfc/rfc6409.html).
- **Provider-specific HTTPS API**: can offer provider features, but couples the
  adapter to one provider.
- **Own mail server**: adds security and deliverability operations beyond this
  project's scope.

## Decision

Use an **SMTP relay**: Mailpit locally and a configured provider in production
with STARTTLS. Notification sends plain-text and HTML parts in one email. The
first template confirms receipt of registration, not acceptance. This ADR does
not select a production provider.

## Status

Accepted

## Consequences

**Positive**

- Notification can change SMTP providers without changing the event contract;
  Tournament needs no SMTP integration.

**Negative**

- Production needs protected credentials and a verified sender address.
- Relay acceptance does not guarantee inbox delivery. A crash after acceptance
  but before recording `SENT` can cause a duplicate email.
