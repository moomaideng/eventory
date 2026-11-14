
#import "appendix/preambles.typ": *
#show: doc.with(title: "Design Document")

// [Anti-Slop Indicator: 100% - @wachi - check ok] \
// ไว้ค่อยมาใส่ contribution เกลี่ยทีหลัง

#cover(
  title: "Design Document",
  date: datetime(year: 2026, month: 11, day: 16),
  contrib: (
    (pct: none, work: [UML API diagram]),     // Ham
    (pct: none, work: [UML API diagram]),     // Tawan
    (pct: none, work: [UML API diagram]),     // Not
    (pct: none, work: [Sequence diagram, design quality]), // Film
  ),
)

#outline()

= Architecture
#include "design/uml-api.typ"
#include "design/sequence.typ"
#include "design/quality.typ"

#include "appendix/iso29110.typ"
