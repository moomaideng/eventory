# Eventory: System Specification (SPECS.md) (OUTDATED)

This document specifies the core end-to-end user journeys, role interactions, and system state transitions across the platform.

---

## 1. System Roles & Identity Architecture

* **Primary Account (Human Identity):** A single central account identified by email, credentials, and a public handle. Any registered user can participate in tournaments as a competitor by default.
* **Contextual Profiles:** From their central account, a user can establish and toggle into:
* **Organizer Profile:** Represents a hosting entity (club, studio, independent organizer) to manage events, invite staff, and receive net event payouts.
* **Sponsor Profile:** Represents a brand or company entity to manage marketing budgets, brand assets (logos, URLs), and track sponsored tournaments.


* **Tournament Staff:** Standard users invited to an individual tournament instance to act as Admins or Referees with operational permissions (score reporting, check-in, announcements) without requiring full organizer privileges.

---

## 2. Core User Flows & Journeys

### Flow A: Organizer Tournament Lifecycle & Crowdfunding Staging

<!--Probably can still sponsor after the crowdfunding stage-->

```
[1. Draft Tournament] ──► [2. Crowdfunding Stage (Optional)] ──► [3. Registration Stage] ──► [4. Match Execution] ──► [5. Completion & Payout]
 - Rules & Formats         - Target Goal & Deadline               - Custom Form Active       - Seeding / Reseeding     - Placements Archived
 - Custom Form Setup       - Tiers Open for Pledges               - Solo / Lobby Entries     - Live Scoring / Referees - Net Funds Released
 - Staff Invitations       - Operations in Draft                  - Organizer Review Queue   - Auto Winner Advance     - Prize Distribution

```

1. **Drafting & Configuration:**
* The organizer creates the event, defines format presets (Single Elimination or FFA) or custom DAG stage rules, sets entry fees, and customizes registration form fields (text, dropdowns, file uploads, pattern rules).
* Operational staff (referees, check-in leads) are invited via handle/email.


2. **Crowdfunding Stage (Optional Gate):**
* If funding is required before venue/logistics commitments, the tournament publishes in a `CROWDFUNDING` status with a funding target, deadline, and tier packages.
* Competitor registrations remain locked while sponsors pledge funds.
* If the target is met (or manually overridden by the organizer), the campaign locks in funds, sponsor assets are placed on the event banner, and the tournament transitions to `REGISTRATION_OPEN`. If expired without meeting the target, all pledges are automatically refunded/released.


3. **Registration & Review Management:**
* The organizer reviews incoming solo applications and complete team packages in a unified dashboard, approving or rejecting entries individually or in bulk.


4. **Match Operations & Scoring:**
* Once registration closes, brackets/lobbies generate automatically.
* The organizer/staff can manually reseed or swap matchups before assigning court/table locations and match start times.
* Assigned staff or referees submit official scores/rankings, which automatically advance winners to subsequent rounds.


5. **Finalization & Settlement:**
* Once the final match finishes, the organizer marks the event `COMPLETED`. Standings are locked to a permanent podium view, and the financial dashboard unlocks payout distribution.



---

### Flow B: Competitor & Team Registration (Lobby Pattern)

```
                       [Browse & Discover Event]
                                   │
                ┌──────────────────┴──────────────────┐
                ▼                                     ▼
        [Solo Tournament]                      [Team Tournament]
                │                                     │
    Fill Out Custom Form                      Captain Creates Lobby
                │                                     │
    Pay Entry Fee (Stripe/Mock)               Share 6-Char Code / Link
                │                                     │
                │                             Teammates Join & Submit Forms
                │                                     │
                │                             Captain Locks Roster (Min Size Met)
                │                                     │
                │                             Captain Pays Full Team Entry Fee
                │                                     │
                └──────────────────┬──────────────────┘
                                   │
                                   ▼
                  [Status: Pending Organizer Review]
                                   │
                ┌──────────────────┴──────────────────┐
                ▼                                     ▼
          [Approved]                              [Rejected]
      Confirmed in Bracket                Automatic Full Refund Processed

```

1. **Discovery:** Competitors filter events by game/sport, entry fee, format, and physical location.
2. **Solo Entry:** The competitor completes the custom form (with auto-filled profile fields and explicit privacy disclosures) and pays the registration fee directly.
3. **Team Entry (Unified Lobby Flow):**
* **Lobby Creation:** The Captain initiates a team lobby, choosing a team name and generating a revocable invite link/code.
* **Roster Assembly:** Teammates join the lobby via the link. Each player must complete their individual custom form questions before their spot is secured in the lobby. Captains can kick members, regenerate invite links, or disband the lobby if needed.
* **Lock & Single-Payer Checkout:** Once the minimum roster size is met, the Captain locks the roster and pays the total team fee in a single transaction.


