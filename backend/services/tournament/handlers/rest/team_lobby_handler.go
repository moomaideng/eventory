package rest

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/tournament/internal/forms"
	"github.com/moomaideng/eventory/services/tournament/internal/usecases"
	"github.com/moomaideng/eventory/services/tournament/models"
)

type TeamLobbyMemberResponse struct {
	RegistrationComplete bool      `json:"registrationComplete" doc:"Answers satisfy the current registration questions"`
	ID                   uuid.UUID `json:"id"`
	AccountID            uuid.UUID `json:"accountId"`
	Handle               string    `json:"handle"`
	DisplayName          string    `json:"displayName"`
	Role                 string    `json:"role" enum:"CAPTAIN,MEMBER"`
	JoinedAt             time.Time `json:"joinedAt"`
}

type TeamLobbyResponse struct {
	PaymentID        *uuid.UUID                `json:"paymentId"`
	PaymentExpiresAt *time.Time                `json:"paymentExpiresAt"`
	ID               string                    `json:"id"`
	Tournament       TournamentResponse        `json:"tournament"`
	CaptainID        uuid.UUID                 `json:"captainId"`
	CaptainName      string                    `json:"captainName"`
	Name             string                    `json:"name"`
	InviteCode       *string                   `json:"inviteCode,omitempty" doc:"Six-character team invite code, or null if viewer is not a team member or registration is closed/solo"`
	Status           string                    `json:"status" enum:"FORMING,PAYMENT_PENDING,LOCKED,ACCEPTED,REJECTED"`
	Members          []TeamLobbyMemberResponse `json:"members"`
	ViewerRole       string                    `json:"viewerRole" enum:"CAPTAIN,MEMBER,INVITEE"`
	CreatedAt        time.Time                 `json:"createdAt"`
}

type TeamLobbyOutput struct {
	Body TeamLobbyResponse
}

type CreateTeamLobbyOutput struct {
	Status int
	Body   TeamLobbyResponse
}

type EmptyTeamLobbyOutput struct {
	Status int
}

type CreateTeamLobbyRequest struct {
	Answers []models.FieldAnswer `json:"answers,omitempty" doc:"Captain's registration answers"`
	Name    string               `json:"name" minLength:"1" maxLength:"120" doc:"Team name"`
}

type CreateTeamLobbyInput struct {
	TournamentID uuid.UUID `path:"tournamentId" doc:"Tournament UUID"`
	Body         CreateTeamLobbyRequest
}

type RegisterSoloRequest struct {
	Answers []models.FieldAnswer `json:"answers,omitempty" doc:"Submitted answers for custom registration form"`
}

type RegisterSoloInput struct {
	TournamentID uuid.UUID           `path:"tournamentId" doc:"Tournament UUID"`
	Body         RegisterSoloRequest `doc:"Registration form answers"`
}

type MyTournamentTeamInput struct {
	TournamentID uuid.UUID `path:"tournamentId" doc:"Tournament UUID"`
}

type InviteCodeInput struct {
	InviteCode string `path:"inviteCode" pattern:"^[A-Za-z0-9]{6}$" doc:"Six-character team invite code"`
}

type InviteResolutionResponse struct {
	TournamentID uuid.UUID `json:"tournamentId"`
	TeamID       string    `json:"teamId"`
}

type InviteResolutionOutput struct {
	Body InviteResolutionResponse
}

type TeamIDInput struct {
	TournamentID uuid.UUID `path:"tournamentId" doc:"Tournament UUID"`
	TeamID       string    `path:"teamId" pattern:"^[A-Za-z0-9]{8}$" doc:"Eight-character team ID"`
}

type JoinTeamLobbyRequest struct {
	InviteCode string               `json:"inviteCode" pattern:"^[A-Za-z0-9]{6}$" doc:"Six-character team invite code"`
	Answers    []models.FieldAnswer `json:"answers,omitempty" doc:"Submitted answers for custom registration form"`
}

type JoinTeamLobbyInput struct {
	TournamentID uuid.UUID `path:"tournamentId" doc:"Tournament UUID"`
	TeamID       string    `path:"teamId" pattern:"^[A-Za-z0-9]{8}$" doc:"Eight-character team ID"`
	Body         JoinTeamLobbyRequest
}

type MemberRegistrationInput struct {
	TournamentID uuid.UUID `path:"tournamentId" doc:"Tournament UUID"`
	TeamID       string    `path:"teamId" pattern:"^[A-Za-z0-9]{8}$" doc:"Eight-character team ID"`
	Body         RegisterSoloRequest
}

type LobbyMemberInput struct {
	TournamentID uuid.UUID `path:"tournamentId" doc:"Tournament UUID"`
	TeamID       string    `path:"teamId" pattern:"^[A-Za-z0-9]{8}$" doc:"Eight-character team ID"`
	MemberID     uuid.UUID `path:"memberId" doc:"Team member UUID"`
}

