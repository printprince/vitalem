package service

import (
	"NotificationService/internal/domain/models"
)

type FileServerNotificationService struct{}

func NewFileServerNotificationService() *FileServerNotificationService {
	return &FileServerNotificationService{}
}

func (s *FileServerNotificationService) Enrich(notification *models.Notification) {
	switch notification.Type {
	case models.TestResultsReady:
		notification.Message = "Ваши результаты анализов готовы для просмотра."

	case models.LabResultsAvailable:
		notification.Message = "Лабораторные результаты доступны в системе."

	case models.SystemMaintenance:
		notification.Message = "Планируется техническое обслуживание системы."
	}
}
