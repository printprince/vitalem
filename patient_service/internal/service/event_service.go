package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/printprince/vitalem/logger_service/pkg/logger"
	"github.com/printprince/vitalem/patient_service/internal/models"
)

// EventService интерфейс сервиса событий
type EventService interface {
	ProcessUserCreatedEvent(ctx context.Context, event models.UserCreatedEvent) error
}

// eventService реализация сервиса событий
type eventService struct {
	patientService PatientService
	logger         *logger.Client
}

// NewEventService создает новый сервис событий
func NewEventService(patientService PatientService, logger *logger.Client) EventService {
	return &eventService{
		patientService: patientService,
		logger:         logger,
	}
}

// ProcessUserCreatedEvent обрабатывает событие создания пользователя
func (s *eventService) ProcessUserCreatedEvent(ctx context.Context, event models.UserCreatedEvent) error {
	// Обрабатываем только пользователей с ролью "patient"
	if event.Role != "patient" {
		s.logger.Info("Пропускаем обработку события не-пациента", map[string]interface{}{
			"userID": event.UserID,
			"email":  event.Email,
			"role":   event.Role,
		})
		return nil
	}

	// Преобразуем строковый ID в UUID
	userID, err := uuid.Parse(event.UserID)
	if err != nil {
		s.logger.Error("Ошибка парсинга UserID", map[string]interface{}{
			"error":  err.Error(),
			"userID": event.UserID,
		})
		return err
	}

	// Создаем предварительный профиль пациента с временными данными
	// Пользователь заполнит полный профиль через API позже
	patient := &models.PatientCreateRequest{
		UserID:              userID,
		Name:                "Не указано",
		Surname:             "Не указана",
		DateOfBirth:         models.Date{Time: time.Now()},
		Gender:              "Не указан",
		Email:               event.Email,
		Phone:               "Не указан",
		IIN:                 nil, // NULL IIN для временного профиля, пользователь заполнит позже
		Height:              0,
		Weight:              0,
		PhysActivity:        models.ActivityInactive,
		Diagnoses:           []string{},
		AdditionalDiagnoses: []string{},
		Allergens:           []string{},
		AdditionalAllergens: []string{},
		Diet:                []string{},
		AdditionalDiets:     []string{},
	}

	// Создаем предварительный профиль
	response, err := s.patientService.CreatePatient(ctx, patient)
	if err != nil {
		s.logger.Error("Ошибка создания предварительного профиля пациента", map[string]interface{}{
			"error":  err.Error(),
			"userID": event.UserID,
			"email":  event.Email,
		})
		return err
	}

	s.logger.Info("Создан предварительный профиль пациента", map[string]interface{}{
		"userID":    event.UserID,
		"email":     event.Email,
		"patientID": response.ID,
	})

	return nil
}
