package router

import (
	"github.com/labstack/echo/v4"

	"NotificationService/internal/delivery/http/handler"
	"NotificationService/internal/service"
)

// SetupRoutes настраивает все маршруты для Echo
func SetupRoutes(e *echo.Echo, notificationService service.NotificationService) {
	api := e.Group("/api/v1")
	h := handler.NewNotificationHandler(notificationService)
	h.RegisterRoutes(api)
}
