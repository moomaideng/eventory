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
#let team = (
  (id: "6732005321", name: "Kittichet Arayasujin", role: "Product Owner"), // Hoof
  (id: "6732023621", name: "Pitiwat Kittiwimonchai", role: "Tech Lead"), // Tawoof
  (id: "6731305421", name: "Kittichon Chaonawig", role: "Senior Backend"), // Noof
  (id: "6732007621", name: "Jirapat Chaiya", role: "Developer"), // Filmfall
  (id: "6732029421", name: "Passakorn Yangmak", role: "Scrum Master"), // Eff
  (id: "6732037421", name: "Wachiraphut Intarachart", role: "Developer"), // Wachii
  (id: "6730434521", name: "Metavin Karnjanarat", role: "Developer"), // Mavin
  (id: "6732022021", name: "Nirawit Jongprasartsuk", role: "Developer"), // Imtaez
  (id: "6732025921", name: "Pasin Thanyakasikol", role: "Developer"), // Boof
)
