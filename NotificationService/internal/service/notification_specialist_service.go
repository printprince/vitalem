package service

import "NotificationService/internal/domain/models"

type SpecialistNotificationService struct{}

func NewSpecialistNotificationService() *SpecialistNotificationService {
	return &SpecialistNotificationService{}
}

func (s *SpecialistNotificationService) Enrich(notification *models.Notification) {
	switch notification.Type {
	case models.AppointmentNew:
		notification.Message = "У вас новая запись пациента."

	case models.AppointmentCanceled:
		notification.Message = "Пациент отменил запись."

	case models.UserProfileUpdated:
		notification.Message = "Ваш профиль специалиста был обновлен."

	case models.AppointmentRescheduled:
		notification.Message = "Время приема было изменено."
	}
}
