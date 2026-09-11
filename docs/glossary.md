# Glossary

Plain-language meanings for the words used in this project. No software background needed.

Jump to: [Architecture](#architecture) · [Frontend](#frontend) · [Backend](#backend) · [Tools and Workflow](#tools-and-workflow)

## Architecture

| Word | Plain meaning |
| --- | --- |
| **Frontend** | The part of the website you see and click. Lives in `frontend/`. |
| **Backend** | The hidden server that does the real work (checks rules, saves data). Lives in `backend/`. |
| **Database** | Where information is stored permanently (accounts, tournaments, lobbies). We use **PostgreSQL** ("Postgres"). |
| **Supabase** | A third-party service we use for "Sign in with Google" and for storing uploaded images. It does *not* store our tournament data locally; that is in Postgres. In production the database is also hosted on Supabase. |

## Frontend

| Word | Plain meaning |
| --- | --- |
| **Next.js** | The main toolkit that builds our web pages. Version 16. |
| **React** | The library Next.js is built on. Lets us build the page out of small reusable pieces ("components"). |
| **TypeScript** | JavaScript with type checking, so many mistakes are caught before running. |
| **Tailwind CSS** | A way to style pages by adding short class names (like `text-red-500`) instead of writing separate style files. |
| **shadcn/ui** | A set of ready-made, good-looking building blocks (buttons, forms, dialogs) that we copy into `components/ui/`. |
| **Lucide** | The icon set we use. |
| **TanStack Query (react-query)** | Handles fetching data from the backend and remembering it so pages feel fast. |
| **openapi-fetch / openapi:generate** | The backend publishes a "menu" of its endpoints (OpenAPI). This tool reads that menu and generates TypeScript types so the frontend can't call the backend wrongly. Run `npm run openapi:generate` after backend endpoints change. |
| **App Router** | Next.js rule: each folder inside `app/` becomes a URL. `app/tournaments/page.tsx` = `/tournaments`. |
| **Route group** | Folders like `(organizer)` in `app/`. The parentheses mean "group these pages together but don't add this name to the URL". |
| **Dev Quick Login** | A fake login button for development, so you can test pages without real Google login. |
| **ESLint / Prettier** | Tools that check code style and auto-format it. |
| **npm** | Node's package manager. Installs the libraries listed in `package.json`. |

## Backend

| Word | Plain meaning |
| --- | --- |
| **Go (Golang)** | The programming language of the backend. |
| **Huma** | A Go toolkit that defines API endpoints and automatically writes documentation for them (visible at `http://localhost:8080/docs`). |
| **Chi** | A small Go library that decides which code runs for which URL. |
| **GORM** | A Go library that translates Go code into database commands, so we rarely write raw SQL. |
| **API / endpoint** | An "API" is the list of things the backend can do. An "endpoint" is one of them, e.g. `GET /api/v1/tournaments` = "give me the tournament list". |
| **REST** | The common style of API where each URL is a thing and the verb (GET, POST, PATCH, DELETE) says what to do with it. |
| **OpenAPI** | A standard file format that describes every endpoint. Huma generates it for us at `/openapi.json`. |
| **Migration** | A script that creates or updates the database tables. `go run ./cmd/migrate`. Adding `--reset` wipes everything and rebuilds. |
| **Seed** | A script that fills the database with fake sample data for testing. `go run ./cmd/seed`. |
| **Air** | A helper that restarts the Go server automatically every time you save a file ("live reload"). |
| **Viper** | The library that reads settings from `backend/.env`. |
| **JWKS / JWT** | After Google login, Supabase gives the browser a signed "ticket" (JWT). The backend checks the signature using public keys (JWKS) to trust who you are. |
| **Ports & Adapters** | A code-organizing style: business rules in the middle (`usecases`), database and web stuff on the outside (`repositories`, `handlers`). |
| **Foreign key** | A database rule saying "this row must point at a row that exists". Stops bad data being saved. |

## Tools and Workflow

Everything around the code: how you run it, and how you share it with the team.

### Running the project

| Word | Plain meaning |
| --- | --- |
| **Docker** | Runs programs inside isolated "containers" so everyone gets the same setup no matter their OS. |
| **Docker Compose** | Starts several containers together from one file (`docker-compose.yml`). Ours has 4 services: `postgres`, `migrate`, `backend`, `frontend`. |
| **Container / image** | An image is the recipe; a container is a running copy of it. |
| **Volume** | Docker storage that survives when a container stops (e.g. `postgres-data` keeps your database). |
| **Makefile / make** | A file of named shortcuts. `make dev` runs a longer command so you don't have to type it. Needs the `make` program installed. |
| **Native** | Running the program directly on your computer, not inside Docker. |
| **.env file** | A plain text file of settings (passwords, URLs). Never committed to git. Copy from `.env.example`. |
| **localhost:3000 / :8080** | Addresses on your own computer. 3000 = frontend website, 8080 = backend API. |
| **Port** | A numbered "door" a program listens on. Two programs cannot share one port. |
| **Health check** | A tiny URL (`/health`) that answers "I'm alive" so Docker knows the service started. |
| **CI/CD** | Automation on GitHub: every pull request gets tested, and merging to `main` deploys to the Oracle server. |
| **GHCR** | GitHub Container Registry, where built Docker images are stored for deployment. |
| **node_modules** | The folder holding every downloaded frontend package. Not in git, so each person builds their own with `npm ci`. Docker keeps a separate copy. |
| **Stale install** | Your `node_modules` is older than the code, so a package the code needs is missing. Cause of most `Module not found` errors. |
| **Idempotent** | Safe to run more than once. `make seed` is meant to be; `make reset` deliberately is not. |

### Git: our team rules

General git is assumed. Only the rules specific to this project are listed.

| Word | Our rule |
| --- | --- |
| **Conventional Commits** | Commit messages start with `feat:`, `fix:`, `refactor:` or `chore:`. |
| **Rebase** | Rebase onto the newest `main` before opening a PR, so your work sits on current code. |
| **Squash and Merge** | Our merge style. All commits of a PR become one, keeping history tidy. |
| **CI checks** | Tests, lint and build run automatically on every PR. Red blocks the merge. |

**Our cycle:** branch → commit → push → PR → approval → squash and merge.
