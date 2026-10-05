package models

func All() []any {
	return []any{
		&Tournament{},
		&TournamentTeam{},
		&TournamentTeamMember{},
		&TournamentFunding{},
		&TournamentStatusChange{},
	}
}
