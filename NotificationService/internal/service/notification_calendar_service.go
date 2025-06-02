package service

import "NotificationService/internal/domain/models"

type CalendarNotificationService struct{}

func NewCalendarNotificationService() *CalendarNotificationService {
	return &CalendarNotificationService{}
}

func (s *CalendarNotificationService) Enrich(notification *models.Notification) {
	switch notification.Type {
	case models.AppointmentReminder:
		notification.Message = "Напоминание о приеме через 1 час."

	case models.AppointmentNew:
		notification.Message = "У вас новая запись на прием."

	case models.AppointmentRescheduled:
		notification.Message = "Время вашего приема было изменено."

	case models.AppointmentBooked:
		notification.Message = "Ваша запись подтверждена."

	case models.AppointmentCanceled:
		notification.Message = "Ваша запись была отменена."
	}
}
