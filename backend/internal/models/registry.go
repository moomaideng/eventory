package models

func All() []any {
	return []any{
		&Account{},
		&OrganizerProfile{},
		&SponsorProfile{},
		// ... TODO: implement other models
	}
}
