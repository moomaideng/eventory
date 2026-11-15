// Owner: Film
// Reviewed against docs/arch/uml-api-diagram.puml and implementation at
// main commit 4c8a0b6072dfdecbfa72fe33130692119e193a13.
// Supporting evidence: account_handler.go and account_usecase.go;
// internal/middlewares/auth.go; tournament/internal/ports/payment.go;
// tournament/internal/usecases/announcement_usecase.go;
// frontend/package.json; docs/arch/adrs/01, 12; docs/arch/risk-matrix.md.
#import "../appendix/preambles.typ": *

== Design Quality <quality>

The UML API diagram in @uml-api groups resources into Account, Tournament,
Payment, and Notification services. Its Path arrows describe URL structure,
while its use arrows connect resources to data representations. These arrows
do not show runtime calls between services. The discussion below uses the
diagram together with the implementation and architecture decisions to assess
the design. Improvements are proposed work, not completed features.

=== Reusability

*Strength.* Several resources use the same representation. For example,
`TeamLobbyJoin`, `TeamLobbyLock`, and `TeamInviteRegenerate` all link to
`TeamLobby`. This gives clients a consistent model for different steps of team
registration. The separate Payment service also provides a common payment
capability for registration and sponsorship. In the implementation, Tournament
accesses payment operations through a `PaymentService` interface, allowing
business logic to use a test substitute without requiring a live payment service.

*Weakness.* Shared models can couple unrelated screens: changing a field may
affect several clients. The diagram also simplifies different responses into
one representation. For example, both `MySponsorProfile` and `SponsorDirectory`
link to `SponsorProfile`, although the directory response in code contains only
the account ID, sponsor name, tags, and contact email.

*Improvement.* Keep common business rules and payment interfaces reusable, but
document separate response models where audiences need different fields. Show
the smaller directory response explicitly instead of encouraging every endpoint
to return the full profile.

=== Maintainability

*Strength.* The service packages make responsibilities visible. Profile changes
belong to Account, while `PaymentReceipt` and `PaymentReceiptPDF` belong to
Payment. Within Tournament, resources such as `Announcements`, `RegistrationForm`,
and `Bracket` identify distinct features. The implementation separates handlers,
use cases, and repositories, and generates frontend API types from OpenAPI.
These boundaries help developers locate a change and check its effect on clients.

*Weakness.* Tournament still contains many responsibilities, including teams,
registration, brackets, staffing, and sponsorship. A change to registration or
payment state can affect several of these areas. The UML uses domain
representations rather than listing every request and response, so it cannot
serve alone as an exact client contract. It must be kept aligned with the API
documentation as the implementation changes.

*Improvement.* Keep feature modules and their tests separate within Tournament
before considering more services. Review the UML and generated API types when
endpoints change, and add contract checks for the responses consumed by other
parts of the system. This reduces accidental incompatibility without adding
unnecessary deployment complexity.

=== Scalability

*Strength.* The separate service packages provide boundaries for scaling busy
parts of the system independently. For example, browsing `Tournaments` and
reading `Notifications` need not use the same service process. The implementation
also bounds sponsor-directory pages to at most 50 results, limiting the size of
each response. These choices support growth, but the UML alone does not prove
that the system can handle a particular load.

*Weakness.* Production currently uses Docker Compose on one host, so the
services still share machine resources and a host failure affects the whole
system. ADR-12 also notes that adding HTTP replicas does not automatically
balance the persistent gRPC connection to Account. More service instances would
therefore not remove every bottleneck.

*Improvement.* Load-test high-traffic reads and registration flows first, then
use the results to tune database queries and page sizes. Add replicas and
connection-aware load balancing where measurements justify them. Move to
multiple hosts when availability or capacity requires it; independent boxes in
the diagram should not be treated as automatic scaling or failover.

=== Security

*Strength.* The diagram identifies bearer-token authentication for protected
operations and separates personal resources such as `MyAccount` and
`MySponsorProfile` from public-facing resources. The backend verifies JWT
signatures and applies additional access rules: the sponsor directory requires
an active account with an organizer profile, and posting an announcement requires
tournament owner or admin permission. Creating an organizer or sponsor profile
requires explicit privacy consent before saving, and the consent time is stored.

*Weakness.* The diagram does not specify permissions for each operation or
fully distinguish private and public response fields. Authentication alone does
not establish that a user may access a particular tournament or another person's
data. The diagram's shared profile representations can also make the intended
data exposure unclear, even where the code already uses narrower responses.

*Improvement.* Add an access matrix for public, authenticated, owner, and staff
operations, and test unauthenticated and wrong-owner requests. Keep public and
private response contracts explicit, with only the fields each audience needs.
Explain profile visibility when requesting consent, and keep development login
tokens restricted to local/development environments as the middleware currently
does.

=== Summary
#table(
  columns: (2.9cm, 1fr, 1fr),
  table.header[*Quality*][*Strength*][*Weakness*],
  [Reusability], [Shared models and payment interfaces.],
  [Shared models can couple clients and hide response differences.],
  [Maintainability], [Clear service ownership and generated API types.],
  [Broad Tournament scope and manually maintained UML.],
  [Scalability], [Separate services and bounded directory pages.],
  [One host and uneven scaling of internal connections.],
  [Security], [JWT checks, access rules, and profile consent.],
  [Operation permissions and data visibility are not explicit in the UML.],
)
