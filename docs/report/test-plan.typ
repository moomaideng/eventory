#import "appendix/preambles.typ": *
#show: doc.with(title: "Test Plan")

// [Anti-Slop Indicator: 100% - @wachi - check ok] \
// ไว้ค่อยมาใส่ contribution เกลี่ยทีหลัง

#cover(
  title: "Test Plan",
  date: datetime(year: 2026, month: 11, day: 16),
  contrib: (
    (pct: none, work: [--]), // Ham
    (pct: none, work: [--]), // Tawan
    (pct: none, work: [--]), // Not
    (pct: none, work: [--]), // Film
    (pct: none, work: [Test plan: EPIC A]), // Eff
    (pct: none, work: [Test plan: EPIC A]), // Pun
    (pct: none, work: [Test plan: EPIC B]), // Mavin
    (pct: none, work: [Test plan: EPIC B]), // Imtaez
  ),
)

#outline(depth: 2)

#include "test-plan/epic-a.typ"
#include "test-plan/epic-b.typ"

= Test Coverage Matrix
// Built automatically from every test-cases(...) above. Do not edit.
#coverage-matrix()

#include "appendix/iso29110.typ"
