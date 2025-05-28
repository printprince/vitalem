package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/printprince/vitalem/logger_service/pkg/logger"
	"github.com/printprince/vitalem/patient_service/internal/models"
	"github.com/printprince/vitalem/patient_service/internal/repository"
)

type PatientService interface {
	CreatePatient(ctx context.Context, req *models.PatientCreateRequest) (*models.PatientResponse, error)
	GetPatientByID(ctx context.Context, id uuid.UUID) (*models.PatientResponse, error)
	GetPatientByUserID(ctx context.Context, userID uuid.UUID) (*models.PatientResponse, error)
	GetAllPatients(ctx context.Context) ([]*models.PatientResponse, error)
	UpdatePatient(ctx context.Context, id uuid.UUID, req *models.PatientCreateRequest) (*models.PatientResponse, error)
	UpdatePatientProfile(ctx context.Context, userID uuid.UUID, req *models.PatientCreateRequest) (*models.PatientResponse, error)
	DeletePatient(ctx context.Context, id uuid.UUID) error
}

type patientService struct {
	patientRepo repository.PatientRepository
	logger      *logger.Client
}

func NewPatientService(patientRepo repository.PatientRepository, logger *logger.Client) PatientService {
	return &patientService{
		patientRepo: patientRepo,
		logger:      logger,
	}
}

func (s *patientService) CreatePatient(ctx context.Context, req *models.PatientCreateRequest) (*models.PatientResponse, error) {
	patient := req.ToPatient()

	createdPatient, err := s.patientRepo.Create(ctx, patient)
	if err != nil {
		s.logger.Error("Failed to create patient", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, err
	}

	return createdPatient.ToPatientResponse(), nil
}

func (s *patientService) GetPatientByID(ctx context.Context, id uuid.UUID) (*models.PatientResponse, error) {
	patient, err := s.patientRepo.FindByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get patient by ID", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		})
		return nil, err
	}

	if patient == nil {
		return nil, nil
	}

	return patient.ToPatientResponse(), nil
}

func (s *patientService) GetPatientByUserID(ctx context.Context, userID uuid.UUID) (*models.PatientResponse, error) {
	patient, err := s.patientRepo.FindByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get patient by user ID", map[string]interface{}{
			"error":  err.Error(),
			"userID": userID,
		})
		return nil, err
	}

	if patient == nil {
		return nil, nil
	}

	return patient.ToPatientResponse(), nil
}

func (s *patientService) GetAllPatients(ctx context.Context) ([]*models.PatientResponse, error) {
	patients, err := s.patientRepo.FindAll(ctx)
	if err != nil {
		s.logger.Error("Failed to get all patients", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, err
	}

	var response []*models.PatientResponse
	for _, patient := range patients {
		response = append(response, patient.ToPatientResponse())
	}

	return response, nil
}

func (s *patientService) UpdatePatient(ctx context.Context, id uuid.UUID, req *models.PatientCreateRequest) (*models.PatientResponse, error) {
	patient, err := s.patientRepo.FindByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to find patient for update", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		})
		return nil, err
	}

	if patient == nil {
		return nil, nil
	}

	// Обновляем поля
	patient.Name = req.Name
	patient.Surname = req.Surname
	patient.DateOfBirth = req.DateOfBirth
	patient.Gender = req.Gender
	patient.Email = req.Email
	patient.Phone = req.Phone
	patient.Height = req.Height
	patient.Weight = req.Weight
	patient.PhysActivity = req.PhysActivity
	patient.Diagnoses = req.Diagnoses
	patient.AdditionalDiagnoses = req.AdditionalDiagnoses
	patient.Allergens = req.Allergens
	patient.AdditionalAllergens = req.AdditionalAllergens
	patient.Diet = req.Diet
	patient.AdditionalDiets = req.AdditionalDiets

	updatedPatient, err := s.patientRepo.Update(ctx, patient)
	if err != nil {
		s.logger.Error("Failed to update patient", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		})
		return nil, err
	}

	return updatedPatient.ToPatientResponse(), nil
}

func (s *patientService) UpdatePatientProfile(ctx context.Context, userID uuid.UUID, req *models.PatientCreateRequest) (*models.PatientResponse, error) {
	patient, err := s.patientRepo.FindByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to find patient by user ID for profile update", map[string]interface{}{
			"error":  err.Error(),
			"userID": userID,
		})
		return nil, err
	}

	if patient == nil {
		// Если профиль не найден, но у нас есть данные для создания, создаем новый
		s.logger.Info("Creating new patient profile for user", map[string]interface{}{
			"userID": userID,
		})

		newPatient := req.ToPatient()
		newPatient.UserID = userID // Устанавливаем ID пользователя

		createdPatient, err := s.patientRepo.Create(ctx, newPatient)
		if err != nil {
			s.logger.Error("Failed to create patient profile", map[string]interface{}{
				"error":  err.Error(),
				"userID": userID,
			})
			return nil, err
		}

		return createdPatient.ToPatientResponse(), nil
	}

	// Обновляем поля существующего профиля
	patient.Name = req.Name
	patient.Surname = req.Surname
	patient.DateOfBirth = req.DateOfBirth
	patient.Gender = req.Gender
	patient.Email = req.Email
	patient.Phone = req.Phone
	patient.Height = req.Height
	patient.Weight = req.Weight
	patient.PhysActivity = req.PhysActivity
	patient.Diagnoses = req.Diagnoses
	patient.AdditionalDiagnoses = req.AdditionalDiagnoses
	patient.Allergens = req.Allergens
	patient.AdditionalAllergens = req.AdditionalAllergens
	patient.Diet = req.Diet
	patient.AdditionalDiets = req.AdditionalDiets

	updatedPatient, err := s.patientRepo.Update(ctx, patient)
	if err != nil {
		s.logger.Error("Failed to update patient profile", map[string]interface{}{
			"error":  err.Error(),
			"userID": userID,
		})
		return nil, err
	}

	s.logger.Info("Patient profile updated successfully", map[string]interface{}{
		"userID":    userID,
		"patientID": patient.ID,
	})

	return updatedPatient.ToPatientResponse(), nil
}

func (s *patientService) DeletePatient(ctx context.Context, id uuid.UUID) error {
	err := s.patientRepo.Delete(ctx, id)
	if err != nil {
		s.logger.Error("Failed to delete patient", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		})
		return err
	}

	return nil
}
