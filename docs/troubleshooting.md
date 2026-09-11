# Troubleshooting

Find your error message, apply the fix.

---

## Quick table

| Symptom | Cause | Fix |
| --- | --- | --- |
| `Module not found: Can't resolve '...'` | Your packages are older than the code | `npm --prefix frontend ci` |
| Same error, but running in Docker | Docker keeps its **own** copy of packages | `make frontend-deps` |
| `violates foreign key constraint` while seeding | Your database holds rows from an older version of the code | `make reset && make seed` |
| `connection refused` on port 5432 | The database is not running yet | `make db` first |
| `port is already allocated` | Something else uses 3000 / 8080 / 5432 | See [Port already in use](#port-already-in-use) |
| `Cannot connect to the Docker daemon` | Docker Desktop is not running | Open Docker Desktop, wait for the whale icon |
| `Missing backend/.env. Copy ...` | Settings files were never created | See [step 2 of Getting Started](getting-started.md#2-create-your-settings-files) |
| `Required command not found: docker` (or `go`, `npm`) | That tool is missing, or not on your `PATH` | Install it, then open a new terminal |
| `Air is not installed. Run 'make install' first.` | The live-reload tool was never installed | See [Air problems](#air-problems) |
| Tournament list is empty | Sample data was never added | `make seed` |
| Edited code but nothing changes (Docker) | Container has a stale copy | `make down && make dev-docker` |

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
    P -->|"make frontend-deps"| D
```

| You run the project with | Fix command |
| --- | --- |
| Way 2 or Way 3 (native) | `npm --prefix frontend ci` |
| Way 1 (all Docker) | `make frontend-deps` |
| Both, sometimes | Run both commands |

**Rule of thumb:** after any `git pull` that changes `frontend/package.json`, run `make install`.

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
make reset
make seed
```

> ⚠️ `make reset` deletes everything in your local database, including any test account you made. That is fine for development. Never run it against production.

---

## Port already in use

Three ports must be free: **3000** (frontend), **8080** (backend), **5432** (database).

**Find the culprit:**

```bash
# Windows
netstat -ano | findstr ":3000 :8080 :5432"

# macOS / Linux
lsof -i :3000 -i :8080 -i :5432
```

**Most common cause:** the project is already running in another terminal, or leftover containers. Clear them:

```bash
make down
```

**If you have your own PostgreSQL installed** it will already hold port 5432. Change the port in `backend/.env`:

```ini
POSTGRES_PORT=5433
DB_DSN=postgresql://admin:password123@localhost:5433/app_db?sslmode=disable
```

Change **both** lines, or the backend will look in the wrong place.

---

## Air problems

`air` restarts the backend automatically when you save a file.

**Almost always, it simply was never installed.** Fix:

```bash
make install
```

Plain equivalent: `go install github.com/air-verse/air@v1.67.3`

**If it is installed but typing `air` still says "command not found",** Go's bin folder is not on your `PATH`. Only Way 3 needs this.

| System | Add to `PATH` |
| --- | --- |
| Windows | `%USERPROFILE%\go\bin` |
| macOS / Linux | `~/go/bin` |

Open a new terminal afterwards. Don't want to bother? Use `go run ./cmd/api` instead, but then you must restart it yourself after every edit.

---

## Nuclear option

When nothing makes sense, reset everything except your settings files:

```bash
make down                        # stop all containers
docker volume rm eventory_postgres-data eventory_frontend-node-modules \
                 eventory_frontend-next eventory_backend-air-tmp
rm -rf frontend/node_modules     # Windows: rmdir /s frontend\node_modules
make install
make db
make reset && make seed
make dev
```

This keeps `backend/.env` and `frontend/.env.local`, and deletes everything else that can go stale.

---

## Still stuck?

Collect these before asking a teammate:

1. Which way you ran it (1, 2, or 3).
2. The **first** error in the terminal, not the last one.
3. Output of `docker ps -a` and `git log --oneline -3`.
