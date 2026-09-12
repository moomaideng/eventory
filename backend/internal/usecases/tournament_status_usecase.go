package usecases

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/internal/models"
	"github.com/moomaideng/eventory/internal/repositories"
)

var (
	ErrInvalidStatusTransition = errors.New("this status change is not allowed")
	ErrStatusUnchanged         = errors.New("tournament is already in that status")
	ErrOverrideReasonRequired  = errors.New("a reason between 10 and 500 characters is required")
	ErrUnknownStatus           = errors.New("unknown tournament status")
)

const (
	overrideReasonMinLength = 10
	overrideReasonMaxLength = 500
)

// allowedTransitions maps each status to the statuses an organizer may manually
// override it to.
//
// COMPLETED and CANCELLED are terminal: completed tournaments are a read-only
// archive, and cancelling triggers batch refunds that cannot be reversed by
// flipping a status back.
//
// DRAFT and CROWDFUNDING never appear as targets. They are declared in the
// model as placeholders (see TODO markers there) but no screen or flow handles
// them yet, so moving a tournament into one would strand it in a state the app
// cannot render. CROWDFUNDING does appear as a *source*: force-starting an
// under-funded campaign into REGISTRATION_OPEN is a supported override, and the
// reason field is where the organizer records accepting that risk.
var allowedTransitions = map[models.TournamentStatus][]models.TournamentStatus{
	models.TournamentStatusDraft: {
		models.TournamentStatusRegistrationOpen,
		models.TournamentStatusCancelled,
	},
	models.TournamentStatusCrowdfunding: {
		models.TournamentStatusRegistrationOpen,
		models.TournamentStatusCancelled,
	},
	models.TournamentStatusRegistrationOpen: {
		models.TournamentStatusRegistrationClosed,
		models.TournamentStatusOngoing, // skip ahead when registration is cut short
		models.TournamentStatusCompleted,
		models.TournamentStatusCancelled,
	},
	models.TournamentStatusRegistrationClosed: {
		models.TournamentStatusRegistrationOpen, // reopen after a delay
		models.TournamentStatusOngoing,
		models.TournamentStatusCompleted,
		models.TournamentStatusCancelled,
	},
	models.TournamentStatusOngoing: {
		models.TournamentStatusRegistrationClosed, // started early by mistake
		models.TournamentStatusCompleted,
		models.TournamentStatusCancelled,
	},
	models.TournamentStatusCompleted: {}, // terminal
	models.TournamentStatusCancelled: {}, // terminal
}

// TournamentStatusOverride is the result of a successful manual override.
type TournamentStatusOverride struct {
	TournamentID uuid.UUID
	FromStatus   models.TournamentStatus
	ToStatus     models.TournamentStatus
	Reason       string
	Change       models.TournamentStatusChange
}

type TournamentStatusUseCase struct {
	repo repositories.TournamentStatusRepository
}

func NewTournamentStatusUseCase(repo repositories.TournamentStatusRepository) *TournamentStatusUseCase {
	return &TournamentStatusUseCase{repo: repo}
}

// Override manually moves an owned tournament to a new status, recording who
// did it and why.
//
// TODO(sprint-next): notify participants, staff, and sponsors when a status
// changes. No notification system exists in the codebase yet, so the "notifies
// related parties" half of US3-3's acceptance criteria is deliberately out of
// scope for this story.
//
// TODO(US6-x): warn when a tournament is completed without a recorded winner.
// Not implementable today: no winner, placement, or bracket model exists, so
// there is nothing to check. The frontend shows an unconditional "this is
// final" confirmation instead.
func (u *TournamentStatusUseCase) Override(
	ctx context.Context,
	accountID, tournamentID uuid.UUID,
	rawStatus, rawReason string,
) (*TournamentStatusOverride, error) {
	to := models.TournamentStatus(strings.ToUpper(strings.TrimSpace(rawStatus)))
	if _, known := allowedTransitions[to]; !known {
		return nil, ErrUnknownStatus
	}

	reason := strings.TrimSpace(rawReason)
	if n := utf8.RuneCountInString(reason); n < overrideReasonMinLength || n > overrideReasonMaxLength {
		return nil, ErrOverrideReasonRequired
	}

	var from models.TournamentStatus
	change, err := u.repo.ApplyOverride(ctx, accountID, tournamentID, reason,
		func(current *models.Tournament) (models.TournamentStatus, error) {
			from = current.Status
			if current.Status == to {
				return "", ErrStatusUnchanged
			}
			if !canTransition(current.Status, to) {
				return "", ErrInvalidStatusTransition
			}
			return to, nil
		})
	if err != nil {
		if errors.Is(err, repositories.ErrTournamentNotFound) {
			return nil, ErrTournamentNotFound
		}
		return nil, err
	}

	return &TournamentStatusOverride{
		TournamentID: tournamentID, FromStatus: from, ToStatus: to,
		Reason: reason, Change: *change,
	}, nil
}

// TournamentStatusDetail describes where a tournament stands and where an
// organizer may move it next.
type TournamentStatusDetail struct {
	CurrentStatus models.TournamentStatus
	// AllowedTransitions is served to the client so the status modal offers
	// only reachable options. Keeping the transition table on the server means
	// the UI cannot drift out of step with the rules it is meant to reflect.
	AllowedTransitions []models.TournamentStatus
	Changes            []models.TournamentStatusChange
}

// History returns an owned tournament's current status, the statuses it may be
// moved to, and its override trail, newest first.
func (u *TournamentStatusUseCase) History(
	ctx context.Context,
	accountID, tournamentID uuid.UUID,
) (*TournamentStatusDetail, error) {
	history, err := u.repo.ListHistory(ctx, accountID, tournamentID)
	if err != nil {
		if errors.Is(err, repositories.ErrTournamentNotFound) {
			return nil, ErrTournamentNotFound
		}
		return nil, err
	}
	return &TournamentStatusDetail{
		CurrentStatus:      history.CurrentStatus,
		AllowedTransitions: AllowedTransitionsFrom(history.CurrentStatus),
		Changes:            history.Changes,
	}, nil
}

// canTransition reports whether `to` is a permitted manual override from `from`.
// An unrecognised source status permits nothing, so bad data fails closed.
func canTransition(from, to models.TournamentStatus) bool {
	for _, candidate := range allowedTransitions[from] {
		if candidate == to {
			return true
		}
	}
	return false
}

// AllowedTransitionsFrom lists the statuses an organizer may move `from` to.
// The handler exposes this so the modal only offers reachable options.
func AllowedTransitionsFrom(from models.TournamentStatus) []models.TournamentStatus {
	allowed := allowedTransitions[from]
	out := make([]models.TournamentStatus, len(allowed))
	copy(out, allowed)
	return out
}
