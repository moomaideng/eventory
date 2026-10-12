# Backend Endpoints (OUTDATED)

### Epic 1: Account & Profile Management

| Method | Endpoint | Description | Story Ref |
| --- | --- | --- | --- |
| `POST` | `/api/v1/accounts` | Register primary user account | US1-1 |
| `GET` | `/api/v1/accounts/me` | Fetch authenticated user's account details | US1-1 |
| `PATCH` | `/api/v1/accounts/me` | Update primary account info (handle, contact) | US1-1 |
| `GET` | `/api/v1/accounts/{id}` | Get public profile of an account | US1-1, US2-3 |
| `PUT` | `/api/v1/accounts/me/organizer-profile` | Upsert linked organizer profile details | US1-2, US1-3 |
| `GET` | `/api/v1/accounts/me/organizer-profile` | Get current user's organizer profile | US1-3 |
| `PUT` | `/api/v1/accounts/me/sponsor-profile` | Upsert linked sponsor profile details | US1-2, US1-3 |
| `GET` | `/api/v1/accounts/me/sponsor-profile` | Get current user's sponsor profile | US1-3 |

---

### Epic 2 & 3: Tournament Discovery, Hosting & Management

| Method | Endpoint | Description | Story Ref |
| --- | --- | --- | --- |
| `GET` | `/api/v1/tournaments` | Browse & filter tournaments (status, game, dates) | US2-1 |
| `POST` | `/api/v1/tournaments` | Create a new tournament | US3-1 |
| `GET` | `/api/v1/tournaments/{id}` | Get public tournament details | US2-2 |
| `PATCH` | `/api/v1/tournaments/{id}` | Update tournament configuration | US3-1 |
| `GET` | `/api/v1/tournaments/{id}/dashboard` | Fetch organizer operational dashboard summary | US3-2 |
| `PATCH` | `/api/v1/tournaments/{id}/status` | Override lifecycle status (e.g., cancel, force-start) | US3-3, US6-6 |
| `GET` | `/api/v1/tournaments/{id}/announcements` | List tournament announcements | US3-4 |
| `POST` | `/api/v1/tournaments/{id}/announcements` | Publish a new tournament announcement | US3-4 |
| `GET` | `/api/v1/tournaments/{id}/registration-form` | Fetch registration form field schema | US3-5 |
| `PUT` | `/api/v1/tournaments/{id}/registration-form` | Configure custom registration form fields | US3-5 |
| `GET` | `/api/v1/tournaments/{id}/applications` | Organizer queue of solo/team submissions | US3-6 |
| `PATCH` | `/api/v1/tournaments/{id}/applications/{appId}` | Approve or reject a submitted entry | US3-6 |
| `GET` | `/api/v1/tournaments/{id}/staff` | List assigned staff and referee accounts | US3-7 |
| `POST` | `/api/v1/tournaments/{id}/staff` | Assign role (referee/staff) to an account | US3-7 |
| `DELETE` | `/api/v1/tournaments/{id}/staff/{userId}` | Remove tournament staff assignment | US3-7 |

---

### Epic 4: Registrations & Team Lobbies

| Method | Endpoint | Description | Story Ref |
| --- | --- | --- | --- |
| `POST` | `/api/v1/tournaments/{id}/registrations/solo` | Submit registration for solo tournaments | US4-4 |
| `POST` | `/api/v1/tournaments/{id}/lobbies` | Create team lobby & generate invite link/code along with team captain form answers | US4-1 |
| `GET` | `/api/v1/lobbies/invite/{code}` | Get lobby roster via invite code | US4-2 |
| `POST` | `/api/v1/lobbies/{id}/join` | Join lobby with user form answers | US4-2 |
| `DELETE` | `/api/v1/lobbies/{id}/members/{userId}` | Kick team member from lobby (Captain only) | US4-5 |
| `POST` | `/api/v1/lobbies/{id}/invite-code` | Regenerate invite link and invalidate previous | US4-6 |
| `DELETE` | `/api/v1/lobbies/{id}` | Disband lobby | US4-7 |
| `POST` | `/api/v1/lobbies/{id}/submit` | Lock roster & submit entry for organizer review | US4-3 |
