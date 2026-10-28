# ADR-13: Observability Stack

## Context

A request now crosses the gateway, several services, and their databases (ADR-01, ADR-05, ADR-07). When something is slow or fails, the symptom often appears in a different service than the cause, and per-container logs cannot show the whole path. The stack must fit single VM and should not tie our instrumentation to one backend.

Options considered:

- **Grafana stack (Loki, Tempo, Grafana, Alloy) with OpenTelemetry**: low memory use, one UI for logs and traces, trace-to-log links, OTLP is a standard. Several moving parts to configure.
- **Elastic stack (Elasticsearch, Kibana, APM)**: mature and full-featured, but Elasticsearch needs far more memory than our VM budget allows.
- **Jaeger and Prometheus with a separate log tool**: each tool is good alone, but logs and traces live in different UIs with no built-in correlation.
- **Managed SaaS (Grafana Cloud, Datadog)**: no servers to run, but sends data off the host, adds an external dependency and a cost.
- **Logs only**: simplest, but cannot show a request across services.

## Decision

Use the **Grafana stack with OpenTelemetry**.

- Services emit OpenTelemetry traces to a collector (**Alloy**) that stores them in **Tempo**; Alloy also collects container logs into **Loki**. **Grafana** is the single UI for both.
- Trace context travels with every call, whether HTTP, gRPC, or a broker message, so one request is one trace from the gateway down to the databases.
- Log entries carry the trace ID, so a trace and its logs can be opened from each other.
- Only Grafana is exposed outside the Compose network, behind its own login. Data is kept briefly (72 hours) to fit the VM.

## Status

Accepted

## Consequences

**Positive**

- One trace shows gateway, services, and database queries for a request, and its log entries are one click away (NFR9).
- OpenTelemetry keeps the services independent of Tempo and Loki.
- Fits the memory budget and is simple to run locally.

**Negative**

- Four more containers to run and monitor in production.
- No metrics or alerting yet, so problems are found by looking, not by being told.
- Traces are not sampled, so storage grows with traffic.
- Retention is only 72 hours.
- Trace-to-log links only cover log lines written while handling a request.
- Grafana is reachable from the internet, so its login is part of the attack surface.
