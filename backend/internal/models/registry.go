package models

func All() []any {
	return []any{
		&Account{},
		&OrganizerProfile{},
		&SponsorProfile{},
		&Tournament{},
		&TournamentTeam{},
		&TournamentFunding{},
		// TODO??????
	}
}
