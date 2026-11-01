package usecases

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/notification/internal/repositories"
	"github.com/moomaideng/eventory/services/notification/models"
)

type InAppNotificationUseCase struct {
	repo repositories.NotificationRepository
}

func NewInAppNotificationUseCase(repo repositories.NotificationRepository) *InAppNotificationUseCase {
	return &InAppNotificationUseCase{repo: repo}
}

type NotificationListOutput struct {
	Notifications []models.Notification `json:"notifications"`
	UnreadCount   int64                 `json:"unread_count"`
}

func (uc *InAppNotificationUseCase) ListNotifications(ctx context.Context, userID uuid.UUID, limit, offset int) (*NotificationListOutput, error) {
	items, unread, err := uc.repo.ListByUser(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list notifications: %w", err)
	}

	if items == nil {
		items = []models.Notification{}
	}

	return &NotificationListOutput{
		Notifications: items,
		UnreadCount:   unread,
	}, nil
}

func (uc *InAppNotificationUseCase) MarkAsRead(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.Notification, error) {
	return uc.repo.MarkAsRead(ctx, id, userID)
}

func (uc *InAppNotificationUseCase) MarkAllAsRead(ctx context.Context, userID uuid.UUID) (int64, error) {
	return uc.repo.MarkAllAsRead(ctx, userID)
}

func (uc *InAppNotificationUseCase) CreateNotification(ctx context.Context, n *models.Notification) (*models.Notification, error) {
	return uc.repo.Create(ctx, n)
}
