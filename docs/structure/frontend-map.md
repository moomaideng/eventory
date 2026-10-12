# Eventory Frontend: User Journeys & Page Map (Sprint 1) (OUTDATED)

> **Document Purpose:** This document defines the initial user journeys, page architecture, and routing blueprint for the **Eventory** frontend. Use this specification to align development tasks across team members.

---

## 1. System Architecture & Core Concept

Eventory is a tournament, competition, and crowdfunding platform built with a **Single Primary Identity** and **Contextual Persona Switching**.

```mermaid
flowchart TD
    User["Single Primary Account (Google OAuth)"]
    RoleContext["Active Role Context (Navbar Switcher)"]
    Competitor["
    🎮 Competitor Persona
    - Browse & Filter Tournaments
    - Join Team Lobbies
    - Compete in Brackets
    "]
    Organizer["
    🏆 Organizer Persona (Max 1 Profile)
    - Host & Configure Tournaments
    - Crowdfunding & Rules Setup
    - Manage Matchday Staff & Referees
    "]
    Sponsor["
    💼 Sponsor Persona (Max 1 Profile)
    - Fund Prize Pools & Tiers
    - Brand Asset Management
    - Track Pledged Tournaments
    "]
    
    User --> RoleContext
    RoleContext --Default Persona--> Competitor
    RoleContext --Switch Persona--> Organizer
    RoleContext --Switch Persona--> Sponsor
```

---

## 2. Core User Journeys (Sprint 1)

### Journey 1: Authentication & Role Switching (EPIC 1: US1-1, US1-2, US1-3)

```mermaid
sequenceDiagram
    autonumber
    actor User as User / Developer
    participant App as Eventory Frontend
    participant Auth as Supabase Auth / Dev Store
    participant Context as Role Context

    alt Production Flow
        User->>App: Clicks "Sign In"
        App->>Auth: Google OAuth Sign-in
        Auth-->>App: Return User Session
    else Zero-Friction Dev Flow
        User->>App: Clicks "Dev Quick Login"
        App->>Context: Injects Mock User State
    end

    App->>Context: Initialize Default Role ('competitor')
    Context-->>App: Navbar updates to Role Switcher Dropdown
    User->>App: Clicks Dropdown -> Selects 'Organizer' or 'Sponsor'
    App->>Context: setRole('organizer')
    Context-->>App: Navigation links instantly adapt to Organizer Workspace
```

---

### Journey 2: Tournament Discovery & Overview (EPIC 2: US2-1, US2-2)

```mermaid
flowchart TD
    Explore["
    Explore Page (/tournaments)
    - Status Pills (Crowdfunding, Open, Live)
    - Search by Game & Format
    "]
    Detail["
    Detail Page (/tournaments/[id])
    - Header & Prize Pool / Goal
    - Multi-Tab Overview, Rules, Rosters
    "]
    Join["Join as Solo / Create Team Lobby"]
    Checkout["Sponsor Tier Pledge Checkout"]
    
    Explore --Select Tournament--> Detail
    Detail --If Solo / Team Open--> Join --> Fill[Fill Registration Form]
    Detail --If Crowdfunding Active--> Checkout
```

---

### Journey 3: Organizer Tournament Hosting & Lifecycle (EPIC 3: US3-1, US3-2, US3-3, US3-7)

```mermaid
flowchart TD
    Hub["
    1. Organizer Hub (/organizer)
    Overview of Hosted Events
    "]
    Wizard["
    2. Create Tournament Wizard (/organizer/tournaments/new)
    - Step 1: Basic Info & Banner
    - Step 2: Format & Roster Restrictions
    - Step 3: Crowdfunding Target (Optional)
    - Step 4: Custom Registration Questions
    "]
    ControlCenter["3. Tournament Control Center (/organizer/tournaments/[id])"]
    StatusController["
    4. Status Controller (US3-3)
    Override: Draft -> Crowdfunding -> Open -> Live -> Completed
    "]
    StaffManagement["
    5. Staff Management (US3-7)
    Invite Referees & Admins via Handle/Email
    "]

    Hub --> Wizard
    Wizard --> ControlCenter
    ControlCenter --> StatusController
    ControlCenter --> StaffManagement
```

---

### Journey 4: Team Formation & Lobby Room (EPIC 4: US4-1, US4-2)

