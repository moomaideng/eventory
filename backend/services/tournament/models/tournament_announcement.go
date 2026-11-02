package models

import (
	"time"

	"github.com/google/uuid"
)

// TournamentAnnouncement is a message an organizer or admin posts to a
// tournament's announcement feed (US3-4). Announcements are append-only:
// there is no edit or delete in this story.
type TournamentAnnouncement struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TournamentID uuid.UUID  `gorm:"type:uuid;not null;index:idx_tournament_announcements_feed,priority:1"`
	Tournament   Tournament `gorm:"foreignKey:TournamentID;constraint:OnDelete:CASCADE;"`

	Title string `gorm:"type:varchar(150);not null"`
	Body  string `gorm:"type:text;not null"`

	// AuthorAccountID is the account that posted the announcement.
	AuthorAccountID uuid.UUID `gorm:"type:uuid;not null;index"`
	// AuthorDisplayName is captured at posting time so the feed keeps showing
	// who posted, even if the account later renames.
	AuthorDisplayName string `gorm:"type:varchar(64);not null"`

	CreatedAt time.Time `gorm:"not null;autoCreateTime;index:idx_tournament_announcements_feed,priority:2,sort:desc"`
}
