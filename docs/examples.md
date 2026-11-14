# Eventory Playground & Seed Examples

This guide lists concrete workflows you can explore and test right after running `task seed` — both by switching between the three built-in **Dev Quick Login** profiles (`Alex Dev`, `Bob Dev`, `Carter Dev`) and by signing in as a **brand-new user**.

---

## 1. Setup & Switching Profiles

Start the stack and seed fresh data:

```bash
task dev
task db:reset && task seed
```

> Re-running `task seed` on an already populated database is safe and idempotent: static fixtures (accounts, tournaments, forms, brackets) are updated cleanly while active payment/pledge checkouts and notification read states you changed in the UI are preserved.

To switch dev profiles in the UI:
1. Open `http://localhost:3000`.
2. Before signing in, click **Dev Quick Login** (in the top navbar or on `/login`) and choose **Alex Dev**, **Bob Dev**, or **Carter Dev**.
3. Once signed in, open the avatar menu in the top-right navbar and pick any account under **Dev Profile** to switch identities immediately.

| Dev Token | Display Name | Account Email | Handle | Organizer Profile | Sponsor Profile | Best Used For |
|---|---|---|---|---|---|---|
| `alex-dev` | **Alex Dev** | `alex.dev@eventory.gg` | `@alex_dev` | `Alex Events` | `Alex Ventures` | Full-stack all-rounder: owns feature tests, acts as `ADMIN`/`REFEREE`, competes, and sponsors |
| `bob-dev` | **Bob Dev** | `bob.dev@eventory.gg` | `@bob_dev` | `Bob Events` | — | Organizer-only workflows, owning live/crowdfunding events, and **Sponsor Onboarding** prompt |
| `carter-dev` | **Carter Dev** | `carter.dev@eventory.gg` | `@carter_dev` | — | `Carter Fundings` | Sponsor-first dashboard, receipts/payment history, and **Organizer Onboarding** prompt |

---

## 2. Playbook with `Alex Dev` (`alex-dev`)

