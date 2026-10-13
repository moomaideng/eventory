# ADR-11: SMTP for Notification Email Delivery

## Context

ADR-01 gives Notification its own service, and ADR-08 and ADR-10 decouple it
from the registration request through RabbitMQ. Notification still needs a way
to submit email to an external delivery service. The team needs to inspect
messages locally and use a real email provider in production without changing
the event contract or the Tournament Service. Provider outages must be handled
by Notification rather than delaying registration.

Options considered:

- **SMTP relay**: use a standard mail-submission protocol with a local test
  server and a configured production provider. It requires MIME generation,
  credentials, and TLS configuration; relay acceptance does not prove delivery
  to a recipient's inbox. See [RFC 6409](https://www.rfc-editor.org/rfc/rfc6409.html)
  and [RFC 5321](https://www.rfc-editor.org/rfc/rfc5321.html).
- **Provider-specific HTTPS email API**: may expose provider features such as
  delivery status or templates, but Notification would depend on that
  provider's request format and credentials. Changing provider would require
  a new adapter and corresponding tests.
- **Operate Eventory's own mail server**: avoids an external submission API,
  but adds mail-server security, reputation, and delivery operations beyond
  this project's scope.

## Decision

Use an **SMTP relay** for outbound Notification email. During local development,
send to Mailpit so the team can inspect messages without external delivery.
In production, configure an external SMTP provider and sender address through
environment variables; require STARTTLS for the current production adapter.
Notification renders both plain-text and HTML parts and submits them as one
message. The SMTP provider is not fixed by this ADR.

The first template acknowledges a `registration.submitted` event and says the
registration is awaiting organizer review. Choosing SMTP does not change the
asynchronous event decision in ADR-08 or the broker choice in ADR-10.

## Status

Accepted

## Consequences

**Positive**

- The same Notification sender can target Mailpit locally and a configured
  SMTP relay in production.
- Tournament does not need SMTP credentials or a synchronous email call.

**Negative**

- The team must configure and protect SMTP credentials, a verified sender
  address, and TLS settings for its chosen provider.
- An SMTP success response means the relay accepted the message; it does not
  guarantee delivery to the recipient's inbox.
- SMTP submission has no application idempotency key. If the relay accepts a
  message but Notification stops before recording `SENT`, a retry can send a
  duplicate email.
