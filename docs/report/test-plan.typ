#import "appendix/preambles.typ": *
#show: doc.with(title: "Test Plan")

// [Anti-Slop Indicator: 100% - @wachi - check ok] \
#let cover-work = (
  [Test plan], // Hoof
  [Test plan], // Tawoof
  [Test plan], // Noof
  [Document organization], // Filmfall
  [Test plan, Playwright], // Eff
  [Test plan], // Wachii
  [Test plan, Jest], // Mavin
  [Test plan, Jest], // Imtaez
  [Document organization], // Boof
)

#page(numbering: none)[
  #set align(center)
  #set par(justify: false)
  #v(2cm)
  #text(18pt, weight: "bold")[Test Plan] \
  #text(18pt, weight: "bold")[Eventory]
  #v(2cm)
  #text(13pt)[Presented by]
  #v(0.4cm)
  #table(
    columns: (2.6fr, auto, 2fr),
    stroke: none,
    inset: (x: 5pt, y: 4pt),
    align: left,
    ..team.enumerate().map(((i, member)) => (
      if i == 0 { [#member.name (Leader)] } else { member.name },
      member.id,
      cover-work.at(i),
    )).flatten(),
  )
  #v(1fr)
  2110423 Software Engineering \
  November 16, 2026 \
  Department of Computer Engineering, Faculty of Engineering \
  Chulalongkorn University
]

#outline(depth: 2)

#include "test-plan/epic-a.typ"
#include "test-plan/epic-b.typ"

= Test Coverage Matrix
// Built automatically from every test-cases(...) above. Do not edit.
#coverage-matrix()

#include "appendix/iso29110.typ"
