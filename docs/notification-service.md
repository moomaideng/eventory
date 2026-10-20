# Notification Service

The Notification Service is an independently deployed Go worker. It consumes
`registration.submitted` events from RabbitMQ, sends an English confirmation
email through SMTP, and records delivery state in `notification_db`.

For team registrations the recipient is the captain. For solo registrations
the recipient is the applicant. A submitted registration is still waiting for
organizer review; the email does not claim that it has been accepted.

## Local smoke test

Start the application stack, publish a sample event, then open Mailpit:

```bash
task apps
task notification:smoke
```

- Mailpit: <http://localhost:8025>
- RabbitMQ management: <http://localhost:15672> (`eventory` / `eventory` locally)
- Notification readiness: <http://localhost:8084/ready>

The smoke publisher is development tooling. It is deliberately not connected
to roster locking because a locked roster is not yet a submitted registration.
The Tournament Service must eventually publish the event only after its real
registration transaction commits. The producer must use a transactional outbox:
the registration update and outbox row commit together, then a relay publishes
with confirms. A direct "commit, then publish" sequence has a crash window and
does not provide end-to-end at-least-once delivery.

## Event contract

The versioned Go contract is `internal/events.RegistrationSubmitted`. The
producer supplies the email recipient in the event so Notification does not
add a synchronous Account Service dependency. Do not log the payload because
it contains personal data.

RabbitMQ resources:

| Resource | Value |
| --- | --- |
| Domain exchange | `eventory.domain.v1` |
| Routing key | `registration.submitted` |
| Consumer queue | `notification.email.registration-submitted.v1` |
| Retry exchange | `eventory.notification.retry.v1` |
| Dead-letter queue | `notification.email.dlq.v1` |

## Reliability policy and references

The implementation separates externally supported practices from project
policy:

| Decision | Source | Status / rationale |
| --- | --- | --- |
| Manual consumer acknowledgements, persistent messages, and publisher confirms | [RabbitMQ Reliability Guide](https://www.rabbitmq.com/docs/reliability) and [Acknowledgements and Confirms](https://www.rabbitmq.com/docs/confirms) | Referenced reliability mechanism for at-least-once processing. |
| Transactional outbox for Tournament publishing | [AWS Transactional Outbox pattern](https://docs.aws.amazon.com/prescriptive-guidance/latest/cloud-design-patterns/transactional-outbox.html) | Referenced mechanism that closes the database-commit/broker-publish dual-write gap. It becomes required when the real `SubmitRegistration` producer is implemented. |
| TTL retry queues and dead-letter exchanges | [RabbitMQ Dead Letter Exchanges](https://www.rabbitmq.com/docs/dlx) and [TTL](https://www.rabbitmq.com/docs/ttl) | Referenced broker mechanism. RabbitMQ does not prescribe our delay values. |
| Increasing delays, bounded attempts, and jitter as a general retry strategy | [AWS Well-Architected REL05-BP03](https://docs.aws.amazon.com/wellarchitected/latest/framework/rel_mitigate_interaction_failure_limit_retries.html) | Referenced general principle. This implementation uses fixed TTL queues, not jitter, to keep the topology observable for the course project. |
| Retry delays `30s,5m,30m` | Eventory project choice | Chosen to recover quickly from a short local/provider interruption, back off for a continuing outage, and still make the final result observable during a demo. They are configurable through `RETRY_DELAYS` and are not a RabbitMQ or SMTP standard. |
| Three retries (four total attempts) | Eventory project choice | A bounded retry budget prevents an indefinite queue backlog and fits the project demo. Tune from production delivery metrics. |
| Consumer prefetch `1` | Eventory project choice | Matches the current serial worker. Increase it only when a bounded concurrent worker pool is implemented and tested. |
| Processing lease `20s` | Eventory project choice | Longer than the chosen `10s` SMTP timeout but shorter than the first `30s` retry, so a replacement consumer can reclaim work after a crash. |
| SMTP timeout `10s` | Eventory project choice | Keeps a slow SMTP relay from holding a delivery indefinitely while allowing normal TLS and SMTP handshakes. |
| RabbitMQ reconnect delay `5s` | Eventory project choice | Avoids a tight reconnect loop while restoring the consumer quickly after a local broker restart. |
| SMTP `4xx` retries and `5xx` permanent failure | [RFC 5321 reply-code semantics](https://www.rfc-editor.org/rfc/rfc5321.html) | `4yz` means temporary negative completion and `5yz` means permanent negative completion. |
| SMTP sender generally retries remote delivery for much longer intervals | [RFC 5321 section 4.5.4.1](https://www.rfc-editor.org/rfc/rfc5321.html#section-4.5.4.1) | Applies to the SMTP relay after it accepts the message. Eventory only retries submission to its configured relay. |

SMTP cannot provide an application idempotency key. The unique delivery row
and deterministic `Message-ID` suppress normal duplicate events, but a process
crash after SMTP accepts a message and before PostgreSQL records `SENT` can
still produce a duplicate. This is an explicit at-least-once limitation.

The delivery identity is `(event_id, channel, template_key)`. A SHA-256 hash of
the normalized event detects a producer that incorrectly reuses an event ID
with different content; such a message is moved to the DLQ without sending.

## Dead-letter redrive

`FAILED` is terminal during normal consumption, so blindly republishing a DLQ
message is intentionally skipped. After correcting the underlying cause, reset
exactly one failed event and then republish that event's DLQ message:

```bash
go -C backend run ./services/notification/cmd/migrate -redrive-event EVENT_UUID
```

For the production image, run `./migrate -redrive-event EVENT_UUID` in a
one-shot `notification-migrate` container. The reset is visible in the delivery
row's timestamps. RabbitMQ management is still used to republish and remove the
matching DLQ message; automated bulk redrive is deliberately out of scope for
this first slice.

## Configuration

See `backend/services/notification/.env.default` for all defaults. Production
must provide its own RabbitMQ password and SMTP credentials. RabbitMQ and the
Notification health endpoint stay internal and are not routed through Traefik.
