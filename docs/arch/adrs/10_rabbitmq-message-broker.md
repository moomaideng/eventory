# ADR-10: RabbitMQ as Message Broker

## Context

ADR-08 requires asynchronous collaboration but does not select a broker. The
first use case is delivering `registration.submitted` to Notification without
delaying registration. Messages must survive restarts and failed work must be
retryable and inspectable.

Options considered:

- **RabbitMQ**: durable, service-owned queues with acknowledgements, confirms,
  TTL retries, and dead-lettering fit this worker. See the
  [RabbitMQ reliability guide](https://www.rabbitmq.com/docs/reliability).
- **Kafka**: useful for retained event history and high throughput, but adds
  offset, retention, and delayed-retry complexity not needed here.
- **Redis Streams**: supports consumer groups, but requires more application
  logic for scheduled retries and dead-lettering.

## Decision

Use **RabbitMQ** for the events in ADR-08. Publishers send versioned events to
an exchange; consumers own their queues. Notification consumes
`registration.submitted` with manual acknowledgements, persistent messages,
publisher confirms, retry queues, and a dead-letter queue. Consumers must be
idempotent. Retry intervals remain an implementation policy in the
[Notification Service guide](../../notification-service.md).

## Status

Accepted

## Consequences

**Positive**

- Producers and consumers remain decoupled; failed work can be retried or
  inspected without blocking registration.

**Negative**

- RabbitMQ requires deployment and monitoring; duplicate delivery remains
  possible.
- Tournament's future producer needs a transactional outbox to avoid losing
  events between its database commit and broker publish.
