package router

import (
	"NotificationService/internal/delivery/http/handler"
	"NotificationService/internal/service"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// SetupRoutes настраивает все маршруты для Echo
func SetupRoutes(e *echo.Echo, notifService service.NotificationService) {
	// Создаем хендлер
	h := handler.NewNotificationHandler(notifService)

	// Группа API v1
	v1 := e.Group("/api/v1")

	// Middleware для всех маршрутов
	v1.Use(middleware.Logger())
	v1.Use(middleware.Recover())

	// Маршруты уведомлений
	notifications := v1.Group("/notifications")
	{
		notifications.POST("", h.Send)
		notifications.GET("/:id", h.Get)
		notifications.GET("/recipient/:recipient_id", h.List)
		notifications.PUT("/:id/sent", h.MarkAsSent)
	}
}
