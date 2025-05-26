package service

import (
	"context"

	"github.com/printprince/vitalem/logger_service/pkg/logger"
	"github.com/printprince/vitalem/specialist_service/internal/models"
)

// EventService интерфейс сервиса событий
type EventService interface {
	ProcessUserCreatedEvent(ctx context.Context, event models.UserCreatedEvent) error
}

// eventService реализация сервиса событий
type eventService struct {
	doctorService DoctorService
	logger        *logger.Client
}

// NewEventService создает новый сервис событий
func NewEventService(doctorService DoctorService, logger *logger.Client) EventService {
	return &eventService{
		doctorService: doctorService,
		logger:        logger,
	}
}

// ProcessUserCreatedEvent обрабатывает событие создания пользователя
func (s *eventService) ProcessUserCreatedEvent(ctx context.Context, event models.UserCreatedEvent) error {
	// Обрабатываем только пользователей с ролью "doctor"
	if event.Role != "doctor" {
		s.logger.Info("Пропускаем обработку события не-доктора", map[string]interface{}{
			"userID": event.UserID,
			"email":  event.Email,
			"role":   event.Role,
		})
		return nil
	}

	// Создаем предварительный профиль доктора с временными данными
	// Пользователь заполнит полный профиль через API позже
	doctor := &models.DoctorCreateRequest{
		UserID:      event.UserID,
		FirstName:   "Не указано",
		LastName:    "Не указана",
		Email:       event.Email,
		Phone:       "Не указан",
		Description: "Профиль создан автоматически и требует заполнения",
		Roles:       []string{models.RoleNotSpecified},
	}

	// Создаем предварительный профиль
	response, err := s.doctorService.CreateDoctor(ctx, doctor)
	if err != nil {
		s.logger.Error("Ошибка создания предварительного профиля врача", map[string]interface{}{
			"error":  err.Error(),
			"userID": event.UserID,
			"email":  event.Email,
		})
		return err
	}

	s.logger.Info("Создан предварительный профиль врача", map[string]interface{}{
		"userID":   event.UserID,
		"email":    event.Email,
		"doctorID": response.ID,
	})

	return nil
}
