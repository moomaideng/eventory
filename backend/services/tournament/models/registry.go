package models

func All() []any {
	return []any{
		&Tournament{},
		&TournamentStaff{},
		&TournamentTeam{},
		&TournamentTeamMember{},
		&TournamentFunding{},
		&TournamentTeamReview{},
		&TournamentAnnouncement{},
		&SponsorshipTier{},
		&Match{},
		&MatchEntry{},
		&SponsorshipPledge{},
		&SponsorshipLogo{},
	}
}
