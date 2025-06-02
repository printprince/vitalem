package router

import (
	"CalendarService/internal/delivery/http/handler"
	"CalendarService/internal/service"
	"CalendarService/pkg/logger"
	"context"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	utilsMiddleware "github.com/printprince/vitalem/utils/middleware"
)

type Router struct {
	echo    *echo.Echo
	service *service.CalendarService
	logger  *logger.Logger
}

func NewRouter(svc *service.CalendarService, log *logger.Logger) *Router {
	e := echo.New()

	// Настраиваем middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Создаем обработчики
	h := handler.NewEventHandler(svc, log)

	// Получаем JWT секрет из переменной окружения
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "4324pkh23sk4jh342alhdlfl2sdjf" // fallback
	}

	// Группируем API роуты
	api := e.Group("/calendar")
	api.Use(utilsMiddleware.JWTMiddleware(jwtSecret))
	// Маршруты событий
	api.POST("/", h.CreateEvent)             // Создать событие
	api.GET("//:id", h.GetEventByID)         // Получить событие по ID
	api.GET("/", h.GetEvents)                // Список событий (можно с фильтрами)
	api.POST("/:id/book", h.BookEvent)       // Забронировать событие
	api.POST("/:id/cancel", h.CancelBooking) // Отменить бронь
	api.POST("/slots", h.CreateSlots)

	return &Router{
		echo:    e,
		service: svc,
		logger:  log,
	}
}

func (r *Router) Start(address string) error {
	return r.echo.Start(address)
}

func (r *Router) Shutdown(ctx context.Context) error {
	return r.echo.Shutdown(ctx)
}