4. **Review & Refund Guarantees:** The entire team bundle enters the Organizer Review Queue. If rejected or if the tournament cancels, the platform automatically triggers a full refund to the Captain's payment method.

---

### Flow C: Sponsor Discovery & Self-Serve Checkout

<!--There's a sponsor profile--no need to upload brand logo and URL unless specified further-->

```
[Browse Sponsor Directory / Public Page] ──► [Select Tier Package] ──► [Upload Brand Logo & URL] ──► [Complete Checkout]
                                                                                                            │
                                                                                                            ▼
                                                                                           [Logo Featured on Public Event Page]

```

1. **Discovery & Contact:** Sponsors browse tournaments seeking funding; organizers can also browse the public Sponsor Directory to initiate external outreach.
2. **Self-Serve Pledge:** The sponsor navigates to the tournament page, selects an active sponsorship tier (e.g., Gold, Silver), enters their redirect URL, and uploads their brand logo.
3. **Escrow & Refund Protection:** The sponsor checks out via the payment gateway. If the event is in crowdfunding and fails to reach its goal, the system automatically voids or refunds the transaction.

---

### Flow D: Tournament Progression & Match Day

<!--May want organizers to have the ability to change the bracket graph after registration closes, and have them confirm the brackets later before the final brackets are published.-->

```
[Registration Closes] ──► [Bracket Generated] ──► [Seeding / Adjustments] ──► [Schedule & Venue Assigned]
                                                                                         │
                                                                                         ▼
[Archive & Podium Standings] ◄── [Final Match Scored] ◄── [Auto-Advancement] ◄── [Referees Input Scores]

```

1. **Bracket Initialization:** When registration closes, the system populates the match nodes (Single Elimination tree or FFA scorecard) using confirmed participants.
2. **Operational Adjustments:** Organizers/staff adjust seedings or swap slots to resolve scheduling conflicts, then publish match start times and physical stations (e.g., "Court 3", "Station B").
3. **Live Match Reporting:** Referees access the restricted scorekeeper view to record scores or multi-team rankings. The progression engine calculates winners and moves them to the next round in real time.
4. **Permanent Archive:** Upon tournament completion, the bracket switches to an immutable, read-only state displaying the final 1st, 2nd, and 3rd place standings.

---

### Flow E: Financial Ledger, Settlement & Payouts

<!--What is this? not reviewed yet-->

```
[Incoming Payments (Fees + Pledges)] ──► [Platform Settlement Ledger] ──► [Event Status: COMPLETED]
                                                                                     │
                                           ┌─────────────────────────────────────────┴─────────────────────────────────────────┐
                                           ▼                                                                                    ▼
                            [Organizer Settlement Summary]                                                       [Prize Pool Allocation]
                             - Total Entry Fees Collected                                                         - 1st / 2nd Place Rewards
                             - Gross Sponsorship Funding                                                          - Direct Transfer to Winner Accounts
                             - Less Platform Commission
                             - Net Transfer to Bank Account

```

1. **Inbound Processing:** All entry fees and sponsorship tier payments are processed via external gateways (Stripe/sandbox) and recorded as immutable ledger entries.
2. **Automated Safety Rails:** Webhooks automatically process refunds back to original payment methods if applications are rejected or campaigns fail.
3. **Post-Event Settlement:** Once a tournament reaches `COMPLETED`, the organizer views a financial breakdown (Gross Revenue, Platform Deductions, Net Balance) and initiates outbound payouts to their registered bank account, alongside automated prize distribution to verified winners.

---

## 3. High-Level Tournament State Machine

$$\text{Draft} \longrightarrow \left[ \text{Crowdfunding} \right] \longrightarrow \text{Registration Open} \longrightarrow \text{Registration Closed} \longrightarrow \text{Live / Ongoing} \longrightarrow \text{Completed}$$

* **Draft:** Setup phase; rules, formats, forms, and staff are configured.
* **Crowdfunding (Optional):** Sponsorship packages active; competitor sign-ups locked until financial goal is achieved.
* **Registration Open:** Solo registration and Team Lobbies active; forms submitted and reviewed.
* **Registration Closed:** Rosters locked; automated bracket/lobby generated; seeding and scheduling underway.
* **Live / Ongoing:** Matches played; scores submitted; brackets progress dynamically.
* **Completed:** Read-only archive; placements finalized; financial settlement and payouts unlocked.
* **Cancelled (Any Stage):** Triggers automated batch refunds to all participating sponsors and competitors.
