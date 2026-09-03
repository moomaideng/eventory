<!-- BEGIN:nextjs-agent-rules -->

# This is NOT the Next.js you know

This version has breaking changes — APIs, conventions, and file structure may all differ from your training data. Read the relevant guide in `node_modules/next/dist/docs/` (resolved from this file's directory; in monorepos the `next` package may not be visible from the repo root) before writing any code. Heed deprecation notices.

This block is written and re-added by `next dev` — verify at `node_modules/next/dist/server/lib/generate-agent-files.js`. Removing it from a diff only re-creates the uncommitted change; committing it with your work keeps the tree clean.

<!-- END:nextjs-agent-rules -->

---

# Eventory Rules for Agents

Before writing, modifying, or refactoring any code, you **MUST** read the source documentation and embedded skill rules directly:

1. **Next.js 16 APIs & Breaking Changes:**
   - **Do not rely on training weights.** Read the relevant guides in `node_modules/next/dist/docs/` directly before implementing routes, proxies, boundaries, or server components.
   - Key differences: Next.js 16 uses `proxy.ts` (not `middleware.ts`), and all request APIs (`cookies()`, `headers()`, `params`, `searchParams`) are asynchronous Promises.

2. **UI & Components (`shadcn/ui`):**
   - **Always read `.agents/skills/shadcn/` before writing UI.**
   - We use Base UI (`base-vega` style). Inspect `.agents/skills/shadcn/rules/` for `render` prop rules (never use Radix `asChild`), `data-icon` button attributes, `<FieldGroup>` forms, `<Skeleton>` loaders, and semantic color tokens.

3. **Authentication & SSR (`supabase`):**
   - **Always read `.agents/skills/supabase/SKILL.md` before touching auth or sessions.**
   - Always verify sessions cryptographically on the server using `supabase.auth.getUser()`.

4. **Verification & Code Quality:**
   - Run `npm run lint` and `npx tsc --noEmit` before finishing. `npm run format` is optional for formatting cleanliness.
