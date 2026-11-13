// ============================================================
// Owners: Ham, Tawan, Not
// Task:   UML API diagram + short explanation
//
// 1. Draw a UML class diagram of the API layer, one box per
//    service. Each box lists:
//    endpoints as operations, request/response DTOs, and links to
//    the domain classes it uses.
//    Sources: Swagger UI (endpoints), backend/services/*/
//    Reuse the style of docs/arch/c4-level4-code-teamlobby.puml (PlantUML).
// 2. Export SVG/PNG into docs/report/img/, then swap the
//    placeholder for the commented #figure line.
// 3. Replace every todo[...] with the real text.
// ============================================================
#import "../appendix/preambles.typ": *

== UML API Diagram <uml-api>

#placeholder[UML API diagram]
// #figure(image("../img/uml-api.svg", width: 100%), caption: [UML API diagram])

=== Overview
#todo[2--3 sentences: what the diagram shows and how to read it]

=== Services and Endpoints
#table(
  columns: (auto, 1fr, 1fr),
  table.header[*Service / Class*][*Responsibility*][*Key operations*],
  [Account], todo[one line], todo[e.g. `register()`, `login()`],
  [Tournament], todo[one line], todo[...],
  [Payment], todo[one line], todo[...],
)

=== Relationships
#todo[explain the important associations / dependencies in the diagram]
