package rest

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/moomaideng/eventory/internal/middlewares"
	"github.com/moomaideng/eventory/services/notification/internal/repositories"
	"github.com/moomaideng/eventory/services/notification/internal/usecases"
	"github.com/moomaideng/eventory/services/notification/models"
)

type NotificationItem struct {
	ID        uuid.UUID               `json:"id"`
	UserID    uuid.UUID               `json:"userId"`
	Type      models.NotificationType `json:"type"`
	Title     string                  `json:"title"`
	Message   string                  `json:"message"`
	Link      string                  `json:"link"`
	IsRead    bool                    `json:"isRead"`
	ReadAt    *time.Time              `json:"readAt"`
	CreatedAt time.Time               `json:"createdAt"`
}

type NotificationListResponse struct {
	Notifications []NotificationItem `json:"notifications"`
	UnreadCount   int64              `json:"unreadCount"`
}

type NotificationOutput struct {
	Body NotificationItem `json:"body"`
}

type NotificationListOutput struct {
	Body NotificationListResponse `json:"body"`
}

type NotificationCountOutput struct {
	Body struct {
		UpdatedCount int64 `json:"updatedCount"`
	} `json:"body"`
}

type NotificationIDInput struct {
	ID uuid.UUID `path:"id"`
}

func toItem(n *models.Notification) NotificationItem {
	return NotificationItem{
		ID:        n.ID,
		UserID:    n.UserID,
		Type:      n.Type,
		Title:     n.Title,
		Message:   n.Message,
		Link:      n.Link,
		IsRead:    n.IsRead,
		ReadAt:    n.ReadAt,
		CreatedAt: n.CreatedAt,
	}
}

func notificationError(err error) error {
	switch {
	case errors.Is(err, repositories.ErrNotFound):
		return huma.Error404NotFound("Notification not found")
	case errors.Is(err, repositories.ErrForbidden):
		return huma.Error403Forbidden("Access to notification is forbidden")
	default:
		return huma.Error500InternalServerError("Unable to process notification request", err)
	}
}

func RegisterNotificationRoutes(api huma.API, uc *usecases.InAppNotificationUseCase) {
	security := []map[string][]string{{"bearer": {}}}

	// List notifications for the authenticated user
	huma.Register(api, huma.Operation{
		OperationID: "list-my-notifications",
		Method:      http.MethodGet,
		Path:        "",
		Summary:     "List personal notifications ordered newest first",
		Security:    security,
	}, func(ctx context.Context, input *struct {
		Limit  int `query:"limit" default:"20" minimum:"1" maximum:"100"`
		Offset int `query:"offset" default:"0" minimum:"0"`
	}) (*NotificationListOutput, error) {
		userID, err := middlewares.GetAuthUserID(ctx)
		if err != nil {
			return nil, huma.Error401Unauthorized("Authentication required")
		}

		out, err := uc.ListNotifications(ctx, userID, input.Limit, input.Offset)
		if err != nil {
			return nil, notificationError(err)
		}

		items := make([]NotificationItem, 0, len(out.Notifications))
		for i := range out.Notifications {
			items = append(items, toItem(&out.Notifications[i]))
		}

		return &NotificationListOutput{
			Body: NotificationListResponse{
				Notifications: items,
				UnreadCount:   out.UnreadCount,
			},
		}, nil
	})

	// Mark a single notification as read
	huma.Register(api, huma.Operation{
		OperationID: "mark-notification-read",
		Method:      http.MethodPatch,
		Path:        "/{id}/read",
		Summary:     "Mark an owned notification as read",
		Security:    security,
	}, func(ctx context.Context, input *NotificationIDInput) (*NotificationOutput, error) {
		userID, err := middlewares.GetAuthUserID(ctx)
		if err != nil {
			return nil, huma.Error401Unauthorized("Authentication required")
		}

		n, err := uc.MarkAsRead(ctx, input.ID, userID)
		if err != nil {
			return nil, notificationError(err)
		}

		return &NotificationOutput{Body: toItem(n)}, nil
	})

	// Mark all notifications for the user as read
	huma.Register(api, huma.Operation{
		OperationID: "mark-all-notifications-read",
		Method:      http.MethodPatch,
		Path:        "/read-all",
		Summary:     "Mark all notifications for the user as read",
		Security:    security,
	}, func(ctx context.Context, _ *struct{}) (*NotificationCountOutput, error) {
		userID, err := middlewares.GetAuthUserID(ctx)
		if err != nil {
			return nil, huma.Error401Unauthorized("Authentication required")
		}

		count, err := uc.MarkAllAsRead(ctx, userID)
		if err != nil {
			return nil, notificationError(err)
		}

		out := &NotificationCountOutput{}
		out.Body.UpdatedCount = count
		return out, nil
	})
}