### A. Organizer Dashboard & Role Filters (`/organizer`)
- Visit [`/organizer`](http://localhost:3000/organizer). Alex holds **12 staff memberships** (`9 OWNER`, `2 ADMIN`, `1 REFEREE`), fitting on Page 1 (`pageSize = 12`) so the **All / Owner / Admin / Referee** filter pills work cleanly.

### B. Draft Setup & Publishing Transitions
- **`[TEST · Draft] Setup & Publish Gate`** ([`/organizer/tournaments/33333333-0000-4000-8000-000000000001`](http://localhost:3000/organizer/tournaments/33333333-0000-4000-8000-000000000001))
  - Starts completely blank (`0` sponsorship tiers, no custom registration form).
  - Edit tournament metadata, build a custom form in the **Form** tab, add sponsorship tiers, and test publishing to `CROWDFUNDING` vs skipping straight to `REGISTRATION`.
- **`Chula Inter-Faculty Valorant Cup 2026`** ([`/organizer/tournaments/33333333-0000-4000-8000-000000000101`](http://localhost:3000/organizer/tournaments/33333333-0000-4000-8000-000000000101))
  - Pre-configured realistic `DRAFT` with markdown rules, 3 tiers, and a 5-field faculty registration form ready to publish.

### C. Crowdfunding Goals & Tier Locks
- **`[TEST · Crowdfunding] Goal Reached (Normal Open)`** ([`/organizer/tournaments/33333333-0000-4000-8000-000000000002`](http://localhost:3000/organizer/tournaments/33333333-0000-4000-8000-000000000002))
  - Confirmed sponsorships total **฿35,000 / ฿30,000 (116%)**. Open registration normally without forcing.
- **`[TEST · Crowdfunding] Underfunded & Tier Locks`** ([`/organizer/tournaments/33333333-0000-4000-8000-000000000003`](http://localhost:3000/organizer/tournaments/33333333-0000-4000-8000-000000000003))
  - Only **฿15,000 / ฿50,000** confirmed (`+ ฿5,000 PENDING`, plus `EXPIRED` and `CANCELLED` pledges):
    - Verify normal open fails with a goal-not-reached error, while force-open succeeds.
    - Inspect tier locking states in the **Sponsorship** tab:
      - `Gold (Sold Out)` (`1/1` confirmed, locked)
      - `Silver (Locked)` (active `PENDING` pledge, locked)
      - `Community (Editable)` (only expired/cancelled attempts)
      - `Supporter (Fresh)` (`0` pledges, safe to edit or delete)

### D. Sponsor Directory (`/organizer/sponsors`)
- Visit [`/organizer/sponsors`](http://localhost:3000/organizer/sponsors) to test search and multi-tag filtering across **10 visible sponsors** (`Alex Ventures`, `Carter Fundings`, `Maya Works Studio`, and 7 showcase brands across tags such as `hardware`, `peripherals`, `esports`, `fintech`, `cloud`, `streaming`, `mobile-gaming`, `thai-community`).
- Confirm `Stealth Sponsor` (seeded without `SponsorEmail` / privacy consent) does **not** appear in directory results.

### E. Registration Forms, Application Queue, Refunds & Audit Trail
- Open **`[TEST · Review] Queue, Forms & Refunds`** ([`/organizer/tournaments/33333333-0000-4000-8000-000000000005`](http://localhost:3000/organizer/tournaments/33333333-0000-4000-8000-000000000005)) -> **Registrations** tab:
  - **Forming & Incomplete Form Blocker:**
    - `TM050001` (`Dev Alpha (Forming)`, invite code `TEST01`) has Alex as captain + 1 teammate (both have valid form answers).
    - `TM050002` (`Missing Form Five`, invite code `TEST02`) intentionally leaves `player_02` without a MongoDB form response — try locking it to see the `registrationComplete: false` guard.
  - **Pending Checkout:** `TM050003` (`Pending Checkout Squad`) is in `PAYMENT_PENDING` awaiting entry-fee settlement.
  - **Reject + Automatic Refund:** `TM050004` (`Neon Pulse (Refund Test)`), `TM050005`, and `TM050006` are `LOCKED` with `PAID` ฿500 invoices. Rejecting `TM050004` in the organizer UI triggers a real refund in Payment Service.
  - **Review History:** Inspect `TM050007`–`TM050010` to view pre-seeded `TournamentTeamReview` audit log entries (`ACCEPTED` and `REJECTED` with reasons).

### F. Bracket Setup, Live Matches & Completed Archive
- **Bracket Setup & Start Gate (`[TEST · Bracket] Setup & Start Gate` — [`.../00000007`](http://localhost:3000/organizer/tournaments/33333333-0000-4000-8000-000000000007)):**
  - `16 ACCEPTED` teams, `0` unresolved entries, and a past `RegistrationDeadline`. Build stages, seed slots, schedule matches, and transition `REGISTRATION -> ONGOING`.
- **Live Match Reporting, Stage Skipping & Swapping (`[TEST · Bracket] Live Match & Swap` — [`.../00000010`](http://localhost:3000/organizer/tournaments/33333333-0000-4000-8000-000000000010)):**
  - Uneven 9-team 3-stage bracket: `Play-In Pod A (3 -> 2)` is completed into Stage 2 `Upper Semifinal (3 -> 2)`, `Wildcard Duel (2 -> 1)` skipped Stage 2 directly into Stage 3 `Grand Final Pod (3 -> 1)`, and `Play-In Pod B (4 -> 1)` remains open to swap teams, reschedule, record scores, and finish the tournament (`ONGOING -> COMPLETED`).
- **Completed Archive (`[TEST · Bracket] Completed Archive` — [`.../00000011`](http://localhost:3000/organizer/tournaments/33333333-0000-4000-8000-000000000011)):**
  - Completed 6-team uneven stage-skipping archive (`Qualifying Pod 4 -> 2` -> `Main Semifinal 2 -> 1`, plus `Fast-Track Duel 2 -> 1` skipping Stage 2 directly into `Championship Final 2 -> 1`). All mutations are locked once `COMPLETED`.

### G. Permission Matrix (`ADMIN` & `REFEREE` Views)
- **`[TEST · Permissions] Admin Role View`** ([`/organizer/tournaments/33333333-0000-4000-8000-000000000008`](http://localhost:3000/organizer/tournaments/33333333-0000-4000-8000-000000000008)) & **`ASEAN Dota 2 Invitational 2026`** ([`.../00000105`](http://localhost:3000/organizer/tournaments/33333333-0000-4000-8000-000000000105)):
  - Alex is `ADMIN` (owned by `Bob Dev`): can review applications, manage brackets, and post announcements, while owner-only actions (`Form`, `Sponsorship`, `Staff Invite/Remove`, `Edit Details`) are hidden/forbidden.
- **`[TEST · Permissions] Referee Role View`** ([`/organizer/tournaments/33333333-0000-4000-8000-000000000009`](http://localhost:3000/organizer/tournaments/33333333-0000-4000-8000-000000000009)):
  - Alex is `REFEREE` (owned by `Bob Dev`): verify read-only **Home**, **Bracket** scoring/swapping/scheduling, and **Leave Tournament** in **Staff**.

### H. Playing as Sponsor & Competitor with Alex
- **Sponsor Checkout Playground:** Switch to Sponsor mode or open [`/sponsor/tournaments/33333333-0000-4000-8000-000000000004`](http://localhost:3000/sponsor/tournaments/33333333-0000-4000-8000-000000000004) (`[TEST · Sponsor] Checkout Playground`, owned by Maya). Alex has no active pledge there and can immediately run Sponsor Checkout + Mock Payment.
- **Paid Solo Form + Checkout:** Visit [`/tournaments/33333333-0000-4000-8000-000000000006`](http://localhost:3000/tournaments/33333333-0000-4000-8000-000000000006) (`[TEST · Solo] Paid Entry & Form Playground`, ฿250) to fill out all 5 custom form field types and complete entry-fee checkout.
- **Free Solo Join:** Visit `TH Street Fighter 6 Community Clash` ([`/tournaments/33333333-0000-4000-8000-000000000104`](http://localhost:3000/tournaments/33333333-0000-4000-8000-000000000104)) and join immediately with no payment step.
- **Payments & Bell:** Check [`/payments`](http://localhost:3000/payments) to view/download the PDF receipt for Alex's ฿40,000 Platinum sponsorship on `ASEAN Dota 2 Invitational 2026`, and click the notification bell (`2` unread, `2` read across all 4 notification types).

---

## 3. Playbook with `Bob Dev` (`bob-dev`) — Organizer Only

- **Organizer Dashboard ([`/organizer`](http://localhost:3000/organizer)):**
  - Bob owns **5 tournaments** (`A8`, `A9`, `Bangkok Cyber Arena 2026`, `TH Street Fighter 6 Community Clash`, `ASEAN Dota 2 Invitational 2026`) and serves as `ADMIN` on `SEA Valorant Challengers Open`.
  - Test Owner-only actions (`Staff` invite/remove, `Sponsorship` tier edits, `Details` editing, status transitions) on **`Bangkok Cyber Arena 2026`** ([`/organizer/tournaments/33333333-0000-4000-8000-000000000102`](http://localhost:3000/organizer/tournaments/33333333-0000-4000-8000-000000000102)).
- **Sponsor Onboarding Trigger:**
  - Switch into **Sponsor** mode (`/sponsor`) while logged in as Bob — because `bob-dev` has **no `SponsorProfile`**, you can immediately test the Sponsor Onboarding prompt (`ErrSponsorRequired`) and create a new sponsor profile.
- **Notifications:** Has unread notifications for `HyperVolt Energy`'s ฿35,000 sponsorship on `Bangkok Cyber Arena 2026` and Alex's broadcast announcement on `ASEAN Dota 2 Invitational 2026`.

---

## 4. Playbook with `Carter Dev` (`carter-dev`) — Sponsor Only

- **Sponsor Portal ([`/sponsor`](http://localhost:3000/sponsor), [`/sponsor/tournaments`](http://localhost:3000/sponsor/tournaments), [`/sponsor/payments`](http://localhost:3000/sponsor/payments)):**
  - Carter already backs **7 confirmed sponsorships** (`A2`, `A3`, `A4`, `A6`, `B3`, `B5`, `B6`).
  - Browse Carter's active sponsored events and download PDF receipts from `/sponsor/payments` or [`/payments`](http://localhost:3000/payments).
  - Run another Sponsor Checkout on an open event Carter has **not** sponsored yet, such as **`Bangkok Cyber Arena 2026`** ([`/sponsor/tournaments/33333333-0000-4000-8000-000000000102`](http://localhost:3000/sponsor/tournaments/33333333-0000-4000-8000-000000000102)) or **`[TEST · Bracket] Live Match & Swap`** ([`/sponsor/tournaments/33333333-0000-4000-8000-000000000010`](http://localhost:3000/sponsor/tournaments/33333333-0000-4000-8000-000000000010)).
- **Organizer Onboarding Trigger:**
  - Switch into **Organizer** mode (`/organizer`) while logged in as Carter — because `carter-dev` has **no `OrganizerProfile`**, you can test the Organizer Onboarding flow from scratch.

---

## 5. Exploring on a Brand-New Profile (Fresh User / Google Login)

If you sign in with a fresh account (never seeded with teams or extended profiles), you can exercise the full Competitor + Onboarding experience:

1. **Public Catalog Pagination & Discovery ([`/tournaments`](http://localhost:3000/tournaments)):**
   - **15 public tournaments** are visible (the `2 DRAFT` events stay hidden), triggering **2 pages** (`pageSize = 12`).
   - Filter across `CROWDFUNDING`, `REGISTRATION`, `ONGOING`, `COMPLETED`, games (`Valorant`, `Counter-Strike 2`, `Dota 2`, `Street Fighter 6`, `Mobile Legends`), and free vs. paid entry fees (฿0 to ฿2,000).
2. **Join an Existing Team Lobby via Invite Code ([`/lobbies`](http://localhost:3000/lobbies)):**
   - All three open lobbies have room for a new member:
     - `VALO26` — **Tonkla Open Lobby** (`4/5` players) on `SEA Valorant Challengers Open` (Paid ฿800, 5v5)
     - `TEST01` — **Dev Alpha (Forming)** (`2/3` players) on `[TEST · Review] Queue, Forms & Refunds` (Paid ฿500)
     - `ADM006` — **Open Admin Lobby** (`2/4` players) on `[TEST · Permissions] Admin Role View` (Free)
3. **Fill Out Custom MongoDB Forms & Register Solo or Create a Team:**
   - **Free Solo:** Register on **`TH Street Fighter 6 Community Clash`** (`...000104`) and fill out its 4-field form (`CFN Player Handle`, `Primary Character`, `Controller Type`, `Self-Rated Seed Tier`).
   - **Paid Solo + Mock Checkout:** Register on **`[TEST · Solo] Paid Entry & Form Playground`** (`...000006`), complete all 5 form types (`SHORT_TEXT`, `LONG_TEXT`, `NUMBER`, `SELECT_ONE`, `SELECT_MULTIPLE`), and test `SUCCEEDED`, `FAILED`, or `CANCELLED` in Mock Payment Checkout.
   - **Create a Team:** Start a new squad on `SEA Valorant Challengers Open` (`...000103`) or `[TEST · Review] Queue, Forms & Refunds` (`...000005`), invite friends by code, answer the intake form, and lock when complete.
4. **Watch Live Brackets, Final Rankings & Announcements:**
   - **`ASEAN Dota 2 Invitational 2026`** ([`/tournaments/33333333-0000-4000-8000-000000000105/bracket`](http://localhost:3000/tournaments/33333333-0000-4000-8000-000000000105/bracket)) — 20-team multi-pod bracket (`5 -> 2` ×4 Groups feeding `4 -> 2` ×2 Playoff Heats and a `4 -> 1` Grand Final Pod; Groups A & B complete, Groups C & D upcoming), plus 3 announcements.
   - **`Thailand Mobile Legends Championship S4`** ([`/tournaments/33333333-0000-4000-8000-000000000106/bracket`](http://localhost:3000/tournaments/33333333-0000-4000-8000-000000000106/bracket)) — completed standard 8-team 3-stage single-elimination bracket (`Siam Titans` 3–1 champion) with full rankings and announcements.
5. **Fresh Account / Organizer / Sponsor Onboarding:**
   - Complete initial Account onboarding (`@handle` & display name), then opt into creating your own Organizer profile or Sponsor profile cleanly.

---

## 6. Seeded Tournament Quick Reference

| Code | Name | UUID | Status | Mode / Fee | Owner | Invite Code(s) |
|---|---|---|---|---|---|---|
| A1 | `[TEST · Draft] Setup & Publish Gate` | `33333333-0000-4000-8000-000000000001` | `DRAFT` | Team (2–5) · ฿0 | Alex Dev | — |
| A2 | `[TEST · Crowdfunding] Goal Reached (Normal Open)` | `33333333-0000-4000-8000-000000000002` | `CROWDFUNDING` | Team (3–5) · ฿500 | Alex Dev | — |
| A3 | `[TEST · Crowdfunding] Underfunded & Tier Locks` | `33333333-0000-4000-8000-000000000003` | `CROWDFUNDING` | Solo · ฿200 | Alex Dev | — |
| A4 | `[TEST · Sponsor] Checkout Playground` | `33333333-0000-4000-8000-000000000004` | `CROWDFUNDING` | Solo · ฿0 | Maya Works | — |
| A5 | `[TEST · Review] Queue, Forms & Refunds` | `33333333-0000-4000-8000-000000000005` | `REGISTRATION` | Team (2–3) · ฿500 | Alex Dev | `TEST01`, `TEST02` |
| A6 | `[TEST · Solo] Paid Entry & Form Playground` | `33333333-0000-4000-8000-000000000006` | `REGISTRATION` | Solo · ฿250 | Maya Works | — |
| A7 | `[TEST · Bracket] Setup & Start Gate` | `33333333-0000-4000-8000-000000000007` | `REGISTRATION` | Team (1–3) · ฿0 | Alex Dev | — |
| A8 | `[TEST · Permissions] Admin Role View` | `33333333-0000-4000-8000-000000000008` | `REGISTRATION` | Team (2–4) · ฿0 | Bob Dev | `ADM006` |
| A9 | `[TEST · Permissions] Referee Role View` | `33333333-0000-4000-8000-000000000009` | `ONGOING` | Team (2–3) · ฿0 | Bob Dev | — |
| A10 | `[TEST · Bracket] Live Match & Swap` | `33333333-0000-4000-8000-000000000010` | `ONGOING` | Team (1–3) · ฿0 | Alex Dev | — |
| A11 | `[TEST · Bracket] Completed Archive` | `33333333-0000-4000-8000-000000000011` | `COMPLETED` | Team (1–3) · ฿0 | Alex Dev | — |
| B1 | `Chula Inter-Faculty Valorant Cup 2026` | `33333333-0000-4000-8000-000000000101` | `DRAFT` | Team (5v5) · ฿1,000 | Alex Dev | — |
| B2 | `Bangkok Cyber Arena 2026` | `33333333-0000-4000-8000-000000000102` | `CROWDFUNDING` | Team (5v5) · ฿1,500 | Bob Dev | — |
| B3 | `SEA Valorant Challengers Open` | `33333333-0000-4000-8000-000000000103` | `REGISTRATION` | Team (5v5) · ฿800 | Alex Dev | `VALO26` |
| B4 | `TH Street Fighter 6 Community Clash` | `33333333-0000-4000-8000-000000000104` | `REGISTRATION` | Solo · ฿0 | Bob Dev | — |
| B5 | `ASEAN Dota 2 Invitational 2026` | `33333333-0000-4000-8000-000000000105` | `ONGOING` | Team (5v5) · ฿2,000 | Bob Dev | — |
| B6 | `Thailand Mobile Legends Championship S4` | `33333333-0000-4000-8000-000000000106` | `COMPLETED` | Team (5v5) · ฿500 | Maya Works | — |
