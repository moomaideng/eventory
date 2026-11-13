// ============================================================
// Owners: Mavin (stories 1–2), Imtaez (stories 3–4). Swap if you like.
// Task:   Black-box test plan for 4 user stories of ONE EPIC
//
// Pick the EPIC with Eff/Pun (epic-a.typ uses a different one).
// Stories + acceptance criteria: docs/backlog/backlog.md,
//                                docs/backlog/acceptance-criteria.md
//
// Per story:
//   1. ec-table: split each input into equivalence classes
//      (EC = a group of inputs the system should treat the same).
//      Give each input at least 1 valid + 1 invalid class, with a concrete example.
//      EC IDs only need to be unique within the story.
//   2. test-cases: cover EVERY EC at least once.
//      TC IDs must be unique across the WHOLE document (TC51, TC52, ...).
//      `actual` + `status` ("P"/"F") are filled after running the tests;
//      leave them out until then.
// The coverage matrix at the end is built automatically.
// Where: `backend/tests/<domain>/` (godog `.feature` + steps), `backend/internal/**/*_test.go`
// ============================================================
#import "../appendix/preambles.typ": *

= EPIC #todo[number]: #todo[name]

// ---------- Story 1 (Mavin) ----------
== US #todo[x-y]: #todo[title]
_As a ..., I want ..., so that ..._

=== Input Equivalence Classes
#ec-table(
  ("EC1", "Email", [valid format, e.g. `a@b.com`], true),
  ("EC2", "Email", [missing `@`, e.g. `ab.com`], false),
)

=== Test Cases
#test-cases(
  story: "US p-q",
  (id: "TC51", inputs: [email = `a@b.com`], expected: [account created], ecs: "EC1"),
  (id: "TC52", inputs: [email = `ab.com`], expected: [error "invalid email"], ecs: "EC2"),
)

// ---------- Story 2 (Mavin) ----------
== US #todo[x-y]: #todo[title]
#todo[copy the Story 1 block]

// ---------- Story 3 (Imtaez) ----------
== US #todo[x-y]: #todo[title]
#todo[copy the Story 1 block]

// ---------- Story 4 (Imtaez) ----------
== US #todo[x-y]: #todo[title]
#todo[copy the Story 1 block]
