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

== Sequence Diagram <sequence>

*Use case:* UC-03 Submit Registration (team captain locks roster and pays entry fee)

#figure(
  image("../img/sequence-diagram.png", width: 100%),
  caption: [Sequence diagram: UC-03 Submit Registration (team captain locks roster and pays entry fee)],
)

=== Walkthrough
+ *Lock roster:* Team Captain initiates `lock roster` from Web Client, which sends `POST lock roster` through API Gateway (Traefik) to Tournament Service. Tournament Service calls `GetAccount (gRPC)` on Account Service to confirm the captain is active, validates team size and deadline, and sets `status = PAYMENT_PENDING`.
+ *Create invoice:* Tournament Service requests `create invoice (entry fee)` from Payment Service, receives `invoice UNPAID`, and responds with `payment required` so Web Client shows the checkout view to the Team Captain.
+ *Pay (success flow):* Team Captain submits `pay` (`POST pay` via API Gateway to `pay invoice` on Payment Service). Payment Service sets `invoice = PAID`, publishes `payment.succeeded` to Message Broker (RabbitMQ) via fire-and-forget (`->>`), and replies `paid` to Web Client. Tournament Service consumes `payment.succeeded`, sets `team = LOCKED`, submits registration to the organizer, and publishes `registration.submitted` to RabbitMQ. Notification Service consumes `payment.succeeded` and `registration.submitted` to send the confirmation email.
+ *Error flow (`payment fails`):* Payment Service returns `payment failed` to Web Client, which informs the Team Captain (`failed, retry allowed`). Registration is not submitted and the roster remains unlocked (FR20).
