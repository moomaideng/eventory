package seeds

import "gorm.io/gorm"

// Task defines a single seeding step
type Task struct {
	Name string
	Run  func(db *gorm.DB) error
}

// All returns the explicit list of seeders in the exact order they must execute.
func All() []Task {
	return []Task{
		{
			Name: "Roles and Accounts",
			Run:  SeedAccounts,
		},
		// {
		// 	Name: "Tournaments",
		// 	Run:  SeedTournaments, // TODO: not implemented
		// },
	}
}
