package models

func All() []any {
	return []any{
		&Tournament{},
		&TournamentStaff{},
		&TournamentTeam{},
		&TournamentTeamMember{},
		&TournamentFunding{},
		&TournamentStatusChange{},
	}
}
