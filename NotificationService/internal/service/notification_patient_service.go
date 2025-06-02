package service

import "NotificationService/internal/domain/models"

type PatientNotificationService struct{}

func NewPatientNotificationService() *PatientNotificationService {
	return &PatientNotificationService{}
}

func (s *PatientNotificationService) Enrich(notification *models.Notification) {
	switch notification.Type {
	case models.AppointmentConfirmed:
		notification.Message = "Ваша запись к врачу подтверждена."

	case models.AppointmentCanceled:
		notification.Message = "Ваша запись была отменена."

	case models.AppointmentReminder:
		notification.Message = "Напоминание о приеме к врачу через 1 час."

	case models.LabResultsAvailable:
		notification.Message = "Результаты ваших анализов готовы."

	case models.UserProfileUpdated:
		notification.Message = "Ваш профиль пациента был обновлен."

	case models.PrescriptionIssued:
		notification.Message = "Вам выписан новый рецепт."

	case models.TreatmentStarted:
		notification.Message = "Началось ваше лечение."
	}
}
