package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"CalendarService/internal/domain/models"
)

// GormEventRepository - реализация EventRepository с использованием GORM
type GormEventRepository struct {
	db *gorm.DB
}

// NewGormEventRepository создает новый репозиторий событий с GORM
func NewGormEventRepository(db *gorm.DB) EventRepository {
	return &GormEventRepository{db: db}
}

func (r *GormEventRepository) CreateEvent(ctx context.Context, event *models.Event) error {
	return r.db.WithContext(ctx).Create(event).Error
}

func (r *GormEventRepository) GetEventByID(ctx context.Context, id uuid.UUID) (*models.Event, error) {
	var event models.Event
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&event).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("event not found")
		}
		return nil, err
	}
	return &event, nil
}

func (r *GormEventRepository) GetEventsBySpecialist(ctx context.Context, specialistID uuid.UUID) ([]*models.Event, error) {
	var events []*models.Event
	err := r.db.WithContext(ctx).Where("specialist_id = ?", specialistID).Find(&events).Error
	return events, err
}

func (r *GormEventRepository) UpdateEvent(ctx context.Context, event *models.Event) error {
	return r.db.WithContext(ctx).Save(event).Error
}

func (r *GormEventRepository) DeleteEvent(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.Event{}, id).Error
}

// BookEvent бронирует событие (устанавливает patient_id и меняет статус)
func (r *GormEventRepository) BookEvent(ctx context.Context, eventID uuid.UUID, patientID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Model(&models.Event{}).
		Where("id = ? AND status = ?", eventID, "available").
		Updates(map[string]interface{}{
			"patient_id": patientID,
			"status":     "booked",
			"updated_at": time.Now(),
		})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("event is not available for booking")
	}
	return nil
}

// CancelEvent отменяет бронирование события (очищает patient_id и меняет статус)
func (r *GormEventRepository) CancelEvent(ctx context.Context, eventID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Model(&models.Event{}).
		Where("id = ? AND status = ?", eventID, "booked").
		Updates(map[string]interface{}{
			"patient_id": nil,
			"status":     "canceled",
			"updated_at": time.Now(),
		})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("event is not booked or already canceled")
	}
	return nil
}

func (r *GormEventRepository) GetAllEvents(ctx context.Context) ([]*models.Event, error) {
	var events []*models.Event
	err := r.db.WithContext(ctx).Find(&events).Error
	return events, err
}
