# Troubleshooting

Find your error message, apply the fix.

---

## Quick table

| Symptom | Cause | Fix |
| --- | --- | --- |
| `Module not found: Can't resolve '...'` | Your packages are older than the code | `npm --prefix frontend ci` |
| Same error, but running in Docker | Docker keeps its **own** copy of packages | See [Stale frontend packages](#the-two-stale-package-traps) |
| `violates foreign key constraint` while seeding | Your database holds rows from an older version of the code | `task db:reset`, then `task dev` or `task apps`, then `task seed` |
| `connection refused` on port 5432 | The database is not running yet | `task migrate` |
| `port is already allocated` | Something else uses 3000, 8080, 8081, 8082, 9091, or 5432 | See [Port already in use](#port-already-in-use) |
| `Cannot connect to the Docker daemon` | Docker is not running | Start Docker and wait until it accepts commands |
| `database "account_db" does not exist` | The Postgres volume was created before the split | `task db:reset`, then `task dev` or `task apps` |
| `Required command not found: docker` (or `go`, `npm`) | That tool is missing, or not on your `PATH` | Install it, then open a new terminal |
| `task: command not found` | Task is not installed | [Install Task](https://taskfile.dev/installation/) |
| Tournament list is empty | Sample data was never added | `task seed` |
| Edited Go but the container did not restart | Air did not see the save, or the container is stopped | See [Air problems](#air-problems) |

---

## The two "stale package" traps

Happens whenever a teammate adds a package: you pull their code, but your computer still has the old package list.

**Important:** the host and Docker keep **separate** package folders. Fixing one does not fix the other.

```mermaid
flowchart TD
    P["package.json<br/>(the shopping list, in git)"]
    H["frontend/node_modules<br/>(your computer's copy)"]
    D["Docker volume<br/>(Docker's copy)"]
    P -->|"npm ci"| H
    P -->|"remove the volume"| D
```

| You run the project with | Fix command |
| --- | --- |
| `task dev` (frontend on the host) | `npm --prefix frontend ci` |
| `task web` (frontend in Docker) | `task down`, then `docker volume rm eventory_frontend_node_modules`, then `task web` |
| Both, sometimes | Run both commands |

**Rule of thumb:** after any `git pull` that changes `frontend/package.json`, run `npm --prefix frontend ci`. `task dev` does this when `package-lock.json` changes. `task web` does not. Its `node_modules` live in the volume.

---

## Seeding fails with a foreign key error

Full message looks like:

```
ERROR: insert or update on table "organizer_profiles"
violates foreign key constraint "fk_accounts_organizer_profile"
```

**What happened.** Your database still holds rows created by an older version of the code. The sample-data script skips accounts whose email already exists, so the new accounts never get created, and the profiles that point at them have nothing to attach to.

**Fix.** Wipe and rebuild:

```bash
task db:reset
task dev   # or task apps
task seed
```

> ⚠️ `task db:reset` deletes the local Postgres volume, including any test account you made. That is fine for development. It does not touch the server.

---

## Port already in use

These ports must be free: **3000** (frontend), **8080** (gateway), **8081** (Account HTTP), **9091** (Account gRPC), **8082** (Tournament HTTP), **5432** (database).

**Find the culprit:**

```bash
# Windows
netstat -ano | findstr ":3000 :8080 :8081 :8082 :9091 :5432"

# macOS / Linux
lsof -i :3000 -i :8080 -i :8081 -i :8082 -i :9091 -i :5432
```

**Most common cause:** the project is already running in another terminal, or leftover containers. Clear them:

```bash
task down
```

**If you have your own PostgreSQL installed** it will already hold port 5432. Start Compose on another host port:

```bash
POSTGRES_PORT=5433 task dev
```

Containers still talk to `postgres:5432` inside Compose. Host `go run` reads `DB_DSN` from each service `.env`. Point both files at `5433`, or the host process looks in the wrong place.

`backend/services/account/.env`:

```ini
DB_DSN=postgresql://admin:password123@localhost:5433/account_db?sslmode=disable
```

`backend/services/tournament/.env`:

```ini
DB_DSN=postgresql://admin:password123@localhost:5433/tournament_db?sslmode=disable
```

---

## Air problems

`task dev` and `task apps` run Air inside the Account and Tournament containers. You do not install it on your machine for that path.

**If you save a Go file and the container does not restart,** check `task logs`. Then rebuild:

```bash
task apps
```

---

## Nuclear option

When nothing makes sense, reset everything except your settings files:

```bash
task down
docker volume rm eventory_postgres_data eventory_frontend_node_modules \
                 eventory_frontend_next eventory_go_mod_cache eventory_go_build_cache
rm -rf frontend/node_modules     # Windows: rmdir /s frontend\node_modules
npm --prefix frontend ci
task dev
task seed
```

This keeps `frontend/.env.local` and any `backend/services/*/.env` overrides, and deletes everything else that can go stale.

---

## Still stuck?

Collect these before asking a teammate:

1. Whether you used `task dev` or `task web`.
2. The **first** error in the terminal, not the last one. For containers, that is `task logs`.
3. Output of `docker ps -a` and `git log --oneline -3`.
