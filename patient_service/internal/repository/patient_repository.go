package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/printprince/vitalem/logger_service/pkg/logger"
	"github.com/printprince/vitalem/patient_service/internal/models"
	"gorm.io/gorm"
)

// PatientRepository интерфейс репозитория пациентов
type PatientRepository interface {
	Create(ctx context.Context, patient *models.Patient) (*models.Patient, error)
	FindByID(ctx context.Context, id uuid.UUID) (*models.Patient, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) (*models.Patient, error)
	FindAll(ctx context.Context) ([]*models.Patient, error)
	Update(ctx context.Context, patient *models.Patient) (*models.Patient, error)
	Delete(ctx context.Context, id uuid.UUID) error
	InitDB() error
}

// patientRepository реализация репозитория пациентов
type patientRepository struct {
	db     *gorm.DB
	logger *logger.Client
}

// NewPatientRepository создает новый репозиторий пациентов
func NewPatientRepository(db *gorm.DB, logger *logger.Client) PatientRepository {
	return &patientRepository{
		db:     db,
		logger: logger,
	}
}

func (r *patientRepository) Create(ctx context.Context, patient *models.Patient) (*models.Patient, error) {
	if err := r.db.Create(patient).Error; err != nil {
		r.logger.Error("Failed to create patient", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, err
	}

	return patient, nil
}

func (r *patientRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Patient, error) {
	var patient models.Patient
	if err := r.db.Where("id = ?", id).First(&patient).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		r.logger.Error("Failed to find patient by ID", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		})
		return nil, err
	}

	return &patient, nil
}

func (r *patientRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*models.Patient, error) {
	var patient models.Patient
	if err := r.db.Where("user_id = ?", userID).First(&patient).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		r.logger.Error("Failed to find patient by user ID", map[string]interface{}{
			"error": err.Error(),
			"id":    userID,
		})
		return nil, err
	}

	return &patient, nil
}

func (r *patientRepository) FindAll(ctx context.Context) ([]*models.Patient, error) {
	var patients []*models.Patient
	if err := r.db.Find(&patients).Error; err != nil {
		r.logger.Error("Failed to find all patients", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, err
	}

	return patients, nil
}

func (r *patientRepository) Update(ctx context.Context, patient *models.Patient) (*models.Patient, error) {
	if err := r.db.Save(patient).Error; err != nil {
		r.logger.Error("Failed to update patient", map[string]interface{}{
			"error": err.Error(),
			"id":    patient.ID,
		})
		return nil, err
	}

	return patient, nil
}

func (r *patientRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.Delete(&models.Patient{}, id).Error; err != nil {
		r.logger.Error("Failed to delete patient", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		})
		return err
	}

	return nil
}

// InitDB инициализирует базу данных
func (r *patientRepository) InitDB() error {
	// Устанавливаем расширение uuid-ossp
	if err := r.db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
		return fmt.Errorf("failed to create uuid-ossp extension: %v", err)
	}

	// Автоматическая миграция моделей
	if err := r.db.AutoMigrate(&models.Patient{}); err != nil {
		return fmt.Errorf("ошибка миграции моделей: %v", err)
	}

	return nil
}
