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

*Use case:* UC-03 Submit Registration: team captain locks roster and pays entry fee

#figure(
  image("../img/sequence-diagram.png", width: 100%),
  caption: [Sequence diagram: UC-03 Submit Registration],
)

=== Walkthrough
#table(
  columns: (auto, 1fr),
  table.header[*Phase*][*Flow*],
  [Lock roster],
  [`Team Captain -> Web Client -> API Gateway -> Tournament`: checks captain status via `Account` gRPC, validates team size and deadline, and sets `status = PAYMENT_PENDING`.],
  [Create invoice],
  [`Tournament -> Payment`: creates an `UNPAID` entry-fee invoice and returns `payment required` so Web Client shows checkout.],
  [Pay success],
  [`Team Captain -> Web Client -> API Gateway -> Payment`: sets `invoice = PAID`, replies `paid` to Web Client, and publishes `payment.succeeded` via `->>`. `Tournament` sets `team = LOCKED`, submits registration, and publishes `registration.submitted` so `Notification` sends the confirmation email.],
  [Pay fails],
  [`Payment -> Web Client -> Team Captain`: returns `payment failed` and `failed, retry allowed`. Registration is not submitted and the roster stays locked.],
)
