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

// UpdatePatient - метод для обновления данных существующего пациента
// Обновляет все поля по ID пациента (не частичное обновление)
// Если пациент не найден - возвращает nil без ошибки
// Подходит для полного обновления профиля через админку
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

	// Обновляем все поля пациента из запроса
	// Жёсткий перезапись всех полей без проверки на nil
	// TODO: Добавить частичное обновление с проверкой заполненности полей
	patient.Name = req.Name
	patient.IIN = req.IIN
	patient.Surname = req.Surname
	patient.DateOfBirth = req.DateOfBirth.Time
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

	// Сохраняем обновленные данные в базу
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

// UpdatePatientProfile - обновление или создание профиля пациента по UserID
// Умная логика: если профиль существует - обновляет, если нет - создает новый
// Используется для второго этапа регистрации, когда юзер уже создан в identity_service,
// но еще не заполнил свой медицинский профиль
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
		// Если профиль не найден - создаем новый на лету
		// Это удобно для фронта - единый эндпоинт для создания и обновления
		s.logger.Info("Creating new patient profile for user", map[string]interface{}{
			"userID": userID,
		})

		// Конвертим DTO в модель и устанавливаем правильный UserID
		newPatient := req.ToPatient()
		newPatient.UserID = userID // Перезаписываем ID из пути, это критично для безопасности

		// Сохраняем нового пациента в базу
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
	// Обновляем только НЕ пустые поля из запроса (гибкое обновление)
	if req.Name != "" {
		patient.Name = req.Name
	}
	if req.Surname != "" {
		patient.Surname = req.Surname
	}
	if !req.DateOfBirth.IsZero() {
		patient.DateOfBirth = req.DateOfBirth.Time
	}
	if req.Gender != "" {
		patient.Gender = req.Gender
	}
	if req.Email != "" {
		patient.Email = req.Email
	}
	if req.Phone != "" {
		patient.Phone = req.Phone
	}
	if req.Height != 0 {
		patient.Height = req.Height
	}
	if req.Weight != 0 {
		patient.Weight = req.Weight
	}
	if req.PhysActivity != "" {
		patient.PhysActivity = req.PhysActivity
	}
	if req.IIN != nil && *req.IIN != "" {
		patient.IIN = req.IIN
	}
	if len(req.Diagnoses) > 0 {
		patient.Diagnoses = req.Diagnoses
	}
	if len(req.AdditionalDiagnoses) > 0 {
		patient.AdditionalDiagnoses = req.AdditionalDiagnoses
	}
	if len(req.Allergens) > 0 {
		patient.Allergens = req.Allergens
	}
	if len(req.AdditionalAllergens) > 0 {
		patient.AdditionalAllergens = req.AdditionalAllergens
	}
	if len(req.Diet) > 0 {
		patient.Diet = req.Diet
	}
	if len(req.AdditionalDiets) > 0 {
		patient.AdditionalDiets = req.AdditionalDiets
	}

	// Сохраняем изменения
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
