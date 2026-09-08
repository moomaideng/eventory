package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/moomaideng/eventory/internal/middlewares"
	"github.com/moomaideng/eventory/internal/models"
	"github.com/moomaideng/eventory/internal/repositories"
	"github.com/moomaideng/eventory/internal/usecases"
)

type TeamLobbyMemberResponse struct {
	ID          uuid.UUID `json:"id"`
	AccountID   uuid.UUID `json:"accountId"`
	Handle      string    `json:"handle"`
	DisplayName string    `json:"displayName"`
	Role        string    `json:"role" enum:"CAPTAIN,MEMBER"`
	JoinedAt    time.Time `json:"joinedAt"`
}

type TeamLobbyResponse struct {
	ID          uuid.UUID                 `json:"id"`
	Tournament  TournamentResponse        `json:"tournament"`
	CaptainID   uuid.UUID                 `json:"captainId"`
	CaptainName string                    `json:"captainName"`
	Name        string                    `json:"name"`
	InviteCode  string                    `json:"inviteCode"`
	Status      string                    `json:"status" enum:"FORMING,LOCKED,ACCEPTED,REJECTED"`
	Members     []TeamLobbyMemberResponse `json:"members"`
	ViewerRole  string                    `json:"viewerRole" enum:"CAPTAIN,MEMBER,INVITEE"`
	CreatedAt   time.Time                 `json:"createdAt"`
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
	Name string `json:"name" minLength:"1" maxLength:"120" doc:"Team name"`
}

type CreateTeamLobbyInput struct {
	TournamentID uuid.UUID `path:"tournamentId" doc:"Tournament UUID"`
	Body         CreateTeamLobbyRequest
}

type MyTournamentTeamInput struct {
	TournamentID uuid.UUID `path:"tournamentId" doc:"Tournament UUID"`
}

type InviteCodeInput struct {
	InviteCode string `path:"inviteCode" pattern:"^[A-Za-z0-9]{6}$" doc:"Six-character team invite code"`
}

type LobbyIDInput struct {
	ID uuid.UUID `path:"id" doc:"Team UUID"`
}

type LobbyMemberInput struct {
	ID       uuid.UUID `path:"id" doc:"Team UUID"`
	MemberID uuid.UUID `path:"memberId" doc:"Team member UUID"`
}