```mermaid
flowchart TD
    CreateLobby["
    Captain Creates Lobby
    for Tournament
    "]
    LobbyRoom["
    Lobby Room (/lobbies/[code])
    Generates 6-Character Code & Share Link
    "]
    OpenInvite["Teammates Open Invite Link"]
    JoinModal["Join Modal: Answer Individual Registration Questions"]
    SlotConfirmed["Teammate Slot Confirmed in Roster"]
    CheckRequirements{"Roster Requirements Met?"}
    SubmitRegistration["Captain Locks Roster & Submits Registration"]

    CreateLobby --> LobbyRoom
    LobbyRoom --> OpenInvite
    OpenInvite --> JoinModal
    JoinModal --> SlotConfirmed
    SlotConfirmed --> CheckRequirements
    CheckRequirements --Yes--> SubmitRegistration
    CheckRequirements --No--> LobbyRoom
```

---

## 3. Sprint 1 Page Map & Route Specifications

Below is the complete table of routes required for Sprint 1. Each route has a single clear purpose, mapped to specific user stories:

| Route Path | Page Name | Persona / Access | Mapped User Story | Core Components & UI Features |
| :--- | :--- | :--- | :--- | :--- |
| `/` | **Landing / Home** | Public / All | - | Clean Hero title, Value proposition, Primary CTAs (`Explore Tournaments`, `Host Tournament`). |
| `/login` | **Authentication** | Public (Unauthed) | **US1-1** | Single-click "Sign in with Google" card + "Dev Quick Login" fallback button. |
| `/onboarding` | **Handle Setup** | Authenticated | **US1-1** | First-login profile setup: Display Name & Avatar confirmation. |
| `/tournaments` | **Tournament Explore** | All (Competitor focus) | **US2-1** | Search bar, Status filter pills (`Crowdfunding`, `Registration Open`, `Live`, `Completed`), Game tags, Tournament Cards with funding/roster progress bars. |
| `/tournaments/[id]` | **Tournament Detail** | All | **US2-2** | Hero banner, Organizer identity, Funding progress bar (if crowdfunding), Tabbed content (Overview, Rules & Format, Participants/Teams, Sponsors), CTA button. |
| `/lobbies/[code]` | **Team Lobby Room** | Competitor (Teams) | **US4-1, US4-2** | 6-character room code with copy link, Roster slot cards, Member question completion status indicator, Captain controls (Kick, Regenerate Code, Lock Roster). |
| `/organizer` | **Organizer Hub** | Organizer | **US3-2** | Dashboard summarizing hosted tournaments, live participant counts, active funding campaigns, "+ Create Tournament" CTA. |
| `/organizer/tournaments/new` | **Create Tournament** | Organizer | **US3-1** | 4-step wizard: 1) General Info, 2) Tournament Format/Rules/Fee, 3) Crowdfunding Target/Tiers, 4) Registration Form Question Builder. |
| `/organizer/tournaments/[id]` | **Tournament Manager** | Organizer | **US3-2, US3-3, US3-7** | Key metrics, Lifecycle Status Transition selector (Status Override), Staff and Referee invitation table with role assignment. |
| `/sponsor` | **Sponsor Dashboard** | Sponsor | **US1-2, US5-5** | Pledged tournaments overview, brand asset preview (Logo, Website redirect link), sponsorship tier tracker. |
| `/settings` | **Profile & Settings** | Authenticated | **US1-2, US1-3** | Tabbed settings: 1) Personal Profile, 2) Organizer Profile (Club name, Bio, Logo), 3) Sponsor Profile (Company name, Website, Logo). |

---

## 4. Team Work Breakdown (How to Divide the Work)

To prevent merge conflicts, team members can be assigned by feature domain:

```mermaid
treeView-beta
frontend/
├── src/
│   ├── (auth)/      ## Person A: Auth & Identity (Login, Onboarding, Settings)
│   ├── tournaments/ ## Person B: Discovery & Details (Explore, Detail Tabs)
│   ├── organizer/   ## Person C: Organizer Hub & Tournament Creation Wizard
│   └── lobbies/     ## Person D: Team Lobbies & Questionnaire Join Flow
```

---

## 5. Peer Review Checklist for Teammates

- [ ] Does the **3-Persona Context Switcher** (Competitor / Organizer / Sponsor) fit our platform vision?
- [ ] Are the **4 core flows** (Auth, Discovery, Organizer Lifecycle, Team Lobby) clear and sufficient for Sprint 1?
- [ ] Is the page routing structure logical and easy to navigate?
- [ ] Are there any missing pages or duplicate responsibilities in the list above?
