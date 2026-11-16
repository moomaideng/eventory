#import "appendix/preambles.typ": *
#show: doc.with(title: "Test Plan")

// [Anti-Slop Indicator: 100% - @wachi - check ok] \
#cover(
  title: "Test Plan",
  date: datetime(year: 2026, month: 11, day: 16),
  contrib: (
    [Test plan], // Hoof
    [Test plan], // Tawoof
    [Test plan], // Noof
    [Test plan], // Filmfall
    [Test plan, Playwright], // Eff
    [Document organization], // Wachii
    [Test plan, GoTest, Playwright], // Mavin
    [Test plan, GoTest, Playwright], // Imtaez
    [GoTest], // Boof
  ),
)

#outline(depth: 2)

#include "test-plan/epic-a.typ"
#include "test-plan/epic-b.typ"

= Test Coverage Matrix
// Built automatically from every test-cases(...) above. Do not edit.
#coverage-matrix()

