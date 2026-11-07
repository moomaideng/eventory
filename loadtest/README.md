# Load tests

[k6](https://k6.io) tests that hit the system through the Traefik gateway. k6 runs in Docker on the Compose network, so nothing needs installing.

## Run

```bash
task dev:backend && task seed   # stack up and seeded
task load:baseline              # ~1 min
task load:stress                # ~3.5 min
task load:spike                 # ~2.5 min
task load:all                   # all three, ~7 min
```

Each run writes an HTML report to `loadtest/results/` (open it in a browser). The tests insert tournaments, so for comparable runs reset first: `task db:reset`, then the two commands above.

On Linux, set `LOCAL_UID` and `LOCAL_GID` if your user is not 1000, so results are writable.

## What the tests do

| Test | Load shape | Question |
| --- | --- | --- |
| Baseline | 20 VUs, flat | What is normal latency? Reference for the others. |
| Stress | Steps of 50, 100, 150, 200, 300 VUs | Where does it start to degrade? |
| Spike | 10 VUs, jump to 300 in 5 s, back to 10 | Does it survive a surge and recover? |

A VU is a simulated user who sends a request, waits 0.3-0.7 s, and repeats. Each request is one of:

| Share | Request | Hits |
| --- | --- | --- |
| 30% | list tournaments | Tournament + Postgres |
| 22% | tournament detail | Tournament + Postgres |
| 12% | bracket | Tournament + Postgres |
| 12% | sponsorship tiers | Tournament + Postgres |
| 22% | `GET /accounts/me` | Account + Postgres |
| 2% | create tournament | Tournament, Account (gRPC), Postgres |

Pass criteria: p95 latency under 500 ms and under 1% failed requests. Results are split by phase (stress level, or warmup/spike/recovery), and failures are counted by HTTP status (`fail_status_*`).

## Results

Single laptop, dev containers, k6 on the same machine. Compare steps with each other, not as absolute capacity.

| Test | | Before pool cap | After pool cap |
| --- | --- | --- | --- |
| Stress (to 300 VUs) | p95 / errors | 755 ms / 3.0% | 6 ms / 0% |
| Stress, 300 VUs step | p95 / errors | 1.1 s / 10.4% | 6 ms / 0% |
| Spike, during surge | p95 / errors | 1.55 s / 9.1% | 4 ms / 0% |
| Baseline | p95 / errors | 6 ms / 0% | 5 ms / 0% |

**Finding.** Before the fix, latency climbed from about 100 VUs and requests failed (HTTP 500/502) from about 200. Postgres logged `FATAL: too many clients already`: the services opened unlimited connections and exhausted Postgres's limit of 100. Recovery after the spike was clean, so the system self-heals once load drops.

**Fix.** `backend/internal/database/database.go` now caps each service's pool at 20 open connections (`DB_MAX_OPEN_CONNS` to override). Keep instances x cap under Postgres `max_connections`. After the fix there were no errors at any level and throughput rose from 241 to 303 req/s.

The before and after reports are kept in `loadtest/results/` as `before-*` and `after-*`.
