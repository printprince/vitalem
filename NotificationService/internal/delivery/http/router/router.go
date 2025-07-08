package router

import (
	"github.com/labstack/echo/v4"

	"NotificationService/internal/delivery/http/handler"
	"NotificationService/internal/service"
)

// SetupRoutes настраивает все маршруты для Echo
func SetupRoutes(e *echo.Echo, notificationService service.NotificationService) {
	// Основные API маршруты
	api := e.Group("/notifications")
	h := handler.NewNotificationHandler(notificationService)
	h.RegisterRoutes(api)
}
