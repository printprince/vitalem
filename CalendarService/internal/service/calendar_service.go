package service

import (
	"CalendarService/pkg/logger"
	"context"
	"errors"
	"time"

	"CalendarService/internal/domain/models"
	"CalendarService/internal/domain/repository"
	"CalendarService/internal/infrastructure/notification"

	"github.com/google/uuid"
)

type CalendarService struct {
	repo           repository.EventRepository
	DoctorRepo     repository.DoctorRepository
	notificationCl *notification.Client
	logger         *logger.Logger
}

func NewCalendarService(repo repository.EventRepository, DoctorRepo repository.DoctorRepository, notif *notification.Client, logger *logger.Logger) *CalendarService {
	return &CalendarService{
		repo:           repo,
		DoctorRepo:     DoctorRepo,
		notificationCl: notif,
		logger:         logger,
	}
}

// Создание нового события
func (s *CalendarService) CreateEvent(ctx context.Context, event *models.Event) error {
	now := time.Now()
	event.CreatedAt = now
	event.UpdatedAt = now
	event.Status = "available" // Предполагаем, что новое событие свободно
	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}
	return s.repo.CreateEvent(ctx, event)
}

// Получение события по ID
func (s *CalendarService) GetEventByID(ctx context.Context, id uuid.UUID) (*models.Event, error) {
	return s.repo.GetEventByID(ctx, id)
}

// Бронирование события
func (s *CalendarService) BookEvent(ctx context.Context, eventID uuid.UUID, patientID uuid.UUID, patientEmail, appointmentType string) error {
	event, err := s.repo.GetEventByID(ctx, eventID)
	if err != nil {
		return err
	}

	if event.Status != "available" {
		return errors.New("event is not available for booking")
	}

	event.Status = "booked"
	event.PatientID = &patientID
	if appointmentType != "" {
		event.AppointmentType = appointmentType
	}
	event.UpdatedAt = time.Now()

	// Логируем для отладки
	s.logger.Infof("Booking event %s for patient %s with appointment_type=%s", eventID.String(), patientID.String(), event.AppointmentType)

	if err := s.repo.UpdateEvent(ctx, event); err != nil {
		return err
	}

	doctor, err := s.DoctorRepo.GetDoctorByID(ctx, event.SpecialistID)
	if err != nil {
		return err
	}

	err = s.notificationCl.SendBookingNotifications(ctx, event, patientEmail, doctor.Email, patientID, doctor.ID)
	if err != nil {
		s.logger.Error("Failed to send notifications:", err)
		// Не прерываем бронирование, но фиксируем ошибку
	}

	return nil
}

// Отмена бронирования события
func (s *CalendarService) CancelBooking(ctx context.Context, eventID uuid.UUID) error {
	event, err := s.repo.GetEventByID(ctx, eventID)
	if err != nil {
		return err
	}
	if event.Status != "booked" {
		return errors.New("event is not booked")
	}

	event.Status = "canceled"
	event.PatientID = nil
	event.UpdatedAt = time.Now()

	if err := s.repo.UpdateEvent(ctx, event); err != nil {
		return err
	}

	// Отправляем уведомление об отмене
	return s.notificationCl.SendCancelNotification(ctx, event)
}

func (s *CalendarService) GetEventsBySpecialist(ctx context.Context, specialistID uuid.UUID) ([]*models.Event, error) {
	return s.repo.GetEventsBySpecialist(ctx, specialistID)
}

func (s *CalendarService) GetAllEvents(ctx context.Context) ([]*models.Event, error) {
	return s.repo.GetAllEvents(ctx)
}

// Добавь другие методы при необходимости: UpdateEvent, DeleteEvent, GetEventsBySpecialist и т.д.