func RegisterTeamLobbyRoutes(api huma.API, lobbyUseCase *usecases.TeamLobbyUseCase, accountUseCase *usecases.AccountUseCase) {
	huma.Register(api, huma.Operation{
		OperationID: "get-my-tournament-team",
		Method:      http.MethodGet,
		Path:        "/tournaments/{tournamentId}/my-team",
		Summary:     "View the authenticated competitor's team for a tournament",
		Tags:        []string{"Team lobbies"},
		Security:    []map[string][]string{{"bearer": {}}},
	}, func(ctx context.Context, input *MyTournamentTeamInput) (*TeamLobbyOutput, error) {
		account, err := authenticatedAccount(ctx, accountUseCase)
		if err != nil {
			return nil, err
		}
		team, err := lobbyUseCase.GetActiveForTournament(ctx, input.TournamentID, account.ID)
		if err != nil {
			return nil, teamLobbyHTTPError(err)
		}
		return &TeamLobbyOutput{Body: toTeamLobbyResponse(team, account.ID)}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "create-team-lobby",
		Method:      http.MethodPost,
		Path:        "/tournaments/{tournamentId}/lobbies",
		Summary:     "Create a team lobby",
		Description: "Creates a forming team and makes the authenticated account its captain.",
		Tags:        []string{"Team lobbies"},
		Security:    []map[string][]string{{"bearer": {}}},
	}, func(ctx context.Context, input *CreateTeamLobbyInput) (*CreateTeamLobbyOutput, error) {
		account, err := authenticatedAccount(ctx, accountUseCase)
		if err != nil {
			return nil, err
		}
		team, err := lobbyUseCase.Create(ctx, input.TournamentID, account.ID, input.Body.Name)
		if err != nil {
			return nil, teamLobbyHTTPError(err)
		}
		return &CreateTeamLobbyOutput{Status: http.StatusCreated, Body: toTeamLobbyResponse(team, account.ID)}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "get-team-lobby",
		Method:      http.MethodGet,
		Path:        "/lobbies/{inviteCode}",
		Summary:     "View a team lobby by invite code",
		Tags:        []string{"Team lobbies"},
		Security:    []map[string][]string{{"bearer": {}}},
	}, func(ctx context.Context, input *InviteCodeInput) (*TeamLobbyOutput, error) {
		account, err := authenticatedAccount(ctx, accountUseCase)
		if err != nil {
			return nil, err
		}
		team, err := lobbyUseCase.GetByInviteCode(ctx, input.InviteCode)
		if err != nil {
			return nil, teamLobbyHTTPError(err)
		}
		return &TeamLobbyOutput{Body: toTeamLobbyResponse(team, account.ID)}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "join-team-lobby",
		Method:      http.MethodPost,
		Path:        "/lobbies/{inviteCode}/join",
		Summary:     "Join a team lobby",
		Tags:        []string{"Team lobbies"},
		Security:    []map[string][]string{{"bearer": {}}},
	}, func(ctx context.Context, input *InviteCodeInput) (*TeamLobbyOutput, error) {
		account, err := authenticatedAccount(ctx, accountUseCase)
		if err != nil {
			return nil, err
		}
		team, err := lobbyUseCase.Join(ctx, input.InviteCode, account.ID)
		if err != nil {
			return nil, teamLobbyHTTPError(err)
		}
		return &TeamLobbyOutput{Body: toTeamLobbyResponse(team, account.ID)}, nil
	})

	registerCaptainAction(api, lobbyUseCase, accountUseCase, "regenerate-team-lobby-invite", "/lobbies/{id}/invite/regenerate", "Regenerate a team invite", func(ctx context.Context, teamID, accountID uuid.UUID) (*models.TournamentTeam, error) {
		return lobbyUseCase.RegenerateInvite(ctx, teamID, accountID)
	})
	registerCaptainAction(api, lobbyUseCase, accountUseCase, "lock-team-lobby", "/lobbies/{id}/lock", "Lock a team roster", func(ctx context.Context, teamID, accountID uuid.UUID) (*models.TournamentTeam, error) {
		return lobbyUseCase.Lock(ctx, teamID, accountID)
	})

	huma.Register(api, huma.Operation{
		OperationID: "remove-team-lobby-member",
		Method:      http.MethodDelete,
		Path:        "/lobbies/{id}/members/{memberId}",
		Summary:     "Remove a lobby member",
		Tags:        []string{"Team lobbies"},
		Security:    []map[string][]string{{"bearer": {}}},
	}, func(ctx context.Context, input *LobbyMemberInput) (*TeamLobbyOutput, error) {
		account, err := authenticatedAccount(ctx, accountUseCase)
		if err != nil {
			return nil, err
		}
		team, err := lobbyUseCase.RemoveMember(ctx, input.ID, input.MemberID, account.ID)
		if err != nil {
			return nil, teamLobbyHTTPError(err)
		}
		return &TeamLobbyOutput{Body: toTeamLobbyResponse(team, account.ID)}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "disband-team-lobby",
		Method:      http.MethodDelete,
		Path:        "/lobbies/{id}",
		Summary:     "Disband a team lobby",
		Tags:        []string{"Team lobbies"},
		Security:    []map[string][]string{{"bearer": {}}},
	}, func(ctx context.Context, input *LobbyIDInput) (*EmptyTeamLobbyOutput, error) {
		account, err := authenticatedAccount(ctx, accountUseCase)
		if err != nil {
			return nil, err
		}
		if err := lobbyUseCase.Disband(ctx, input.ID, account.ID); err != nil {
			return nil, teamLobbyHTTPError(err)
		}
		return &EmptyTeamLobbyOutput{Status: http.StatusNoContent}, nil
	})
}

type captainAction func(context.Context, uuid.UUID, uuid.UUID) (*models.TournamentTeam, error)

