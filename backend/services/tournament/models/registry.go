package models

func All() []any {
	return []any{
		&Tournament{},
		&TournamentStaff{},
		&TournamentTeam{},
		&TournamentTeamMember{},
		&SoloRegistration{},
		&TournamentFunding{},
		&TournamentStatusChange{},
		&TournamentTeamReview{},
	}
}
