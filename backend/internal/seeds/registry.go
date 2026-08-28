package seeds

import "gorm.io/gorm"

type Task struct {
	Name string
	Run  func(db *gorm.DB) error
}

func All() []Task {
	return []Task{
		{
			Name: "Accounts",
			Run:  SeedAccounts,
		},
		// ... TODO: implement other tasks
	}
}
