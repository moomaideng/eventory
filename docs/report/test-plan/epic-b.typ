// ============================================================
// Owners: Mavin (stories 1–2), Imtaez (stories 3–4). Swap if you like.
// Task:   Black-box test plan for 4 user stories of ONE EPIC
//
// Pick the EPIC with Eff/Pun (epic-a.typ uses a different one).
// Stories + acceptance criteria: Backlog Master (Google Sheet)
//   https://docs.google.com/spreadsheets/d/1KqyqqsrhDQ_XMk7KXhWC7pmIDYf_PSL3XvMez0JwPXg/edit?gid=1930282482#gid=1930282482
//
// Per story:
//   1. ec-table: split each input into equivalence classes
//      (EC = a group of inputs the system should treat the same).
//      Give each input at least 1 valid + 1 invalid class, with a concrete example.
//      EC IDs only need to be unique within the story.
//   2. test-cases: cover EVERY EC at least once. One card per case,
//      field list + screenshot example: see test-cases in appendix/preambles.typ.
//      TC IDs must be unique across the WHOLE document (TC51, TC52, ...).
//      `actual` + `status` ("P"/"F") are filled after running the tests;
//      leave them out until then.
// The coverage matrix at the end is built automatically.
// Where: `backend/tests/<domain>/` (godog `.feature` + steps), `backend/internal/**/*_test.go`
// ============================================================
#import "../appendix/preambles.typ": *

#let epic-ec-table(..rows) = table(
  columns: (auto, auto, 1fr, auto),
  table.header[*EC*][*Input*][*Condition*][*Valid?*],
  ..rows.pos().map(((id, input, condition, valid)) =>
    (id, input, condition, if valid [Valid] else [Invalid])
  ).flatten(),
)

#let epic-test-cases(story: "", ..cases) = {
  let cases = cases.pos()
  state("test-cases", ()).update(all => all + cases.map(c => (story: story, id: c.id, status: c.at("status", default: none))))
  for c in cases {
    let pairs = c.inputs.pairs()
    let input-rows = (table.cell(rowspan: pairs.len(), [*Inputs*]),) + pairs.flatten()
    block(below: 1.2em, breakable: true, table(
      columns: (3.4cm, 3.2cm, 1fr),
      [*Test Case ID*], table.cell(colspan: 2, c.id),
      ..input-rows,
      [*Expected Output*], table.cell(colspan: 2, [
        #c.expected

        #for path in c.screenshots { image(path, width: 100%) }
      ]),
      [*Actual Output*], table.cell(colspan: 2, c.actual),
      [*Classes Covered*], table.cell(colspan: 2, c.ecs),
      [*Status*], table.cell(colspan: 2, c.status),
    ))
  }
}

= EPIC 3: Tournament Hosting and Lifecycle

// ---------- Story 1 (Mavin) ----------
== US 3-4: Post tournament announcements
_As an organizer, I want to announce further tournament details (for example, a venue change) to related parties so that they receive updates._

=== Pre-condition
The organizer is signed in and opens the Announcements tab of a published tournament they own.

=== Input Equivalence Classes
#epic-ec-table(
  ("EC1", "Title", [1–150 characters], true),
  ("EC2", "Title", [empty], false),
  ("EC3", "Title", [over 150 characters], false),
  ("EC4", "Message", [1–2,000 characters], true),
  ("EC5", "Message", [empty], false),
  ("EC6", "Message", [over 2,000 characters], false),
)

=== Test Cases
#epic-test-cases(
  story: "US 3-4",
  (
    id: "TC51",
    inputs: (Title: [`Venue update`], Message: [`The venue is Room 301.`]),
    expected: [The announcement appears in the public list with a timestamp; success is shown and both fields clear.],
    actual: [Posted publicly with a timestamp; fields cleared.],
    screenshots: ("../img/tc51.png", "../img/tc51-public.png"),
    status: "P",
    ecs: "EC1, EC4",
  ),
  (
    id: "TC52",
    inputs: (Title: [empty], Message: [`The venue is Room 301.`]),
    expected: [A title error appears; the announcement list stays unchanged.],
    actual: [Title error shown; no new announcement.],
    screenshots: ("../img/tc52.png",),
    status: "P",
    ecs: "EC2, EC4",
  ),
  (
    id: "TC53",
    inputs: (Title: [151 `A` characters], Message: [`The venue is Room 301.`]),
    expected: [The title field accepts at most 150 characters.],
    actual: [Title stopped at 150 characters.],
    screenshots: ("../img/tc53.png",),
    status: "P",
    ecs: "EC3, EC4",
  ),
  (
    id: "TC54",
    inputs: (Title: [`Venue update`], Message: [empty]),
    expected: [A message error appears; the announcement list stays unchanged.],
    actual: [Message error shown; no new announcement.],
    screenshots: ("../img/tc54.png",),
    status: "P",
    ecs: "EC1, EC5",
  ),
  (
    id: "TC55",
    inputs: (Title: [`Venue update`], Message: [2,001 `A` characters]),
    expected: [The message field accepts at most 2,000 characters.],
    actual: [Message stopped at 2,000 characters.],
    screenshots: ("../img/tc55.png",),
    status: "P",
    ecs: "EC1, EC6",
  ),
)

