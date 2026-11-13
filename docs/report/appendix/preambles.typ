// Shared style + helpers. Import in every document:
//   #import "appendix/preambles.typ": *   (or "../appendix/..." from a subfolder)

#import "team.typ": *

// ---------- General helpers ----------

// Unfinished part. Shows as a red box so it can't be missed.
#let todo(body) = box(
  fill: red.lighten(85%),
  stroke: red,
  inset: (x: 4pt, y: 2pt),
  radius: 2pt,
  text(fill: red.darken(30%), size: 0.9em)[*TODO:* #body],
)

// Grey box standing in for a diagram that isn't exported yet.
#let placeholder(height: 6cm, body) = rect(
  width: 100%,
  height: height,
  fill: luma(240),
  stroke: (dash: "dashed", paint: gray),
  align(center + horizon, text(fill: gray.darken(40%), body)),
)

// ---------- Cover page ----------
// `contrib` = one entry per member, same order as `team`:
//   ((pct: 20, work: [UML API diagram]), ...)
#let cover(title: "", date: datetime.today(), contrib: ()) = page(numbering: none)[
  #set align(center)
  #v(3cm)
  #text(14pt, course) \
  #text(11pt, semester)
  #v(1.5cm)
  #text(28pt, weight: "bold", project) \
  #v(4pt)
  #text(18pt, title)
  #v(2cm)

  #set align(left)
  #set par(justify: false)
  // `role` may be missing, a string, or an array of strings.
  #let roles(m) = {
    let r = m.at("role", default: ())
    if type(r) == str { r } else { r.join(", ", default: "") }
  }
  // Role column only appears when someone has a role.
  #let show-role = team.any(m => roles(m) != "")
  #table(
    columns: if show-role { (auto, 1fr, auto, auto, 1.5fr) } else { (auto, 1fr, auto, 2fr) },
    align: if show-role { (left, left, left, center, left) } else { (left, left, center, left) },
    table.header(
      [*Student ID*], [*Name*],
      ..if show-role { ([*Role*],) },
      [*%*], [*Contribution*],
    ),
    ..team.enumerate().map(((i, m)) => {
      let c = contrib.at(i, default: (pct: none, work: todo[fill]))
      (
        m.id,
        m.name,
        ..if show-role { (roles(m),) },
        if c.pct == none [--] else [#c.pct],
        c.work,
      )
    }).flatten(),
  )

  #v(1fr)
  #set align(center)
  #date.display("[month repr:long] [day], [year]")
]

// ---------- Test plan helpers ----------

// Every test case is recorded here so `coverage-matrix()` builds itself.
#let _cases = state("test-cases", ())

// Input equivalence classes of one user story.
// Each row: (EC ID, input, condition, valid?)
//   ("EC1", "Email", [has `@` and domain, e.g. `a@b.com`], true)
#let ec-table(..rows) = table(
  columns: (auto, auto, 1fr, auto),
  table.header[*EC*][*Input*][*Condition*][*Valid?*],
  ..rows.pos().map(((id, input, cond, ok)) =>
    (id, input, cond, if ok [✓ Valid] else [✗ Invalid])
  ).flatten(),
)

// Test cases of one user story.
// Each case: (id, inputs, expected, actual, ecs, status)
//   status: "P" pass, "F" fail, none = not run yet
#let test-cases(story: "", ..cases) = {
  let cases = cases.pos()
  _cases.update(all => all + cases.map(c => (story: story, id: c.id, status: c.at("status", default: none))))
  set text(9pt)
  table(
    columns: (auto, 1fr, 1fr, 1fr, auto, auto),
    table.header[*ID*][*Inputs*][*Expected Output*][*Actual Output*][*EC Covered*][*Status*],
    ..cases.map(c => (
      c.id,
      c.inputs,
      c.expected,
      c.at("actual", default: [--]),
      c.ecs,
      { let st = c.at("status", default: none); if st == none [--] else [#st] },
    )).flatten(),
  )
}

// Test Coverage Matrix: one row per user story, built from all `test-cases` above it.
#let coverage-matrix() = context {
  let all = _cases.final()
  let stories = all.map(c => c.story).dedup()
  table(
    columns: (auto, 1fr, 1fr),
    table.header[*User Story ID*][*Test Case IDs*][*Status*],
    ..stories.map(s => {
      let cs = all.filter(c => c.story == s)
      (s, cs.map(c => c.id).join(", "), cs.map(c => if c.status == none { "–" } else { c.status }).join(", "))
    }).flatten(),
  )
}

// ---------- Document setup ----------
// Usage: #show: doc.with(title: "Design Document")
#let doc(title: "", body) = {
  set document(title: project + " — " + title)
  set page(paper: "a4", margin: 2.5cm, numbering: "1")
  // Thai text falls back to Noto Serif Thai.
  set text(font: ("Libertinus Serif", "Noto Serif Thai"), size: 11pt)
  set par(justify: true)
  set heading(numbering: "1.1")
  show heading.where(level: 1): it => { pagebreak(weak: true); it; v(4pt) }
  set table(stroke: 0.5pt + gray, inset: 6pt)
  set figure(gap: 8pt)
  body
}