func registerLobbyOp[In any, Out any](
	api huma.API,
	auth func(huma.Context, func(huma.Context)),
	op huma.Operation,
	handler func(context.Context, *In) (*Out, error),
) {
	op.Tags = []string{"Team lobbies"}
	op.Middlewares = huma.Middlewares{auth}
	op.Security = []map[string][]string{{"bearer": {}}}
	huma.Register(api, op, handler)
}

func runTeamLobbyAction[In any](
	ctx context.Context,
	input *In,
	fn func(context.Context, *In, uuid.UUID) (*usecases.TeamLobbyView, error),
) (uuid.UUID, *usecases.TeamLobbyView, error) {
	accountID, err := authenticatedAccountID(ctx)
	if err != nil {
		return uuid.Nil, nil, err
	}
	view, err := fn(ctx, input, accountID)
	if err != nil {
		return uuid.Nil, nil, teamLobbyHTTPError(err)
	}
	return accountID, view, nil
}

func teamLobbyHandler[In any](
	fn func(context.Context, *In, uuid.UUID) (*usecases.TeamLobbyView, error),
) func(context.Context, *In) (*TeamLobbyOutput, error) {
	return func(ctx context.Context, input *In) (*TeamLobbyOutput, error) {
		accountID, view, err := runTeamLobbyAction(ctx, input, fn)
		if err != nil {
			return nil, err
		}
		return &TeamLobbyOutput{Body: toTeamLobbyResponse(view, accountID)}, nil
	}
}

func createdTeamLobbyHandler[In any](
	fn func(context.Context, *In, uuid.UUID) (*usecases.TeamLobbyView, error),
) func(context.Context, *In) (*CreateTeamLobbyOutput, error) {
	return func(ctx context.Context, input *In) (*CreateTeamLobbyOutput, error) {
		accountID, view, err := runTeamLobbyAction(ctx, input, fn)
		if err != nil {
			return nil, err
		}
		return &CreateTeamLobbyOutput{Status: http.StatusCreated, Body: toTeamLobbyResponse(view, accountID)}, nil
	}
}

func registerTeamIDCommand(
	api huma.API,
	auth func(huma.Context, func(huma.Context)),
	op huma.Operation,
	fn func(context.Context, uuid.UUID, string, uuid.UUID) (*usecases.TeamLobbyView, error),
) {
	registerLobbyOp(api, auth, op, teamLobbyHandler(func(ctx context.Context, input *TeamIDInput, accountID uuid.UUID) (*usecases.TeamLobbyView, error) {
		return fn(ctx, input.TournamentID, input.TeamID, accountID)
	}))
}

