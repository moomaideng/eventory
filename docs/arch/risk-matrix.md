# Risk Matrix (FMEA)

Scope: architecture and operations risks for Eventory (Traefik gateway, Account / Tournament / Payment / Notification services, Postgres, MongoDB, RabbitMQ, Supabase Auth, single-VM Docker Compose). Sources: ADR-10, 12; `backend/`; `.github/workflows/`; `docker-compose.production.yml`.

## Method

Failure Modes and Effects Analysis, as taught in lecture 01 ("A Risk Analysis in SW ARCH"). Each failure mode is scored 1 to 10:

- **Severity (S):** how bad the effect is. 1 = cosmetic, 4 = one feature degraded, 8 = core flow (register / sponsor / pay) down, 10 = total outage or data/money loss.
- **Occurrence (O):** how likely the cause is. 1 = rare, 5 = plausible within a season, 9-10 = observed / constant.
- **Detection (D):** how likely we notice before users are harmed. 1 = certain (alert or test fires), 10 = undetected until damage is done.

**RPN = S x O x D** (1 to 1000), ranked highest first. The lecture defines no thresholds, so none are set. Scores are team judgment.

## Register

| ID | Failure mode | Effect | Cause / evidence | S | O | D | RPN |
| --- | --- | --- | --- | --- | --- | --- | --- |
| F1 | No usable backup of Postgres/Mongo | Permanent loss of registrations, payments, tournaments | Named volumes only; no backup job, script or doc in the repo. Likely triggers: the manual production workflow `.github/workflows/db-management.yml` (`reset_db` defaults to `true`, runs DROP public CASCADE), disk loss, volume deletion, or a bad migration. Nobody notices a missing backup until a restore is needed | 10 | 4 | 9 | 360 |
| F2 | Tournament event lost between DB commit and broker publish | Announcement / review-decision notification emails not sent; Tournament state itself is correct. The review path discards the publish error; the announcement path only logs a warning | Tournament publishes directly after commit (`announcement_usecase.go`; `team_review_usecase.go`). ADR-10 says a transactional outbox is needed. Payment already has one, Tournament does not. No metric shows lost events | 4 | 5 | 8 | 160 |
| F3 | Single production VM is lost | Total outage until rebuilt; data loss if F1 also true | ADR-12: "host is a single point of failure"; no `replicas`/scale in compose. Outage is obvious to users but no external uptime check exists | 10 | 3 | 3 | 90 |

## Mitigation

| ID | Action | RPN after |
| --- | --- | --- |
| F1 | Scheduled `pg_dump` / `mongodump` off-host + restore drill; make `reset_db` default to `false`. Backups bound data loss (S 10 -> 6) and the drill proves they work (D 9 -> 3) | 72 |
| F2 | Reuse Payment's pattern (`payment_outbox` written in the same transaction + relay in `server/background.go`) in Tournament; alert on unpublished backlog (O 5 -> 1, D 8 -> 4) | 16 |
| F3 | VM snapshot + documented rebuild (S 10 -> 7), external uptime monitor (D 3 -> 2). Redundancy (Swarm/K8s) is deliberately deferred (ADR-12: cost and team size) | 42 |

Residual RPNs assume the actions are completed; they are not yet.
