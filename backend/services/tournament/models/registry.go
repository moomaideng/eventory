package models

func All() []any {
	return []any{
		&Tournament{},
		&TournamentStaff{},
		&TournamentTeam{},
		&TournamentTeamMember{},
		&TournamentTeamReview{},
		&TournamentAnnouncement{},
		&SponsorshipTier{},
		&Match{},
		&MatchEntry{},
		&SponsorshipPledge{},
	}
}
