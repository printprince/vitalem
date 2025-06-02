package repository

import (
	"context"

	"NotificationService/internal/domain/models"

	"github.com/google/uuid"
)

// NotificationRepository определяет методы для работы с уведомлениями
type NotificationRepository interface {
	Create(ctx context.Context, n *models.Notification) error
	GetByID(ctx context.Context, id int64) (*models.Notification, error)
	ListByRecipient(ctx context.Context, recipientID uuid.UUID) ([]*models.Notification, error)
	MarkAsSent(ctx context.Context, id int64) error
	UpdateStatus(ctx context.Context, n *models.Notification) error
}
