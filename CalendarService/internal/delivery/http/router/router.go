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

	// Группируем API роуты с JWT проверкой
	api := e.Group("/calendar")
	api.Use(utilsMiddleware.JWTMiddleware(jwtSecret))

	// Эндпоинты только для врачей
	doctorRoutes := api.Group("")
	doctorRoutes.Use(utilsMiddleware.RequireDoctor())
	doctorRoutes.POST("/", h.CreateEvent)            // Создать событие - только врачи
	doctorRoutes.POST("/slots", h.CreateSlots)       // Создать слоты вручную - только врачи
	doctorRoutes.POST("/schedule", h.CreateSchedule) // Создать график автоматически - только врачи

	// Эндпоинты для врачей и пациентов
	commonRoutes := api.Group("")
	commonRoutes.Use(utilsMiddleware.RequireDoctorOrPatient())
	commonRoutes.GET("/:id", h.GetEventByID)          // Получить событие по ID
	commonRoutes.GET("/", h.GetEvents)                // Список событий (можно с фильтрами)
	commonRoutes.POST("/:id/book", h.BookEvent)       // Забронировать событие
	commonRoutes.POST("/:id/cancel", h.CancelBooking) // Отменить бронь

	// Удобные эндпоинты для пациентов (доступны всем авторизованным)
	commonRoutes.GET("/specialists/:specialist_id/slots", h.GetAvailableSlots) // Слоты врача

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
