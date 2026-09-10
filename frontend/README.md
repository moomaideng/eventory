# Eventory Frontend

Frontend client for **Eventory**, a competition and tournament management platform supporting bracket management, crowdfunding prize pools, team lobbies, and contextual role switching.

---

## Tech Stack

- **Framework:** Next.js 16 (App Router, React 19, TypeScript)
- **UI & Design System:** Tailwind CSS v4 + **shadcn/ui** (Base UI style) + Lucide Icons
- **API & Server State:** `openapi-fetch` + `@tanstack/react-query` / `openapi-react-query` (Contract-first type safety synced with Go Huma)
- **Authentication & Storage:** Supabase Auth (`@supabase/ssr`) with Google OAuth & Supabase Storage
- **State & Context:** React Context (`RoleProvider` with contextual role switching & Dev Mock state) + TanStack Query cache

---

## Core Identity & Contextual Role Model

Eventory implements a **Single Primary Account** architecture allowing a user to toggle between 3 distinct contextual roles without maintaining separate logins:

1. **Competitor Mode (Default):** Browse tournaments, join team lobbies, and compete.
2. **Organizer Mode:** Host tournaments, configure crowdfunding campaigns, and manage staff/referees (1 profile per account).
3. **Sponsor Mode:** Browse crowdfunding campaigns, pledge prize pools, and manage brand assets (1 profile per account).

---

## Project Structure

Eventory follows a **Feature-Driven Architecture** combined with Next.js App Router conventions:

```text
frontend/
├── app/                  # Next.js App Router (Routing, Layouts, Server Component data fetching)
│   ├── (auth)/           # Authentication route group (Login, minimal layout)
│   ├── (public)/         # Main application views (Hub, Lobbies, Tournaments, etc.)
│   └── api/              # Edge API Route Handlers (OAuth callback, webhooks)
│
├── features/             # Domain-specific feature modules
│   └── <feature>/        # e.g., auth, hub, lobbies, tournaments
│       ├── components/   # Feature-specific client/server UI components
│       ├── hooks/        # Feature custom hooks & client logic
│       └── utils/        # Feature helper functions & formatters
│
├── components/           # Shared cross-feature UI components
│   ├── ui/               # shadcn/ui design system primitives (Base UI)
│   └── providers/        # Root client providers (QueryClient, ThemeProvider)
│
├── context/              # Global React contexts (Auth session, role state)
├── lib/                  # Shared utilities, client helpers, API connectors
│   ├── api/              # Contract-first OpenAPI client & generated schema types
│   ├── auth-redirect.ts  # Route redirect sanitization & destination labeling
│   └── proxy-session.ts  # Edge session verification helpers
│
└── proxy.ts              # Next.js 16 Edge Proxy (Centralized auth route guard)
```

### Architectural Principles

1. **Routing Layer (`app/`):** Purely handles URL routing, layouts, metadata, and Server Component data orchestration. Keep complex UI and client-side state out of `page.tsx`.
2. **Feature Modules (`features/<domain>/`):** Self-contained domains encapsulating interactive components, custom hooks, and domain logic. Pages in `app/` import from `@/features/<feature>/components`.
3. **Shared Primitives (`components/ui/` & `lib/`):** Domain-agnostic design system components (shadcn Base UI) and foundational utilities reusable across features.

---

## Developer & Agent Skills (`.agents/skills/`)

Before authoring or refactoring frontend code, consult the embedded skills in `.agents/skills/` and the rules in [`AGENTS.md`](AGENTS.md):

- **`shadcn` (`.agents/skills/shadcn/`):** Contains official guidelines for Base UI (`base-vega`) composition, `render` prop usage, `data-icon` attributes, `<FieldGroup>` forms, `<Skeleton>` loaders, and semantic styling tokens.
  ```bash
  npx shadcn@latest add <component>   # Install new Base UI primitive
  npx shadcn@latest docs <component>  # View component usage and API
  ```
- **`supabase` (`.agents/skills/supabase/`):** Contains Supabase SSR rules, cryptographic `getUser()` server validation, and session proxying.
- **`AGENTS.md` ([`AGENTS.md`](AGENTS.md)):** Mandatory instructions for AI assistants pointing to local docs (`node_modules/next/dist/docs/`) and skills.
- **Code Quality:** Run `npm run lint` to lint and `npm run format` for Prettier formatting (optional for code cleanliness).

---

## Getting Started (Native Development)

You can run the Next.js frontend entirely natively from within this `frontend/` directory using standard `npm` commands.

### 1. Prerequisite

* **Node.js:** Version 24 or higher.
* **Backend Service (Optional for UI prototyping):** Running at `http://localhost:8080` when syncing contracts or testing integrated API flows.

### 2. Configure Environment Variables

Create your local environment file inside this directory (or duplicate it manually):

```bash
cp .env.example .env.local
```

The browser receives `NEXT_PUBLIC_SUPABASE_URL`, `NEXT_PUBLIC_SUPABASE_PUBLISHABLE_KEY`, and `NEXT_PUBLIC_API_URL`. In Docker, server-side requests use the private `INTERNAL_API_URL=http://backend:8080`, while browser code uses `NEXT_PUBLIC_API_URL`.

> **Note on Local Dev Mode:**
> You can develop and test UI features immediately without setting up Supabase keys or running the backend. Click **"Dev Quick Login"** on the Navbar to simulate authenticated states and test role switching between Competitor, Organizer, and Sponsor modes.

### 3. Install Dependencies

```bash
npm install
# or for clean/strict lockfile installation:
npm ci
```

### 4. Run Development Server

Start the Next.js development server:

```bash
npm run dev
```

Open [http://localhost:3000](http://localhost:3000) in your browser. Next.js Fast Refresh automatically reloads pages when code changes.

### 5. Sync OpenAPI Types from Backend

When the Go backend API is running at `http://localhost:8080`, generate end-to-end TypeScript types into `lib/api/schema.d.ts`:

```bash
npm run openapi:generate
```

### 6. Code Quality & Typechecking

Before committing changes, run:

```bash
npm run lint         # ESLint check
npm run typecheck    # Strict TypeScript verification (tsc --noEmit)
npm run format       # Optional: Prettier code formatting
```

---

## Alternative: Root Make Commands (macOS / Linux)

If you are on macOS, Linux, or WSL and prefer orchestrating from the repository root using Make:

```bash
make frontend        # Run only frontend with Next.js
make test            # Run lint, typecheck, and backend tests
make frontend-deps   # Refresh Docker node_modules volume after adding packages
```
