// ============================================================
// Owner: Film
// Task:  Sequence diagram for ONE use case
//
// 1. Pick one use case from docs/arch/usecase-diagram.png
//    (ideally one touching several services, e.g. team registration + payment).
// 2. Lifelines: User → Frontend → API service(s) → DB / other service.
//    Show the main flow and at least one error flow (`alt` block).
//    Use the same operation names as the UML API diagram (Ham/Tawan/Not).
// 3. Export into docs/report/img/, then swap the placeholder for the #figure line.
// ============================================================
#import "../appendix/preambles.typ": *

#pagebreak(weak: true)
== Sequence Diagram <sequence>

*Use case:* UC-03 Submit Registration: team captain locks roster and pays entry fee (paid tournament)

#figure(
  image("../img/sequence-diagram.png", width: 100%),
  caption: [Sequence diagram: UC-03 Submit Registration],
)

=== Walkthrough
#table(
  columns: (auto, 1fr),
  table.header[*Phase*][*Flow*],
  [Lock roster],
  [`aCaptain -> WebClient -> APIGateway -> TournamentService`: checks the captain's account via `AccountService` (gRPC), validates answers, deadline, and team size, then sets the team to `PAYMENT_PENDING` (roster frozen).],
  [Create invoice],
  [`TournamentService -> PaymentService` (gRPC): creates an `UNPAID` entry-fee invoice. The `paymentId` returns to `WebClient`, which opens the checkout page.],
  [Pay succeeds],
  [`PaymentService` marks the invoice `PAID`, replies to `WebClient`, then publishes `payment.confirmed`. `TournamentService` sets the team to `LOCKED` and publishes `registration.submitted`. `NotificationService` emails the captain for each event.],
  [Pay fails],
  [`PaymentService` cancels the invoice, replies to `WebClient`, then publishes `payment.cancelled`. `TournamentService` clears the payment attempt: registration stays unsubmitted, the team stays `PAYMENT_PENDING` (roster frozen), and a retry starts a new attempt.],
)
