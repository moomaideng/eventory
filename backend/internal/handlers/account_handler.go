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
	"github.com/moomaideng/eventory/internal/usecases"
)

// AccountResponse represents the public payload for an Account entity.
type AccountResponse struct {
	ID          uuid.UUID `json:"id" doc:"Unique account UUID matching Supabase Auth UID"`
	Email       string    `json:"email" doc:"User email address verified by Auth Provider"`
	Handle      string    `json:"handle" doc:"Unique public handle (e.g. @moomai_01)"`
	DisplayName string    `json:"displayName" doc:"User display name"`
	AvatarURL   *string   `json:"avatarUrl,omitempty" doc:"User avatar image URL"`
	Phone       *string   `json:"phone,omitempty" doc:"Contact phone number"`
	Status      string    `json:"status" doc:"Account status (ACTIVE, SUSPENDED)"`
	CreatedAt   time.Time `json:"createdAt" doc:"Timestamp of account creation"`
}

// AccountOutput represents the standard HTTP response returning an Account.
type AccountOutput struct {
	Body AccountResponse
}

// CreateAccountOutput represents the HTTP response for creating or ensuring an account.
type CreateAccountOutput struct {
	Status int `default:"201"`
	Body   AccountResponse
}

// GetAccountByIDInput represents path parameters for retrieving an account by ID.
type GetAccountByIDInput struct {
	ID uuid.UUID `path:"id" doc:"Account UUID"`
}

// CreateAccountRequest defines the request body for creating or ensuring an account.
type CreateAccountRequest struct {
	DisplayName string  `json:"displayName,omitempty" doc:"Chosen display name" maxLength:"64"`
	Handle      *string `json:"handle,omitempty" doc:"Optional requested unique handle (3-32 chars)" maxLength:"32"`
	AvatarURL   *string `json:"avatarUrl,omitempty" doc:"Optional avatar image URL"`
}

// CreateAccountInput represents request payload for creating or ensuring an account.
type CreateAccountInput struct {
	Body CreateAccountRequest
}

// UpdateAccountRequest defines the request body for updating user profile fields.
type UpdateAccountRequest struct {
	DisplayName *string `json:"displayName,omitempty" doc:"New display name" minLength:"1" maxLength:"64"`
	Handle      *string `json:"handle,omitempty" doc:"New unique handle" minLength:"3" maxLength:"32"`
	Phone       *string `json:"phone,omitempty" doc:"New contact phone"`
	AvatarURL   *string `json:"avatarUrl,omitempty" doc:"New avatar URL"`
}

// UpdateAccountInput represents request payload for updating profile fields.
type UpdateAccountInput struct {
	Body UpdateAccountRequest
}

func toAccountResponse(acc *models.Account) AccountResponse {
	return AccountResponse{
		ID:          acc.ID,
		Email:       acc.Email,
		Handle:      acc.Handle,
		DisplayName: acc.DisplayName,
		AvatarURL:   acc.AvatarURL,
		Phone:       acc.Phone,
		Status:      acc.Status,
		CreatedAt:   acc.CreatedAt,
	}
}

// getAuthAccount resolves the authenticated account from JWT sub or email.
func getAuthAccount(ctx context.Context, uc *usecases.AccountUseCase) (*models.Account, error) {
	if userID, err := middlewares.GetAuthUserID(ctx); err == nil {
		if acc, _ := uc.GetAccountByID(ctx, userID); acc != nil {
			return acc, nil
		}
	}

	email, err := middlewares.GetAuthEmail(ctx)
	if err != nil {
		return nil, huma.Error401Unauthorized("Authentication required", err)
	}

	acc, err := uc.GetAccountByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, usecases.ErrAccountNotFound) {
			return nil, huma.Error404NotFound("Account not found", err)
		}
		return nil, huma.Error500InternalServerError("Failed to query account", err)
	}
	return acc, nil
}