// ---------- Story 2 (Mavin) ----------
== US 3-5: Configure a custom registration form
_As an organizer, I want to configure a registration form with short-text, long-text, number, single-choice, and multiple-choice questions, and mark questions as required when needed, so that I can collect tournament-specific information to review applicants._

=== Pre-condition
The organizer is signed in, owns a draft tournament, and opens its registration-question editor. The form has a nonempty title and no saved questions. The selected question type is *Number* for an amount.

=== Input Equivalence Classes
#epic-ec-table(
  ("EC1", "Question label", [nonempty], true),
  ("EC2", "Question label", [empty], false),
  ("EC3", "Number range", [minimum ≤ maximum], true),
  ("EC4", "Number range", [minimum > maximum], false),
)

=== Test Cases
#epic-test-cases(
  story: "US 3-5",
  (
    id: "TC56",
    inputs: (Label: [`Budget amount`], Minimum: [`0`], Maximum: [`1000`]),
    expected: [The Number question and its range are saved and appear for registrants.],
    actual: [Number question and range remained after reload and appeared for a registrant.],
    screenshots: ("../img/tc56.png", "../img/tc56-registrant.png"),
    status: "P",
    ecs: "EC1, EC3",
  ),
  (
    id: "TC57",
    inputs: (Label: [empty], Minimum: [`0`], Maximum: [`1000`]),
    expected: [Saving is blocked, a label error appears, and the saved form stays unchanged.],
    actual: [Save blocked with a label error.],
    screenshots: ("../img/tc57.png",),
    status: "P",
    ecs: "EC2, EC3",
  ),
  (
    id: "TC58",
    inputs: (Label: [`Budget amount`], Minimum: [`1000`], Maximum: [`100`]),
    expected: [Save fails.],
    actual: [Save fails.],
    screenshots: ("../img/tc58.png",),
    status: "P",
    ecs: "EC1, EC4",
  ),
)

// ---------- Story 3 (Imtaez) ----------
== US 3-6: Review participant applications
_As an organizer, I want to approve or reject submitted participant applications so that I control who competes in my tournament._

=== Pre-condition
The organizer is signed in and views a tournament with a locked application whose submitted answers can be opened. Each case starts with an undecided application.

=== Input Equivalence Classes
#epic-ec-table(
  ("EC1", "Review decision", [accept with space available], true),
  ("EC2", "Review decision", [reject], true),
  ("EC3", "Review decision", [accept when full], false),
)

=== Test Cases
#epic-test-cases(
  story: "US 3-6",
  (
    id: "TC59",
    inputs: (Decision: [Accept], Capacity: [1 accepted, limit 2], Application: [locked with submitted answers]),
    expected: [The answers are visible; the application becomes Accepted and the accepted count increases to 2.],
    actual: [Answers were visible; acceptance increased the count to 2.],
    screenshots: ("../img/tc59-answers.png", "../img/tc59.png"),
    status: "P",
    ecs: "EC1",
  ),
  (
    id: "TC60",
    inputs: (Decision: [Reject], Capacity: [1 accepted, limit 2], Application: [locked]),
    expected: [The application becomes Rejected, review buttons disappear, and the accepted count stays at 1.],
    actual: [Application rejected; review buttons removed; accepted count stayed at 1.],
    screenshots: ("../img/tc60.png",),
    status: "P",
    ecs: "EC2",
  ),
  (
    id: "TC61",
    inputs: (Decision: [Accept], Capacity: [2 accepted, limit 2], Application: [locked]),
    expected: [Accept can be clicked, but confirmation shows a capacity error; the application remains Waiting For Approval.],
    actual: [Capacity error shown; application remained Waiting For Approval.],
    screenshots: ("../img/tc61.png",),
    status: "P",
    ecs: "EC3",
  ),
)

// ---------- Story 4 (Imtaez) ----------
== US 3-1: Create a tournament
_As an organizer, I want to create a tournament so that I can define the event, funding needs, rules, solo/team size restrictions, tournament format, and rewards._

