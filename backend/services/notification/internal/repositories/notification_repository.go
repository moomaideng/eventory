package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/notification/models"
	"gorm.io/gorm"
)

var (
	ErrNotFound  = errors.New("notification not found")
	ErrForbidden = errors.New("access to notification is forbidden")
)

type NotificationRepository interface {
	Create(ctx context.Context, n *models.Notification) (*models.Notification, error)
	ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.Notification, int64, error)
	GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.Notification, error)
	MarkAsRead(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.Notification, error)
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) (int64, error)
}

type notificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Create(ctx context.Context, n *models.Notification) (*models.Notification, error) {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	now := time.Now().UTC()
	if n.CreatedAt.IsZero() {
		n.CreatedAt = now
	}
	n.UpdatedAt = now

	err := r.db.WithContext(ctx).Create(n).Error
	if err != nil {
		return nil, err
	}
	return n, nil
}

func (r *notificationRepository) ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.Notification, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	var notifications []models.Notification
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&notifications).Error
	if err != nil {
		return nil, 0, err
	}

	var unreadCount int64
	err = r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&unreadCount).Error
	if err != nil {
		return nil, 0, err
	}

	return notifications, unreadCount, nil
}

func (r *notificationRepository) GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.Notification, error) {
	var n models.Notification
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&n).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	if n.UserID != userID {
		return nil, ErrForbidden
	}

	return &n, nil
}

func (r *notificationRepository) MarkAsRead(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.Notification, error) {
	n, err := r.GetByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	if n.IsRead {
		return n, nil
	}

	now := time.Now().UTC()
	err = r.db.WithContext(ctx).
		Model(n).
		Updates(map[string]any{
			"is_read":    true,
			"read_at":    now,
			"updated_at": now,
		}).Error
	if err != nil {
		return nil, err
	}

	n.IsRead = true
	n.ReadAt = &now
	n.UpdatedAt = now
	return n, nil
}

func (r *notificationRepository) MarkAllAsRead(ctx context.Context, userID uuid.UUID) (int64, error) {
	now := time.Now().UTC()
	res := r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Updates(map[string]any{
			"is_read":    true,
			"read_at":    now,
			"updated_at": now,
		})
	return res.RowsAffected, res.Error
}
