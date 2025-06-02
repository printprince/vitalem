package service

import (
	"NotificationService/internal/domain/models"
	"NotificationService/internal/infrastructure/codegen"
)

type IdentityNotificationService struct {
	codegen codegen.Generator
}

func NewIdentityNotificationService(codegen codegen.Generator) *IdentityNotificationService {
	return &IdentityNotificationService{codegen: codegen}
}

func (s *IdentityNotificationService) Enrich(notification *models.Notification) {
	switch notification.Type {
	case models.UserRegistered:
		notification.Message = "Вы успешно зарегистрировались. Добро пожаловать!"

	case models.UserProfileUpdated:
		notification.Message = "Ваш профиль был обновлен."

	case models.UserPasswordChanged:
		notification.Message = "Ваш пароль был успешно изменен."

	case models.SecurityAlert:
		notification.Message = "Вход с нового устройства. Если это не вы — смените пароль."
	}
}