=== Pre-condition
The organizer is signed in and opens *Host a Tournament*. These cases cover the first page, which creates a Draft. Unless overridden, enter Tournament Name `Test Cup`, Game Title `Valorant`, Location `Online`, and Short Description `A local tournament.` Leave Tournament Details & Rules blank.

=== Input Equivalence Classes
#epic-ec-table(
  ("EC1", "Tournament name", [1–160 characters], true),
  ("EC2", "Tournament name", [empty], false),
  ("EC3", "Tournament name", [over 160 characters], false),
  ("EC4", "Game title", [1–80 characters], true),
  ("EC5", "Game title", [empty], false),
  ("EC6", "Game title", [over 80 characters], false),
  ("EC7", "Location", [1–160 characters], true),
  ("EC8", "Location", [empty], false),
  ("EC9", "Location", [over 160 characters], false),
  ("EC10", "Short description", [1–500 characters], true),
  ("EC11", "Short description", [empty], false),
  ("EC12", "Short description", [over 500 characters], false),
  ("EC13", "Details & rules", [0–10,000 characters], true),
  ("EC14", "Details & rules", [over 10,000 characters], false),
)

=== Test Cases
#epic-test-cases(
  story: "US 3-1",
  (
    id: "TC62",
    inputs: (Name: [`Test Cup`], Game: [`Valorant`], Location: [`Online`], Description: [`A local tournament.`]),
    expected: [A Draft tournament is created and the organizer is redirected to its management page.],
    actual: [Draft created; management page opened.],
    screenshots: ("../img/tc62.png",),
    status: "P",
    ecs: "EC1, EC4, EC7, EC10, EC13",
  ),
  (
    id: "TC63",
    inputs: (Name: [empty]),
    expected: [A name error appears and no tournament is created.],
    actual: [Name error shown; no new tournament.],
    screenshots: ("../img/tc63.png",),
    status: "P",
    ecs: "EC2, EC4, EC7, EC10, EC13",
  ),
  (
    id: "TC64",
    inputs: (Game: [empty]),
    expected: [A game error appears and no tournament is created.],
    actual: [Game error shown; no new tournament.],
    screenshots: ("../img/tc64.png",),
    status: "P",
    ecs: "EC1, EC5, EC7, EC10, EC13",
  ),
  (
    id: "TC65",
    inputs: (Location: [empty]),
    expected: [A location error appears and no tournament is created.],
    actual: [Location error shown; no new tournament.],
    screenshots: ("../img/tc65.png",),
    status: "P",
    ecs: "EC1, EC4, EC8, EC10, EC13",
  ),
  (
    id: "TC66",
    inputs: (Description: [empty]),
    expected: [A description error appears and no tournament is created.],
    actual: [Description error shown; no new tournament.],
    screenshots: ("../img/tc66.png",),
    status: "P",
    ecs: "EC1, EC4, EC7, EC11, EC13",
  ),
  (
    id: "TC67",
    inputs: (Name: [161 `A` characters]),
    expected: [The name field accepts at most 160 characters.],
    actual: [Name stopped at 160 characters.],
    screenshots: ("../img/tc67.png",),
    status: "P",
    ecs: "EC3, EC4, EC7, EC10, EC13",
  ),
  (
    id: "TC68",
    inputs: (Game: [81 `A` characters]),
    expected: [The game field accepts at most 80 characters.],
    actual: [Game stopped at 80 characters.],
    screenshots: ("../img/tc68.png",),
    status: "P",
    ecs: "EC1, EC6, EC7, EC10, EC13",
  ),
  (
    id: "TC69",
    inputs: (Location: [161 `A` characters]),
    expected: [The location field accepts at most 160 characters.],
    actual: [Location stopped at 160 characters.],
    screenshots: ("../img/tc69.png",),
    status: "P",
    ecs: "EC1, EC4, EC9, EC10, EC13",
  ),
  (
    id: "TC70",
    inputs: (Description: [501 `A` characters]),
    expected: [The description field accepts at most 500 characters.],
    actual: [Description stopped at 500 characters.],
    screenshots: ("../img/tc70.png",),
    status: "P",
    ecs: "EC1, EC4, EC7, EC12, EC13",
  ),
  (
    id: "TC71",
    inputs: (Details: [10,001 `A` characters]),
    expected: [The details field accepts at most 10,000 characters.],
    actual: [Details stopped at 10,000 characters.],
    screenshots: ("../img/tc71.png",),
    status: "P",
    ecs: "EC1, EC4, EC7, EC10, EC14",
  ),
)
