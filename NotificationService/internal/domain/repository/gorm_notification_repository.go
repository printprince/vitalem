package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"NotificationService/internal/domain/models"
)

// GormNotificationRepository - реализация NotificationRepository с использованием GORM
type GormNotificationRepository struct {
	db *gorm.DB
}

// NewGormNotificationRepository создает новый репозиторий уведомлений с GORM
func NewGormNotificationRepository(db *gorm.DB) NotificationRepository {
	return &GormNotificationRepository{db: db}
}

func (r *GormNotificationRepository) Create(ctx context.Context, n *models.Notification) error {
	return r.db.WithContext(ctx).Create(n).Error
}

func (r *GormNotificationRepository) GetByID(ctx context.Context, id int64) (*models.Notification, error) {
	var notification models.Notification
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&notification).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("notification not found")
		}
		return nil, err
	}
	return &notification, nil
}

func (r *GormNotificationRepository) ListByRecipient(ctx context.Context, recipientID uuid.UUID) ([]*models.Notification, error) {
	var notifications []*models.Notification
	err := r.db.WithContext(ctx).Where("recipient_id = ?", recipientID).Find(&notifications).Error
	return notifications, err
}

func (r *GormNotificationRepository) MarkAsSent(ctx context.Context, id int64) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":  models.StatusSent,
			"sent_at": &now,
		}).Error
}

func (r *GormNotificationRepository) UpdateStatus(ctx context.Context, n *models.Notification) error {
	return r.db.WithContext(ctx).Save(n).Error
}
