package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"NotificationService/internal/domain/models"
	"NotificationService/internal/domain/repository"
	"NotificationService/internal/infrastructure/codegen"
	"NotificationService/internal/infrastructure/email"
	"NotificationService/internal/infrastructure/telegram"

	"github.com/printprince/vitalem/logger_service/pkg/logger"

	"github.com/google/uuid"
)

type NotificationService interface {
	Send(ctx context.Context, notification *models.Notification) error
	Get(ctx context.Context, id int64) (*models.Notification, error)
	List(ctx context.Context, recipientID uuid.UUID) ([]*models.Notification, error)
	MarkAsSent(ctx context.Context, id int64) error
}

type notificationService struct {
	repo       repository.NotificationRepository
	email      email.Sender
	telegram   telegram.Sender
	codegen    codegen.Generator
	log        *logger.Client
	identity   *IdentityNotificationService
	patient    *PatientNotificationService
	specialist *SpecialistNotificationService
	calendar   *CalendarNotificationService
	fileserver *FileServerNotificationService
}

func NewNotificationService(
	repo repository.NotificationRepository,
	emailSender email.Sender,
	telegramSender telegram.Sender,
	codeGenerator codegen.Generator,
	log *logger.Client,
) NotificationService {
	return &notificationService{
		repo:       repo,
		email:      emailSender,
		telegram:   telegramSender,
		codegen:    codeGenerator,
		log:        log,
		identity:   NewIdentityNotificationService(codeGenerator),
		patient:    NewPatientNotificationService(),
		specialist: NewSpecialistNotificationService(),
		calendar:   NewCalendarNotificationService(),
		fileserver: NewFileServerNotificationService(),
	}
}

func (s *notificationService) Send(ctx context.Context, notification *models.Notification) error {
	notification.CreatedAt = time.Now()
	notification.Status = models.StatusPending

	// Генерация сообщения, если не передано
	if notification.Message == "" {
		s.enrichMessage(notification)
	}

	// Сохраняем в БД
	err := s.repo.Create(ctx, notification)
	if err != nil {
		s.log.Error("Failed to create notification", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	// Отправка
	var sendErr error
	switch notification.Channel {
	case models.ChannelEmail:
		sendErr = s.email.Send(notification.Recipient, "Уведомление", notification.Message)
		sendErr = s.telegram.Send(notification.Message)
	case models.ChannelTelegram:
		sendErr = s.telegram.Send(notification.Message)
	default:
		sendErr = errors.New("unsupported notification channel")
	}

	// Обработка результата отправки
	if sendErr != nil {
		s.log.Error("Failed to send notification", map[string]interface{}{
			"error": sendErr.Error(),
		})
		notification.Status = models.StatusFailed
		lastErr := sendErr.Error()
		notification.LastError = &lastErr
		notification.Attempts++
	} else {
		now := time.Now()
		notification.Status = models.StatusSent
		notification.SentAt = &now
	}

	// Обновляем статус в БД
	if err := s.repo.UpdateStatus(ctx, notification); err != nil {
		s.log.Error("Failed to update notification status", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	return sendErr
}

func (s *notificationService) enrichMessage(notification *models.Notification) {
	typ := string(notification.Type)

	switch {
	case strings.HasPrefix(typ, "user."):
		s.identity.Enrich(notification)
	case strings.HasPrefix(typ, "appointment."), strings.HasPrefix(typ, "lab."), strings.HasPrefix(typ, "patient."):
		s.patient.Enrich(notification)
	case strings.HasPrefix(typ, "specialist."):
		s.specialist.Enrich(notification)
	case strings.HasPrefix(typ, "calendar."):
		s.calendar.Enrich(notification)
	case strings.HasPrefix(typ, "file."):
		s.fileserver.Enrich(notification)
	default:
		s.log.Error("Unknown notification type; message left empty", map[string]interface{}{
			"type": typ,
		})
	}
}

func (s *notificationService) Get(ctx context.Context, id int64) (*models.Notification, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *notificationService) List(ctx context.Context, recipientID uuid.UUID) ([]*models.Notification, error) {
	return s.repo.ListByRecipient(ctx, recipientID)
}

func (s *notificationService) MarkAsSent(ctx context.Context, id int64) error {
	return s.repo.MarkAsSent(ctx, id)
}
