# Observability

Grafana + Loki + Alloy, for reading structured logs without SSH-ing into app containers. Alloy scrapes Docker logs and ships them to Loki; Grafana queries Loki.

## Local

```bash
docker compose --profile infra --profile apps --profile obs up -d
```

Grafana: [http://localhost:8080/grafana](http://localhost:8080/grafana), user `admin`, password `admin` (override with `GRAFANA_PASSWORD`).

Go services log JSON in production by default; locally they default to logfmt. Both are structured and equally filterable, just via different LogQL parser stages. `LOGGING_FORMAT=json` only matters for Explore's JSON-aware UI, not for filtering. To switch a service to JSON locally: `backend/services/<service>/.env` (gitignored) with `LOGGING_FORMAT=json`, then `docker compose up -d --force-recreate account`.

### Explore recipe

JSON mode:

```logql
{compose_service="tournament"} | json | request_id="<id>"
```

Logfmt mode (local default, same precision):

```logql
{compose_service="tournament"} | logfmt | request_id="<id>"
```

Swap `tournament` for any scraped service. Scraping is controlled by the `eventory.logs=true` Compose label (see `docker/observability/config.alloy`); no service list here since it drifts.

`frontend` and `traefik` don't use `slog`, so their lines are unstructured. Only substring filtering applies: `{compose_service="traefik"} |= "error"`.

## Production

Grafana is public at `https://eventory.ddns.net/grafana`, not SSH-tunnel-only. Deliberate for this course project: it has a real login (`PROD_GRAFANA_PASSWORD`, see `docs/deployment.md`), and being reachable without an SSH hop outweighs keeping it private.
