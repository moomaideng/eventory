package seeds

import "gorm.io/gorm"

type Seeds struct {
	Name string
	Run  func(db *gorm.DB) error
}

func All() []Seeds {
	return []Seeds{
		{
			Name: "Accounts",
			Run:  SeedAccounts,
		},
		{
			Name: "Tournaments",
			Run:  SeedTournaments,
		},
	}
}