func RegisterTeamLobbyRoutes(
	api huma.API,
	auth func(huma.Context, func(huma.Context)),
	lobbyUseCase *usecases.TeamLobbyUseCase,
) {
	registerLobbyOp(api, auth, huma.Operation{
		OperationID: "get-my-tournament-team",
		Method:      http.MethodGet,
		Path:        "/api/v1/tournaments/{tournamentId}/my-team",
		Summary:     "View the authenticated competitor's team for a tournament",
	}, teamLobbyHandler(func(ctx context.Context, input *MyTournamentTeamInput, accountID uuid.UUID) (*usecases.TeamLobbyView, error) {
		return lobbyUseCase.GetActiveForTournament(ctx, input.TournamentID, accountID)
	}))

	registerLobbyOp(api, auth, huma.Operation{
		OperationID: "register-solo-tournament",
		Method:      http.MethodPost,
		Path:        "/api/v1/tournaments/{tournamentId}/solo-register",
		Summary:     "Register for a solo tournament",
		Description: "Automatically creates and locks a 1-person team for a solo tournament, validating custom registration form answers and initiating payment if an entry fee is required.",
	}, createdTeamLobbyHandler(func(ctx context.Context, input *RegisterSoloInput, accountID uuid.UUID) (*usecases.TeamLobbyView, error) {
		return lobbyUseCase.RegisterSolo(ctx, input.TournamentID, accountID, input.Body.Answers)
	}))

	registerLobbyOp(api, auth, huma.Operation{
		OperationID: "create-team-lobby",
		Method:      http.MethodPost,
		Path:        "/api/v1/tournaments/{tournamentId}/teams",
		Summary:     "Create a team lobby",
		Description: "Creates a forming team and makes the authenticated account its captain.",
	}, createdTeamLobbyHandler(func(ctx context.Context, input *CreateTeamLobbyInput, accountID uuid.UUID) (*usecases.TeamLobbyView, error) {
		return lobbyUseCase.Create(ctx, input.TournamentID, accountID, input.Body.Name, input.Body.Answers)
	}))

	registerLobbyOp(api, auth, huma.Operation{
		OperationID: "resolve-team-invite",
		Method:      http.MethodGet,
		Path:        "/api/v1/lobbies/{inviteCode}",
		Summary:     "Resolve a team invite code to tournament and team IDs",
	}, func(ctx context.Context, input *InviteCodeInput) (*InviteResolutionOutput, error) {
		accountID, err := authenticatedAccountID(ctx)
		if err != nil {
			return nil, err
		}
		res, err := lobbyUseCase.ResolveByInviteCode(ctx, input.InviteCode, accountID)
		if err != nil {
			return nil, teamLobbyHTTPError(err)
		}
		return &InviteResolutionOutput{
			Body: InviteResolutionResponse{
				TournamentID: res.TournamentID,
				TeamID:       res.TeamID,
			},
		}, nil
	})

	registerTeamIDCommand(api, auth, huma.Operation{
		OperationID: "get-team-lobby",
		Method:      http.MethodGet,
		Path:        "/api/v1/tournaments/{tournamentId}/teams/{teamId}",
		Summary:     "View a team lobby",
	}, lobbyUseCase.GetByID)

	registerLobbyOp(api, auth, huma.Operation{
		OperationID: "join-team-lobby",
		Method:      http.MethodPost,
		Path:        "/api/v1/tournaments/{tournamentId}/teams/{teamId}/join",
		Summary:     "Join a team lobby using an invite code",
	}, teamLobbyHandler(func(ctx context.Context, input *JoinTeamLobbyInput, accountID uuid.UUID) (*usecases.TeamLobbyView, error) {
		return lobbyUseCase.Join(ctx, input.TournamentID, input.TeamID, input.Body.InviteCode, accountID, input.Body.Answers)
	}))

	registerLobbyOp(api, auth, huma.Operation{
		OperationID: "update-member-registration-answers",
		Method:      http.MethodPut,
		Path:        "/api/v1/tournaments/{tournamentId}/teams/{teamId}/registration",
		Summary:     "Submit your own registration answers once while the team is forming",
	}, teamLobbyHandler(func(ctx context.Context, input *MemberRegistrationInput, accountID uuid.UUID) (*usecases.TeamLobbyView, error) {
		return lobbyUseCase.SaveRegistrationAnswers(ctx, input.TournamentID, input.TeamID, accountID, input.Body.Answers)
	}))

	registerTeamIDCommand(api, auth, huma.Operation{
		OperationID: "regenerate-team-lobby-invite",
		Method:      http.MethodPost,
		Path:        "/api/v1/tournaments/{tournamentId}/teams/{teamId}/invite/regenerate",
		Summary:     "Regenerate a team invite",
	}, lobbyUseCase.RegenerateInvite)

	registerTeamIDCommand(api, auth, huma.Operation{
		OperationID: "refresh-team-submission",
		Method:      http.MethodPost,
		Path:        "/api/v1/tournaments/{tournamentId}/teams/{teamId}/submission/refresh",
		Summary:     "Verify team registration payment",
	}, lobbyUseCase.RefreshSubmission)

	registerTeamIDCommand(api, auth, huma.Operation{
		OperationID: "lock-team-lobby",
		Method:      http.MethodPost,
		Path:        "/api/v1/tournaments/{tournamentId}/teams/{teamId}/lock",
		Summary:     "Lock a team roster",
	}, lobbyUseCase.Lock)

	registerLobbyOp(api, auth, huma.Operation{
		OperationID: "remove-team-lobby-member",
		Method:      http.MethodDelete,
		Path:        "/api/v1/tournaments/{tournamentId}/teams/{teamId}/members/{memberId}",
		Summary:     "Remove a lobby member",
	}, teamLobbyHandler(func(ctx context.Context, input *LobbyMemberInput, accountID uuid.UUID) (*usecases.TeamLobbyView, error) {
		return lobbyUseCase.RemoveMember(ctx, input.TournamentID, input.TeamID, input.MemberID, accountID)
	}))

	registerLobbyOp(api, auth, huma.Operation{
		OperationID: "disband-team-lobby",
		Method:      http.MethodDelete,
		Path:        "/api/v1/tournaments/{tournamentId}/teams/{teamId}",
		Summary:     "Disband a team lobby",
	}, func(ctx context.Context, input *TeamIDInput) (*EmptyTeamLobbyOutput, error) {
		accountID, err := authenticatedAccountID(ctx)
		if err != nil {
			return nil, err
		}
		if err := lobbyUseCase.Disband(ctx, input.TournamentID, input.TeamID, accountID); err != nil {
			return nil, teamLobbyHTTPError(err)
		}
		return &EmptyTeamLobbyOutput{Status: http.StatusNoContent}, nil
	})
}

