// Project + team info, shown on the cover of every document.
//   project, course, semester → title block of the cover
//   team                      → member table on the cover

#let project = "Eventory"
#let course = "2110423 Software Engineering"
#let semester = "First Semester, Academic Year 2026"

// Each member: (id: "student ID", name: "full name")
// Role (optional): one role or several → "Role" column on the cover
//   (id: "...", name: "...", role: "Leader")
//   (id: "...", name: "...", role: ("Scrum Master", "Developer"))
// Order matters: the `contrib` list in design.typ / test-plan.typ
// follows this order, row by row.
// 
// CHANGE BY YOURSELF
#let team = (
  (id: "6xxxxxxx21", name: "Kittichet Arayasujin", role: "Product Owner"), // Ham
  (id: "6xxxxxxx21", name: "", role: "Tech Lead"), // Tawan
  (id: "6xxxxxxx21", name: "", role: "Senior Backend"), // Not
  (id: "6xxxxxxx21", name: "", role: ""), // Film
  (id: "6xxxxxxx21", name: "", role: "Scrum Master"), // Eff
  (id: "6732037421", name: "Wachiraphut Intarachart", role: "Dev dogshit"), // Pun
  (id: "6xxxxxxx21", name: "", role: ""), // Mavin
  (id: "6xxxxxxx21", name: "", role: ""), // Imtaez
  (id: "6xxxxxxx21", name: "", role: ""), // Boof
)
