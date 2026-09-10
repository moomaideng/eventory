<!-- BEGIN:nextjs-agent-rules -->

# This is NOT the Next.js you know

This version has breaking changes — APIs, conventions, and file structure may all differ from your training data. Read the relevant guide in `node_modules/next/dist/docs/` (resolved from this file's directory; in monorepos the `next` package may not be visible from the repo root) before writing any code. Heed deprecation notices.

This block is written and re-added by `next dev` — verify at `node_modules/next/dist/server/lib/generate-agent-files.js`. Removing it from a diff only re-creates the uncommitted change; committing it with your work keeps the tree clean.

<!-- END:nextjs-agent-rules -->

---

# Eventory Rules for Agents

Before writing, modifying, or refactoring any code, you **MUST** read the source documentation and embedded skill rules directly:

1. **Architecture & File Boundaries:**
   - **`page.tsx` & `layout.tsx` MUST remain Server Components.** Never add `'use client'`. Reserve them strictly for async `params`/`searchParams`, metadata, direct data fetches, and `<Suspense>` orchestration.
   - **Push `'use client'` down** to interactive leaf components in `features/<feature>/components/` (imported via `@/features/...`).
   - **Modular design (~240 lines guideline):** Keep files focused and maintainable. Split complex logic into hooks, sub-components, or feature modules without over-fragmentation.
   - **Path aliases:** Use `@/*`. Do not climb directories (`../../`).

2. **Next.js 16 APIs & Breaking Changes:**
   - **Do not rely on training weights.** Read guides in `node_modules/next/dist/docs/` before implementing routes, boundaries, or server components.
   - Uses `proxy.ts` (not `middleware.ts`). All request APIs (`cookies()`, `headers()`, `params`, `searchParams`) are asynchronous Promises.

3. **UI & Components (`shadcn/ui`):**
   - **Read `.agents/skills/shadcn/` before writing UI.**
   - Base UI (`base-vega` style): Use `render` props (never Radix `asChild`), `data-icon` button attributes, `<FieldGroup>` forms, `<Skeleton>` loaders, and semantic color tokens.
   - Reuse `@/components/ui/` primitives; merge classes with `cn()`.

4. **Authentication & SSR (`supabase`):**
   - **Read `.agents/skills/supabase/SKILL.md` before touching auth.**
   - **Centralized route protection:** Handled at edge in `proxy.ts`. Do not re-implement redundant in-page redirect guards or sign-in cards.
   - Verify sessions with `supabase.auth.getClaims()` for route protection, and `supabase.auth.getUser()` when an up-to-date Auth record is required.

5. **Verification & Code Quality:**
   - Run `npm run lint` and `npx tsc --noEmit` before finishing. `npm run format` is optional.
