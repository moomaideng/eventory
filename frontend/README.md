# Eventory Frontend

Frontend client for **Eventory**, a competition and tournament management platform supporting bracket management, crowdfunding prize pools, team lobbies, and contextual role switching.

---

## 🛠️ Tech Stack

- **Framework:** Next.js 15+ (App Router, React 19, TypeScript)
- **UI & Design System:** Tailwind CSS v4 + **shadcn/ui** (Base UI style) + Lucide Icons
- **Authentication & Storage:** Supabase Auth (`@supabase/ssr`) with Google OAuth & Supabase Storage
- **State & Context:** React Context (`RoleProvider` with contextual role switching & Dev Mock state)

---

## 🎯 Core Identity & Contextual Role Model

Eventory implements a **Single Primary Account** architecture allowing a user to toggle between 3 distinct contextual roles without maintaining separate logins:

1. 🎮 **Competitor Mode (Default):** Browse tournaments, join team lobbies, and compete.
2. 🏆 **Organizer Mode:** Host tournaments, configure crowdfunding campaigns, and manage staff/referees (1 profile per account).
3. 💼 **Sponsor Mode:** Browse crowdfunding campaigns, pledge prize pools, and manage brand assets (1 profile per account).

---

## 📁 Project & Routing Structure

We use Next.js **Route Groups** (`(public)` and `(auth)`) to isolate layouts and keep headers clean:

```text
frontend/
├── app/
│   ├── (auth)/                       # 🔐 Auth Flow (Minimal Header with Logo only)
│   │   ├── layout.tsx                # Auth layout
│   │   ├── login/page.tsx            # Clean Google Sign-In card
│   │   └── onboarding/page.tsx       # Display Name configuration
│   │
│   ├── (public)/                     # 🌐 Public & App Views (Full Navbar with Role Switcher)
│   │   ├── layout.tsx                # Public layout with Navbar
│   │   └── page.tsx                  # Minimal Landing Hero & CTA buttons
│   │
│   ├── api/
│   │   └── auth/callback/route.ts    # Supabase OAuth PKCE code exchange Route Handler
│   │
│   ├── layout.tsx                    # Root HTML layout (Fonts, globals.css, RoleProvider)
│   └── globals.css                   # Tailwind CSS v4 & theme variables
│
├── components/
│   ├── navbar.tsx                    # Header with Base UI DropdownMenu Role Switcher & Dynamic Nav Links
│   └── ui/                           # Pure shadcn/ui primitives
│       ├── avatar.tsx
│       ├── badge.tsx
│       ├── button.tsx
│       ├── card.tsx
│       ├── dropdown-menu.tsx
│       ├── input.tsx
│       └── separator.tsx
│
├── context/
│   └── role-context.tsx              # Role & Auth Context (Google OAuth + Dev Mock Login)
│
├── lib/
│   ├── client.ts                     # Supabase Browser Client helper
│   ├── server.ts                     # Supabase Server Component helper
│   ├── middleware.ts                 # Supabase Session Middleware helper
│   └── utils.ts                      # Tailwind class merge helper (`cn`)
│
└── .agents/skills/shadcn/            # 📖 shadcn/ui Best Practices & Rules
```

---

## 📖 UI Guidelines & shadcn Skill Rules (`.agents/skills/shadcn/`)

All developers working on the frontend must follow the standards and rules defined in `.agents/skills/shadcn/`:

### 1. Base UI Composition (`base-vs-radix.md` & `composition.md`)
- **Triggers use `render`:** In Base UI, use `render={<Button ... />}` (do not use Radix's `asChild`).
- **Group Component Scope:** All `DropdownMenuItem` and `DropdownMenuLabel` elements **MUST** be placed inside `<DropdownMenuGroup>`.
- **Card Structure:** Use full composition: `Card` → `CardHeader` (`CardTitle`, `CardDescription`) → `CardContent` → `CardFooter`.
- **Avatars:** Always include `<AvatarFallback>` inside `<Avatar>`.
- **Separators:** Use `<Separator />` instead of `<hr>` or custom border `<div>`s.

### 2. Spacing & Tailwind Rules (`styling.md`)
- ❌ **No `space-x-*` or `space-y-*`:** Always use `flex` with `gap-*` (e.g. `flex flex-col gap-4` or `flex gap-2`).
- **Equal Dimensions:** Use `size-*` instead of `w-* h-*` when dimensions are equal (e.g. `size-9`, `size-4`).
- **Semantic Color Tokens:** Always use semantic tokens (`bg-primary`, `text-muted-foreground`, `border-border`, `bg-card`) — never hardcode raw hex values or raw colors like `bg-blue-500`.
- **Conditional Classes:** Use `cn()` from `@/lib/utils` for conditional or merged class names.
- **No manual `z-index` on Overlays:** Dialog, Sheet, DropdownMenu, and Popovers manage their own stacking.

### 3. Adding New UI Components
To install new shadcn primitives, run:
```bash
npx shadcn@latest add <component-name>
```
*(Example: `npx shadcn@latest add dialog tabs table skeleton`)*

---

## 🚀 Getting Started

### 1. Install Dependencies
```bash
npm install
```

### 2. Configure Environment Variables (Optional for Local Dev)
Create a `.env.local` file in the `frontend` folder:
```env
NEXT_PUBLIC_SUPABASE_URL=https://your-project.supabase.co
NEXT_PUBLIC_SUPABASE_PUBLISHABLE_KEY=your-publishable-key
```

> **💡 Zero-Friction Local Dev Mode:**
> You can develop and test UI features immediately without setting up Supabase keys. Click **"Dev Quick Login"** on the Navbar to simulate logged-in states and test role switching between Competitor, Organizer, and Sponsor modes.

### 3. Run Development Server
```bash
npm run dev
```

Open [http://localhost:3000](http://localhost:3000) in your browser.