func registerCaptainAction(api huma.API, lobbyUseCase *usecases.TeamLobbyUseCase, accountUseCase *usecases.AccountUseCase, operationID, path, summary string, action captainAction) {
	huma.Register(api, huma.Operation{
		OperationID: operationID,
		Method:      http.MethodPost,
		Path:        path,
		Summary:     summary,
		Tags:        []string{"Team lobbies"},
		Security:    []map[string][]string{{"bearer": {}}},
	}, func(ctx context.Context, input *LobbyIDInput) (*TeamLobbyOutput, error) {
		account, err := authenticatedAccount(ctx, accountUseCase)
		if err != nil {
			return nil, err
		}
		team, err := action(ctx, input.ID, account.ID)
		if err != nil {
			return nil, teamLobbyHTTPError(err)
		}
		return &TeamLobbyOutput{Body: toTeamLobbyResponse(team, account.ID)}, nil
	})
}

func authenticatedAccount(ctx context.Context, accountUseCase *usecases.AccountUseCase) (*models.Account, error) {
	if userID, err := middlewares.GetAuthUserID(ctx); err == nil {
		if acc, _ := accountUseCase.GetAccountByID(ctx, userID); acc != nil {
			return acc, nil
		}
	}

	email, err := middlewares.GetAuthEmail(ctx)
	if err != nil {
		return nil, huma.Error401Unauthorized("Authentication required", err)
	}
	account, err := accountUseCase.GetAccountByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, usecases.ErrAccountNotFound) {
			return nil, huma.Error404NotFound("Account not found.", err)
		}
		return nil, huma.Error500InternalServerError("Failed to retrieve account", err)
	}
	return account, nil
}

func teamLobbyHTTPError(err error) error {
	switch {
	case errors.Is(err, usecases.ErrTournamentNotFound), errors.Is(err, repositories.ErrTeamLobbyNotFound):
		return huma.Error404NotFound("Requested tournament or team lobby was not found", err)
	case errors.Is(err, usecases.ErrInvalidTeamName):
		return huma.Error400BadRequest("Invalid team name", err)
	case errors.Is(err, usecases.ErrLobbyAccessDenied):
		return huma.Error403Forbidden("Only the team captain can perform this action", err)
	case errors.Is(err, repositories.ErrAlreadyInTournamentLobby):
		return huma.Error409Conflict("You are already in a team for this tournament", err)
	case errors.Is(err, usecases.ErrTeamsNotAllowed),
		errors.Is(err, usecases.ErrRegistrationNotOpen),
		errors.Is(err, repositories.ErrTeamLobbyNotForming),
		errors.Is(err, repositories.ErrTeamLobbyFull),
		errors.Is(err, repositories.ErrRosterBelowMinimum),
		errors.Is(err, repositories.ErrTournamentRegistrationClosed),
		errors.Is(err, repositories.ErrCannotRemoveCaptain),
		errors.Is(err, repositories.ErrCannotDisbandLobby):
		return huma.Error409Conflict("Team lobby state does not allow this action", err)
	default:
		return huma.Error500InternalServerError("Team lobby operation failed", err)
	}
}

func toTeamLobbyResponse(team *models.TournamentTeam, viewerID uuid.UUID) TeamLobbyResponse {
	members := make([]TeamLobbyMemberResponse, 0, len(team.Members))
	viewerRole := "INVITEE"
	var captainID uuid.UUID
	var captainName string
	for _, member := range team.Members {
		members = append(members, TeamLobbyMemberResponse{
			ID: member.ID, AccountID: member.AccountID,
			Handle: member.Account.Handle, DisplayName: member.Account.DisplayName,
			Role: string(member.Role), JoinedAt: member.JoinedAt,
		})
		if member.AccountID == viewerID {
			viewerRole = "MEMBER"
		}
		if member.Role == models.TournamentTeamMemberRoleCaptain {
			captainID = member.AccountID
			captainName = member.Account.DisplayName
			if captainName == "" {
				captainName = member.Account.Handle
			}
			if member.AccountID == viewerID {
				viewerRole = "CAPTAIN"
			}
		}
	}
	return TeamLobbyResponse{
		ID: team.ID, Tournament: toTournamentResponse(team.Tournament, 0),
		CaptainID: captainID, CaptainName: captainName, Name: team.Name,
		InviteCode: team.InviteCode, Status: string(team.Status), Members: members,
		ViewerRole: viewerRole, CreatedAt: team.CreatedAt,
	}
}
