package repository

import (
	"context"
	"errors"
	"fmt"

	"CalendarService/internal/domain/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GormDoctorRepository - реализация DoctorRepository с использованием GORM
type GormDoctorRepository struct {
	db *gorm.DB
}

// NewGormDoctorRepository создает новый репозиторий врачей с GORM
func NewGormDoctorRepository(db *gorm.DB) DoctorRepository {
	return &GormDoctorRepository{db: db}
}

func (r *GormDoctorRepository) GetDoctorByID(ctx context.Context, id uuid.UUID) (*models.Doctor, error) {
	var doctor models.Doctor
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&doctor).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("doctor not found: %w", err)
		}
		return nil, err
	}
	return &doctor, nil
}

func (r *GormDoctorRepository) GetDoctorByEmail(ctx context.Context, email string) (*models.Doctor, error) {
	var doctor models.Doctor
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&doctor).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("doctor not found: %w", err)
		}
		return nil, err
	}
	return &doctor, nil
}

// GetOrCreateDoctor ищет врача по ID, если не найден - создает нового
func (r *GormDoctorRepository) GetOrCreateDoctor(ctx context.Context, id uuid.UUID, email string) (*models.Doctor, error) {
	// Пробуем найти доктора
	doctor, err := r.GetDoctorByID(ctx, id)
	if err == nil {
		return doctor, nil
	}

	// Если не найден - создаём нового
	newDoctor := &models.Doctor{
		ID:    id,
		Email: email,
	}

	err = r.db.WithContext(ctx).Create(newDoctor).Error
	if err != nil {
		return nil, fmt.Errorf("failed to create doctor: %w", err)
	}

	return newDoctor, nil
}
