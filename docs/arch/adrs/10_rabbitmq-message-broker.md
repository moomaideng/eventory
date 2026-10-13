# ADR-10: RabbitMQ as Message Broker

## Context

ADR-08 decides that Eventory will use a message broker for asynchronous
collaboration between Tournament, Payment, and Notification. It does not choose
a broker product. Registration email is the first implementation: Notification
needs to receive `registration.submitted` without delaying the registration
request, and a future consumer should be able to subscribe to the same domain
event independently. Messages must survive ordinary service restarts, and
failed email work must be inspectable and retried without blocking producers.

Options considered:

- **RabbitMQ**: exchanges route an event to service-owned queues; durable
  queues, consumer acknowledgements, publisher confirms, TTL, and dead-letter
  exchanges support this worker-and-retry flow. See the
  [RabbitMQ reliability guide](https://www.rabbitmq.com/docs/reliability) and
  [dead-letter exchange guide](https://www.rabbitmq.com/docs/dlx). It adds a
  broker to operate and does not remove the need for an idempotent consumer.
- **Apache Kafka**: retained, partitioned logs make replay and independent
  consumers useful when Eventory needs event history or substantially higher
  throughput. For this first email worker, managing topics, offsets, retention,
  and a delayed retry path adds work that the current use case does not need.
  See [Kafka documentation](https://kafka.apache.org/documentation/).
- **Redis Streams**: consumer groups can distribute work and track pending
  messages, but this use case would need additional application logic for
  scheduled retries and dead-letter handling. See the
  [Redis Streams documentation](https://redis.io/docs/latest/develop/data-types/streams/).

## Decision

Choose **RabbitMQ** as the message broker for the asynchronous events selected
in ADR-08. Publishers send versioned domain events to an exchange; each
consumer owns its queue and bindings. Notification uses manual acknowledgements
and publisher confirms, with durable queues and persistent messages. The first
consumer handles `registration.submitted`; retry queues and a dead-letter queue
hold failed email work. Notification must tolerate duplicate deliveries.

This selects the broker, not the exact retry intervals or an email provider.
Those implementation policies are recorded in the
[Notification Service guide](../../notification-service.md).

## Status

Accepted

## Consequences

**Positive**

- A publisher can emit one event without knowing which consumers subscribe.
- Queue routing, acknowledgements, and dead-lettering fit the first
  Notification worker without building a custom queue on PostgreSQL.

**Negative**

- The team must deploy, secure, monitor, and back up another service.
- Broker delivery can be repeated; consumers still need idempotency.
- A producer must also handle the gap between committing its own database and
  publishing an event. The real Tournament producer is not yet implemented;
  its submission flow will need a transactional outbox.
