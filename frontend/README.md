# Eventory Frontend

Frontend client for **Eventory**, a competition and tournament management platform supporting bracket management, crowdfunding prize pools, and team lobbies.

---

## 🛠️ Tech Stack

- **Framework:** Next.js 15+ (App Router, React 19, TypeScript)
- **Styling & UI:** Tailwind CSS v4 + shadcn/ui + Lucide Icons
- **Auth & Storage:** Supabase Auth (`@supabase/ssr`) & Supabase Storage
- **State & Context:** React Context (`RoleProvider` with contextual role switching)

---

## 🎯 Core Identity & Contextual Role Model

Eventory uses a **Single Primary Account** model that allows a user to toggle between 3 distinct contextual roles without maintaining separate logins:

1. 🎮 **Competitor Mode (Default):** Browse tournaments, create/join team lobbies, and compete.
2. 🏆 **Organizer Mode:** Host tournaments, manage bracket lifecycles, and assign staff/referees (1 profile per account).
3. 💼 **Sponsor Mode:** Browse crowdfunding campaigns, pledge prize pools, and manage brand assets (1 profile per account).

---

## 📁 Project Structure

```text
frontend/
├── app/                      # Next.js App Router hierarchy
│   ├── layout.tsx            # Root layout wrapping RoleProvider and Navbar
│   ├── page.tsx              # Minimal landing hero page
│   └── globals.css           # Tailwind CSS & theme tokens
│
├── context/
│   └── role-context.tsx      # Auth & Role state management (with Dev Mock Login)
│
├── components/
│   ├── navbar.tsx            # Top navigation & interactive Role Switcher dropdown
│   └── ui/                   # shadcn UI primitives (Button, Avatar, Badge)
│
└── lib/
    ├── client.ts             # Supabase Browser Client helper
    ├── server.ts             # Supabase Server Component helper
    ├── middleware.ts         # Supabase Session Middleware helper
    └── utils.ts              # Tailwind class merge helper (`cn`)
```

---

## 🚀 Getting Started

### 1. Install Dependencies

```bash
npm install
```

### 2. Environment Variables (Optional for Dev)

Create a `.env.local` file in the `frontend` root:

```env
# Supabase Auth & Storage
NEXT_PUBLIC_SUPABASE_URL=https://your-project.supabase.co
NEXT_PUBLIC_SUPABASE_PUBLISHABLE_KEY=your-anon-or-publishable-key

# Go Huma API Backend (Upcoming)
NEXT_PUBLIC_API_URL=http://localhost:8080
```

> **💡 Zero-Friction Local Dev Mode:**
> You can develop and test UI features immediately without setting up Supabase keys. Simply click **"Dev Quick Login"** on the Navbar or Landing page to simulate logged-in states and test role switching.

### 3. Run Development Server

```bash
npm run dev
```

Open [http://localhost:3000](http://localhost:3000) in your browser.

---

## 🧭 Team Development Guidelines

- **Shared UI Primitives:** Place re-usable base components in `components/ui/`.
- **Role Context:** Access the current role and user state anywhere via `const { user, activeRole, setRole } = useRole()`.
- **Clean Route Isolation:** Keep feature pages modular to minimize git merge conflicts.
