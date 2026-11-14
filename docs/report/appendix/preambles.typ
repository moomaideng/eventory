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

// Test cases of one user story. Each case is a vertical card: field | value.
// Fields (only `id` + `expected` required; empty fields are hidden):
//   id:       "TC01"
//   pre:      [per-case pre-condition]
//   action:   [what the tester does]
//   inputs:   (Email: [`a@b.com`], Password: [`123`])   → one row per input
//             or plain content: [email = `a@b.com`]
//   expected: [text + screenshot]   e.g. [Logged in \ #image("../img/tc01.png", width: 80%)]
//   actual:   [after running]       default "--"
//   ecs:      "EC1, EC3"
//   status:   "P" / "F"             none = not run yet
#let test-cases(story: "", ..cases) = {
  let cases = cases.pos()
  _cases.update(all => all + cases.map(c => (story: story, id: c.id, status: c.at("status", default: none))))
  set par(justify: false)
  for c in cases {
    let row(label, value) = ([*#label*], table.cell(colspan: 2, value))
    let inputs = c.at("inputs", default: none)
    let input-rows = if type(inputs) == dictionary {
      let pairs = inputs.pairs()
      (table.cell(rowspan: pairs.len(), [*Inputs*]),)
      for (k, v) in pairs { (k, v) }
    } else if inputs != none { row("Inputs", inputs) } else { () }
    let opt(key, label) = {
      let v = c.at(key, default: none)
      if v == none { () } else { row(label, v) }
    }
    let st = c.at("status", default: none)
    block(breakable: true, below: 1.2em, table(
      columns: (3.4cm, 3.2cm, 1fr),
      align: (left + horizon, left + horizon, left),
      ..row("Test Case ID", c.id),
      ..opt("pre", "Pre-condition"),
      ..opt("action", "Action"),
      ..input-rows,
      ..row("Expected Output", c.expected),
      ..row("Actual Output", c.at("actual", default: [--])),
      ..opt("ecs", "Classes Covered"),
      ..row("Status", if st == none [--] else [#st]),
    ))
  }
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
