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

## Project & Routing Structure

We use Next.js **Route Groups** (`(public)` and `(auth)`) alongside App Router boundaries:

```text
frontend/
├── proxy.ts                          # Next.js 16 Proxy (Asymmetric JWT verification & session refresh)
│
├── app/
│   ├── (auth)/                       # Auth Flow (Minimal Header with Logo only)
│   │   ├── layout.tsx                # Auth layout
│   │   ├── loading.tsx               # Instant card skeleton loading boundary
│   │   ├── login/
│   │   │   ├── page.tsx              # Server Component (Verified Auth redirect check)
│   │   │   └── login-form.tsx        # Clean Google Sign-In card with Spinner
│   │   └── onboarding/
│   │       ├── page.tsx              # Server Component (Verified Auth redirect check)
│   │       ├── actions.ts            # Server Action (Direct Go backend onboarding)
│   │       └── onboarding-form.tsx   # React 19 Native useActionState Component
│   │
│   ├── (public)/                     # Public & App Views (Full Navbar with Role Switcher)
│   │   ├── layout.tsx                # Public layout with Navbar
│   │   ├── loading.tsx               # Instant public skeleton loading boundary
│   │   └── page.tsx                  # Minimal Landing Hero & CTA buttons (Server Component)
│   │
│   ├── api/
│   │   └── auth/callback/route.ts    # Supabase OAuth PKCE code exchange Route Handler
│   │
│   ├── error.tsx                     # Global App Router Error Boundary
│   ├── not-found.tsx                 # Branded 404 Not Found Page
│   ├── layout.tsx                    # Root HTML layout (Fonts, globals.css, QueryProvider, RoleProvider)
│   └── globals.css                   # Tailwind CSS v4 & theme variables
│
├── components/
│   ├── navbar.tsx                    # Header with Base UI DropdownMenu Role Switcher & Dynamic Nav Links
│   ├── providers/
│   │   └── query-provider.tsx        # TanStack QueryClient Provider wrapper
│   └── ui/                           # Pure shadcn/ui Base UI primitives
│       ├── alert.tsx
│       ├── avatar.tsx
│       ├── button.tsx
│       ├── card.tsx
│       ├── dropdown-menu.tsx
│       ├── field.tsx
│       ├── input.tsx
│       ├── label.tsx
│       ├── separator.tsx
│       ├── skeleton.tsx
│       └── spinner.tsx
│
├── context/
│   └── role-context.tsx              # Role & Auth Context (TanStack Query + Supabase session + Go API sync)
│
├── lib/
│   ├── api/                          # Type-safe OpenAPI Client
│   │   ├── schema.d.ts               # Auto-generated types from Go Huma OpenAPI 3.1
│   │   └── client.ts                 # openapi-fetch & openapi-react-query client instances
│   ├── client.ts                     # Supabase Browser Client helper
│   ├── server.ts                     # Supabase Server Component helper (getUser & cookies)
│   ├── proxy-session.ts              # Supabase Session Proxy helper (getClaims & dev fallback)
│   └── utils.ts                      # Tailwind class merge helper (`cn`)
│
└── .agents/skills/
    ├── shadcn/                       # shadcn/ui Best Practices & Rules
    └── supabase/                     # Supabase Auth, SSR, & Database Rules
```

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

## Getting Started

Frontend development is orchestrated from the repository root so it shares one local environment file with the backend and Docker Compose.

### 1. Prerequisite

Install Docker Desktop. Node.js is only required for running frontend tooling directly on the host.

### 2. Configure Environment Variables
Create the single ignored environment file at the repository root:
```bash
cp .env.example .env
```

The frontend receives `SUPABASE_URL`, `SUPABASE_PUBLISHABLE_KEY`, and `API_URL` directly. `next.config.ts` explicitly exposes these public values to browser code. In Docker, server actions use the private `INTERNAL_API_URL=http://backend:8080` while browser code uses `API_URL`.

> **Note on Local Dev Mode:**
> You can develop and test UI features immediately without setting up Supabase keys. Click **"Dev Quick Login"** on the Navbar to simulate logged-in states and test role switching between Competitor, Organizer, and Sponsor modes.

### 3. Sync OpenAPI Types from Backend
When the Go backend is running, sync the latest TypeScript types by running:
```bash
cd frontend && npm run openapi:generate && cd ..
```

### 4. Run Development Server
```bash
make frontend
```

This delegates to Docker Compose and starts the frontend plus its required backend and database dependencies.
Next.js runs in development mode, so saved frontend source changes use Hot Module Reloading automatically.

Open [http://localhost:3000](http://localhost:3000) in your browser.
