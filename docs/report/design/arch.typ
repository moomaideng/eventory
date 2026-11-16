#import "../appendix/preambles.typ": *

== System Architecture <arch>

#figure(
  image("../img/architecture-diagram.png", width: 100%),
  caption: [Eventory system architecture],
)

- *Entry & Auth:* Web Client signs users in via Supabase Auth and routes API requests through Traefik API Gateway.
- *Services & Data:* Account, Tournament, Payment, and Notification services each own a `PostgreSQL` database, with `MongoDB` for dynamic forms and S3-compatible `Object Storage` for sponsor logos.
- *Communication:* `Tournament Service` calls `Account` and `Payment` over gRPC using `BatchGetAccounts`, `CreatePayment`, and `RefundRegistration`, and publishes events to the `Message Broker` for `Notification Service`.
- *Observability:* Services ship traces and logs to `Alloy`, which writes them to `Tempo` and `Loki` for inspection in `Grafana`.