// RegisterAccountRoutes registers all account-related HTTP endpoints under the account group.
func RegisterAccountRoutes(api huma.API, accountUseCase *usecases.AccountUseCase) {
	// 1. GET /api/v1/accounts/me (Get Authenticated User's Account)
	huma.Register(api, huma.Operation{
		OperationID: "get-my-account",
		Method:      http.MethodGet,
		Path:        "/me",
		Summary:     "Get Current User Account",
		Description: "Retrieves the account of the authenticated user via Bearer JWT.",
		Tags:        []string{"Accounts"},
		Security:    []map[string][]string{{"bearer": {}}},
	}, func(ctx context.Context, input *struct{}) (*AccountOutput, error) {
		acc, err := getAuthAccount(ctx, accountUseCase)
		if err != nil {
			return nil, err
		}
		return &AccountOutput{Body: toAccountResponse(acc)}, nil
	})

	// 2. POST /api/v1/accounts (Create or Ensure Account - Idempotent JIT Provisioning)
	huma.Register(api, huma.Operation{
		OperationID:   "create-account",
		Method:        http.MethodPost,
		Path:          "",
		Summary:       "Create Account",
		Description:   "Creates or ensures an account exists for the authenticated user using JWT sub and email.",
		DefaultStatus: http.StatusCreated,
		Tags:          []string{"Accounts"},
		Security:      []map[string][]string{{"bearer": {}}},
	}, func(ctx context.Context, input *CreateAccountInput) (*CreateAccountOutput, error) {
		userID, err := middlewares.GetAuthUserID(ctx)
		if err != nil {
			return nil, huma.Error401Unauthorized("Authentication required: invalid or missing user id in token", err)
		}

		email, err := middlewares.GetAuthEmail(ctx)
		if err != nil {
			return nil, huma.Error401Unauthorized("Authentication required: missing email in token", err)
		}

		var handle string
		if input.Body.Handle != nil {
			handle = *input.Body.Handle
		}

		acc, err := accountUseCase.CreateAccount(ctx, usecases.CreateAccountInput{
			ID:          userID,
			Email:       email,
			DisplayName: input.Body.DisplayName,
			Handle:      handle,
			AvatarURL:   input.Body.AvatarURL,
		})
		if err != nil {
			if errors.Is(err, usecases.ErrInvalidAccountID) {
				return nil, huma.Error400BadRequest("Valid account ID is required", err)
			}
			return nil, huma.Error500InternalServerError("Failed to create account", err)
		}

		return &CreateAccountOutput{
			Status: http.StatusCreated,
			Body:   toAccountResponse(acc),
		}, nil
	})

	// 3. PATCH /api/v1/accounts/me (Update Current User's Profile)
	huma.Register(api, huma.Operation{
		OperationID: "update-my-account",
		Method:      http.MethodPatch,
		Path:        "/me",
		Summary:     "Update Current User Account",
		Description: "Updates profile fields (displayName, handle, phone, avatarUrl) of the authenticated user.",
		Tags:        []string{"Accounts"},
		Security:    []map[string][]string{{"bearer": {}}},
	}, func(ctx context.Context, input *UpdateAccountInput) (*AccountOutput, error) {
		acc, err := getAuthAccount(ctx, accountUseCase)
		if err != nil {
			return nil, err
		}

		updated, err := accountUseCase.UpdateAccount(ctx, acc.ID, usecases.UpdateAccountInput{
			DisplayName: input.Body.DisplayName,
			Handle:      input.Body.Handle,
			Phone:       input.Body.Phone,
			AvatarURL:   input.Body.AvatarURL,
		})
		if err != nil {
			if errors.Is(err, usecases.ErrHandleAlreadyExists) {
				return nil, huma.Error409Conflict("Handle is already taken. Please choose another one.", err)
			}
			if errors.Is(err, usecases.ErrInvalidHandle) {
				return nil, huma.Error400BadRequest("Invalid handle provided", err)
			}
			if errors.Is(err, usecases.ErrInvalidDisplayName) {
				return nil, huma.Error400BadRequest("Display name cannot be empty", err)
			}
			return nil, huma.Error500InternalServerError("Failed to update account", err)
		}

		return &AccountOutput{Body: toAccountResponse(updated)}, nil
	})

	// 4. GET /api/v1/accounts/{id} (Lookup Account by ID)
	huma.Register(api, huma.Operation{
		OperationID: "get-account-by-id",
		Method:      http.MethodGet,
		Path:        "/{id}",
		Summary:     "Get Account by ID",
		Description: "Retrieves public account details for a given internal user UUID.",
		Tags:        []string{"Accounts"},
		Security:    []map[string][]string{{"bearer": {}}},
	}, func(ctx context.Context, input *GetAccountByIDInput) (*AccountOutput, error) {
		acc, err := accountUseCase.GetAccountByID(ctx, input.ID)
		if err != nil {
			if errors.Is(err, usecases.ErrAccountNotFound) {
				return nil, huma.Error404NotFound("Account not found", err)
			}
			return nil, huma.Error500InternalServerError("Failed to retrieve account", err)
		}
		return &AccountOutput{Body: toAccountResponse(acc)}, nil
	})
}
