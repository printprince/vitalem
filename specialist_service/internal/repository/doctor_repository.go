package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/printprince/vitalem/logger_service/pkg/logger"
	"github.com/printprince/vitalem/specialist_service/internal/models"
	"gorm.io/gorm"
)

// DoctorRepository интерфейс репозитория врачей
type DoctorRepository interface {
	Create(ctx context.Context, doctor *models.Doctor) (*models.Doctor, error)
	FindByID(ctx context.Context, id uuid.UUID) (*models.Doctor, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) (*models.Doctor, error)
	FindAll(ctx context.Context) ([]*models.Doctor, error)
	Update(ctx context.Context, doctor *models.Doctor) (*models.Doctor, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// doctorRepository реализация репозитория врачей
type doctorRepository struct {
	db     *gorm.DB
	logger *logger.Client
}

// NewDoctorRepository создает новый репозиторий врачей
func NewDoctorRepository(db *gorm.DB, logger *logger.Client) DoctorRepository {
	return &doctorRepository{
		db:     db,
		logger: logger,
	}
}

func (r *doctorRepository) Create(ctx context.Context, doctor *models.Doctor) (*models.Doctor, error) {
	if err := r.db.Create(doctor).Error; err != nil {
		r.logger.Error("Failed to create doctor", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, err
	}

	return doctor, nil

}

func (r *doctorRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Doctor, error) {
	var doctor models.Doctor
	if err := r.db.Where("id = ?", id).First(&doctor).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		r.logger.Error("Failed to find doctor by ID", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		})
		return nil, err
	}

	return &doctor, nil

}

func (r *doctorRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*models.Doctor, error) {
	var doctor models.Doctor
	if err := r.db.Where("user_id = ?", userID).First(&doctor).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		r.logger.Error("Failed to find doctor by user ID", map[string]interface{}{
			"error": err.Error(),
			"id":    userID,
		})
		return nil, err
	}

	return &doctor, nil

}

func (r *doctorRepository) FindAll(ctx context.Context) ([]*models.Doctor, error) {
	var doctors []*models.Doctor
	if err := r.db.Find(&doctors).Error; err != nil {
		r.logger.Error("Failed to find all doctors", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, err
	}

	return doctors, nil

}

func (r *doctorRepository) Update(ctx context.Context, doctor *models.Doctor) (*models.Doctor, error) {
	if err := r.db.Save(doctor).Error; err != nil {
		r.logger.Error("Failed to update doctor", map[string]interface{}{
			"error": err.Error(),
			"id":    doctor.ID,
		})
		return nil, err
	}

	return doctor, nil

}

func (r *doctorRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.Delete(&models.Doctor{}, id).Error; err != nil {
		r.logger.Error("Failed to delete doctor", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		})
		return err
	}

	return nil
}
