package models

func All() []any {
	return []any{
		&Role{},
		&Account{},
		&AccountRole{},
		&OrganizerProfile{},
		&SponsorProfile{},
		// &Tournament{}, // TODO: not implemented
	}
}
