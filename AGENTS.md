# Eventory Agent Guidelines

Before writing or editing code in this repository, you **MUST** consult the embedded skills and local documentation directly:

- **Frontend (`frontend/`):** Read [`frontend/AGENTS.md`](frontend/AGENTS.md). It mandates reading `frontend/node_modules/next/dist/docs/` for Next.js 16 APIs and `frontend/.agents/skills/` for shadcn Base UI and Supabase SSR rules.
- **Backend (`backend/`):** Built with Go Huma v2 and GORM. Run `npm --prefix frontend run openapi:generate` whenever API endpoints change to keep frontend types in sync.
