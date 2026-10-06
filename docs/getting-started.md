# Getting Started

How to set up Eventory and run it on your machine.

---

## 1. Install these first

| Tool | Version | Why |
| --- | --- | --- |
| **[Docker](https://www.docker.com)** | any recent | Runs Postgres, Traefik, and the backend services |
| **[Go](https://go.dev/doc/install)** | 1.26+ | Tests and `task seed` |
| **[Node.js](https://nodejs.org/en/download)** | 24+ | Runs the frontend |
| **[Task](https://taskfile.dev/installation)** | 3+ | [taskfile.dev](https://taskfile.dev/installation/) shortcuts such as `task dev` |

```bash
docker --version && go version && node --version && task --version
```

---

## 2. Frontend settings

```bash
cp frontend/.env.example frontend/.env.local
```

Backend defaults are already in git:

- `backend/services/account/.env.default`
- `backend/services/tournament/.env.default`

Add `backend/services/account/.env` or `backend/services/tournament/.env` only when you need to override a default. Those files are not committed. Real Supabase keys are only required to test Google login.

---

## 3. Run it

```bash
task dev
```

That starts Postgres and Traefik (`infra`) and Account and Tournament (`apps`) in Docker, then runs Next.js on your machine. It migrates Account, then Tournament, before either API starts. The first build compiles those migrate commands and both services, and can take a few minutes. Later starts reuse the Go build cache.

Stop Next.js with Ctrl+C. The containers keep running until `task down`.

To run the frontend in Docker as well:

```bash
task web
```

---

## 4. Open it

| URL | What |
| --- | --- |
| http://localhost:3000 | The website |
| http://localhost:8080/health | Gateway. Empty 204 means Account is up |
| http://localhost:8081/docs | Account endpoints |
| http://localhost:8082/docs | Tournament endpoints |

The website calls `http://localhost:8080`. Traefik sends `/api/v1/accounts` to Account, and `/api/v1/tournaments` and `/api/v1/lobbies` to Tournament.

No Supabase keys? Use the **Dev Quick Login** button in the navbar, then switch modes from the avatar menu.

---

## 5. Sample data

Sample rows are separate from migrate. Postgres must already be running (`task dev` or `task migrate`):

```bash
task seed
```

Account is seeded first, then Tournament.

---

## Daily routine

```bash
task dev     # write code
task test    # before a pull request
```

`task test` runs `go test ./...` in `backend/`, then frontend lint and typecheck.

**Changed a backend endpoint?** With the apps up (`task apps`):

```bash
npm --prefix frontend run openapi:generate
```

That merges both OpenAPI documents into `frontend/lib/api/schema.d.ts`.

---

## Command cheat sheet

| Task | Does |
| --- | --- |
| `task dev` | Infra and apps in Docker, frontend on the host |
| `task web` | Infra, apps, and the frontend container |
| `task apps` | Rebuild and restart infra and apps |
| `task logs` | Follow infra and app logs |
| `task down` | Stop containers. The database volume stays |
| `task migrate` | Apply `account_db`, then `tournament_db` |
| `task seed` | Sample accounts, then sample tournaments |
| `task db:reset` | Delete the Postgres volume and start infra and apps again |
| `task frontend` | Host Next.js only |
| `task test` | Backend tests, frontend lint, frontend typecheck |

To run a service on the host, use `task migrate` first. That starts Postgres and applies both schemas. It does not start the API containers.

```bash
go -C backend run ./services/account/cmd/api
go -C backend run ./services/tournament/cmd/api
```

`8081`, `9091`, and `8082` must be free, so do not leave `task apps` running. Traefik on `:8080` forwards to the containers, not to these processes. Host runs read `.env.default`, so they use `localhost` and `127.0.0.1:9091`.

---

Something not working? → **[Troubleshooting](troubleshooting.md)** \
Word you don't know? → **[Glossary](glossary.md)**