func teamLobbyHTTPError(err error) error {
	if mapped := mapAccountPortError(err); mapped != nil {
		return mapped
	}
	switch {
	case errors.Is(err, usecases.ErrResponseAlreadyExists):
		return huma.Error409Conflict("Registration answers have already been submitted and cannot be edited", err)
	case errors.Is(err, usecases.ErrNotTeamMember):
		return huma.Error403Forbidden(err.Error(), err)
	case errors.Is(err, usecases.ErrMemberAnswersIncomplete), errors.Is(err, usecases.ErrRosterChanged):
		return huma.Error409Conflict(err.Error(), err)
	case errors.Is(err, usecases.ErrTournamentNotFound), errors.Is(err, usecases.ErrTeamLobbyNotFound):
		return huma.Error404NotFound("Requested tournament or team lobby was not found", err)
	case errors.Is(err, usecases.ErrPaymentUnavailable):
		return huma.Error503ServiceUnavailable(err.Error())
	case errors.Is(err, usecases.ErrInvalidRegistrationFee):
		return huma.Error409Conflict(err.Error())
	case errors.Is(err, usecases.ErrInvalidTeamName), errors.Is(err, usecases.ErrInvalidInviteCode),
		errors.Is(err, forms.ErrUnrecognizedField), errors.Is(err, forms.ErrDuplicateFieldAnswer),
		errors.Is(err, forms.ErrFieldRequired), errors.Is(err, forms.ErrInvalidFieldValue),
		errors.Is(err, forms.ErrInvalidFieldConfig), errors.Is(err, forms.ErrUnknownFieldType):
		return huma.Error400BadRequest("Invalid input", err)
	case errors.Is(err, usecases.ErrLobbyAccessDenied):
		return huma.Error403Forbidden("Only the team captain can perform this action", err)
	case errors.Is(err, usecases.ErrAlreadyInTournamentLobby):
		return huma.Error409Conflict("You are already in a team for this tournament", err)
	case errors.Is(err, usecases.ErrNotSoloTournament):
		return huma.Error409Conflict("This tournament requires a team and does not support solo registration", err)
	case errors.Is(err, usecases.ErrRegistrationNotOpen):
		return huma.Error409Conflict("Tournament registration is closed or the deadline has passed", err)
	case errors.Is(err, usecases.ErrTeamLobbyFull):
		return huma.Error409Conflict("Tournament or team lobby is full", err)
	case errors.Is(err, usecases.ErrRosterBelowMinimum):
		return huma.Error409Conflict("Team roster does not meet the minimum requirement", err)
	case errors.Is(err, usecases.ErrCannotRemoveCaptain):
		return huma.Error409Conflict("Cannot remove the team captain", err)
	case errors.Is(err, usecases.ErrCannotDisbandLobby):
		return huma.Error409Conflict("Cannot disband team lobby at this stage", err)
	case errors.Is(err, usecases.ErrTeamLobbyNotForming):
		return huma.Error409Conflict("Team lobby state does not allow this action", err)
	default:
		return huma.Error500InternalServerError("Team lobby operation failed", err)
	}
}

func toTeamLobbyResponse(view *usecases.TeamLobbyView, viewerID uuid.UUID) TeamLobbyResponse {
	team := view.Team
	members := make([]TeamLobbyMemberResponse, 0, len(view.Members))
	viewerRole := "INVITEE"
	var captainID uuid.UUID
	var captainName string
	for _, memberView := range view.Members {
		member := memberView.Member
		members = append(members, TeamLobbyMemberResponse{
			RegistrationComplete: memberView.RegistrationComplete,
			ID:                   member.ID, AccountID: member.AccountID,
			Handle: memberView.Handle, DisplayName: memberView.DisplayName,
			Role: string(member.Role), JoinedAt: member.JoinedAt,
		})
		if member.AccountID == viewerID {
			viewerRole = "MEMBER"
		}
		if member.Role == models.TournamentTeamMemberRoleCaptain {
			captainID = member.AccountID
			captainName = memberView.DisplayName
			if captainName == "" {
				captainName = memberView.Handle
			}
			if member.AccountID == viewerID {
				viewerRole = "CAPTAIN"
			}
		}
	}
	var paymentID *uuid.UUID
	var inviteCode *string
	if viewerRole == "CAPTAIN" {
		paymentID = team.PaymentID
	}
	if viewerRole != "INVITEE" {
		inviteCode = team.InviteCode
	}
	return TeamLobbyResponse{PaymentID: paymentID, PaymentExpiresAt: team.PaymentExpiresAt,
		ID: team.ID, Tournament: toTournamentResponse(team.Tournament, view.OrganizerName, 0),
		CaptainID: captainID, CaptainName: captainName, Name: team.Name,
		InviteCode: inviteCode, Status: string(team.Status), Members: members,
		ViewerRole: viewerRole, CreatedAt: team.CreatedAt,
	}
}
