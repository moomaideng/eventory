// ============================================================
// Owners: Ham, Tawan, Not
// Task:   UML API diagram + short explanation
//
// 1. Draw a UML class diagram of the API layer, one box per
//    service. Each box lists:
//    endpoints as operations, request/response DTOs, and links to
//    the domain classes it uses.
//    Sources: Swagger UI (endpoints), backend/services/*/
//    Reuse the style of docs/arch/c4-level4-code-teamlobby.puml (PlantUML).
// 2. Export SVG/PNG into docs/report/img/, then swap the
//    placeholder for the commented #figure line.
// 3. Replace every todo[...] with the real text.
// ============================================================
#import "../appendix/preambles.typ": *

== UML API Diagram <uml-api>

// #placeholder[UML API diagram]
#figure(image("../img/uml-api-diagram.png", width: 90%), caption: [UML API diagram])

=== Overview
- *Base path:* All REST endpoints sit under `EventoryAPI` with `basePath = "/api/v1"`.
- *Boxes:* Teal `<<Resource>>` boxes list HTTP methods per endpoint; pink `<<Representation>>` boxes show the JSON payloads.
- *Health & Auth:* Each service exposes `GET /health` outside `/api/v1`, while Notification Service also exposes `GET /ready`. Protected operations require `Authorization: Bearer <JWT>`.

=== Services and Endpoints
#table(
  columns: (auto, 1fr, 1.4fr),
  table.header[*Service*][*Responsibility*][*Key endpoints*],
  [Account],
  [User accounts, organizer profiles, and sponsor directory],
  [`/accounts` \ `/accounts/me` \ `/accounts/me/organizer-profile` \ `/accounts/me/sponsor-profile` \ `/accounts/sponsors`],
  [Tournament],
  [Tournaments, registration forms, team lobbies, brackets, and sponsorships],
  [`/tournaments` \ `/tournaments/{tournamentId}/teams` \ `/tournaments/{tournamentId}/bracket` \ `/tournaments/{tournamentId}/sponsorship-tiers` \ `/lobbies/{inviteCode}`],
  [Payment],
  [Checkout invoices, mock payment results, and receipts],
  [`/payments` \ `/payments/{id}` \ `/payments/{id}/mock-result` \ `/payments/{id}/receipt` \ `/payments/{id}/receipt.pdf`],
  [Notification],
  [User notification feed and read-status updates],
  [`/notifications` \ `/notifications/read-all` \ `/notifications/{id}/read`],
)

=== Relationships
- *URL nesting via `<<Path>>`:* Orange arrows build routes from `EventoryAPI` down to each resource.
- *Payload links via `<<use>>`:* Pink arrows connect resources to the representations they accept or return.
- *Shared models:* `TeamLobbyJoin`, `TeamLobbyLock`, `TeamInviteRegenerate`, and `TeamSubmissionRefresh` reuse the same `TeamLobby` representation.
- *Out of scope:* Cross-service gRPC calls and broker events are shown in @arch and @sequence.
