package router

import (
	"github.com/labstack/echo/v4"

	"NotificationService/internal/delivery/http/handler"
	"NotificationService/internal/service"
)

// SetupRoutes настраивает все маршруты для Echo
func SetupRoutes(e *echo.Echo, notificationService service.NotificationService) {
	// Основные API маршруты
	api := e.Group("/api/v1")
	h := handler.NewNotificationHandler(notificationService)
	h.RegisterRoutes(api)

	// 🆕 Удобные маршруты для пользователей (защита добавляется в main.go)
	notifications := e.Group("/notifications")
	notifications.GET("/my", h.GetMyNotifications) // Мои уведомления
	notifications.GET("/:id", h.GetByID)           // Конкретное уведомление
	notifications.PUT("/:id/sent", h.MarkAsSent)   // Отметить как отправленное
}
