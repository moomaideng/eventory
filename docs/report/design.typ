
#import "appendix/preambles.typ": *
#show: doc.with(title: "Design Document")

// [Anti-Slop Indicator: 100% - @wachi - check ok] \
// TODO (Wachi) before submit:
//   [ ] Fill contribution for Eff, Wachii, Mavin, Imtaez, Boof (red "TODO: fill" on cover)
//   [ ] Confirm the 4 entries below are still accurate (arch.typ by Noof not listed)
//   [ ] Rebuild PDF: zero red TODO boxes

#cover(
  title: "Design Document",
  date: datetime(year: 2026, month: 11, day: 16),
  contrib: (
    [UML API, sys arch], // Hoof
    [UML API], // Tawoof
    [UML API, sys arch], // Noof
    [Sequence diagram, design quality], // Filmfall
    [UML API], // Eff
    [Document organization], // Wachii
    [UML API], // Mavin
    [UML API, sys arch], // Imtaez
    [Sequence diagram], // Boof
  ),
)

#outline()

= Architecture
#include "design/arch.typ"
#include "design/uml-api.typ"
#include "design/sequence.typ"
#include "design/quality.typ"
