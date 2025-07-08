package router

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/printprince/vitalem/appointment_service/internal/handlers"
	utilsMiddleware "github.com/printprince/vitalem/utils/middleware"
)

// SetupRoutes - настройка маршрутов
func SetupRoutes(e *echo.Echo, handler *handlers.AppointmentHandler, jwtSecret string) {
	// Основные middleware
	e.Use(middleware.CORS())
	e.Use(utilsMiddleware.LoggerMiddleware()) // Кастомное middleware без health check логирования
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())

	// Настройка валидатора
	e.Validator = utilsMiddleware.NewValidator()

	// Health check (без авторизации)
	e.GET("/health", handler.HealthCheck)

	// Основная группа с JWT middleware
	protected := e.Group("")
	protected.Use(utilsMiddleware.JWTMiddleware(jwtSecret))

	// Doctor routes (только для врачей)
	doctor := protected.Group("/appointments/schedules")
	doctor.Use(utilsMiddleware.RequireDoctor())
	{
		// Schedule management - только врачи могут управлять расписанием
		doctor.POST("", handler.CreateSchedule)                       // Создать расписание
		doctor.GET("", handler.GetDoctorSchedules)                    // Получить все расписания
		doctor.PUT("/:id", handler.UpdateSchedule)                    // Обновить расписание
		doctor.DELETE("/:id", handler.DeleteSchedule)                 // Удалить расписание
		doctor.PATCH("/:id/toggle", handler.ToggleSchedule)           // Активировать/деактивировать
		doctor.POST("/:id/generate-slots", handler.GenerateSlots)     // Генерация слотов
		doctor.DELETE("/:id/slots", handler.DeleteScheduleSlots)      // Удалить слоты расписания
		doctor.GET("/:id/generated-slots", handler.GetGeneratedSlots) // Получить детали сгенерированных слотов
	}

	// Doctor exceptions (только для врачей)
	doctorExceptions := protected.Group("/appointments/exceptions")
	doctorExceptions.Use(utilsMiddleware.RequireDoctor())
	{
		doctorExceptions.POST("", handler.AddException)
		doctorExceptions.GET("", handler.GetDoctorExceptions)
	}

	// All appointments (для всех авторизованных пользователей)
	// Врачи видят свои записи, пациенты - свои записи
	appointments := protected.Group("/appointments")
	appointments.Use(utilsMiddleware.RequireDoctorOrPatient())
	{
		appointments.GET("", func(c echo.Context) error {
			// Проверяем роль пользователя
			role, ok := c.Get("role").(string)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "Role not found")
			}

			if role == "doctor" {
				return handler.GetDoctorAppointments(c)
			} else {
				return handler.GetPatientAppointments(c)
			}
		})

		appointments.GET("/:id", func(c echo.Context) error {
			// Проверяем роль пользователя
			role, ok := c.Get("role").(string)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "Role not found")
			}

			if role == "doctor" {
				return handler.GetDoctorAppointmentByID(c)
			} else {
				return handler.GetPatientAppointmentByID(c)
			}
		})

		appointments.POST("/:id/book", handler.BookAppointment)                     // Бронирование записи
		appointments.POST("/:id/cancel", handler.CancelAppointment)                 // Отмена записи
		appointments.GET("/doctors/:id/available-slots", handler.GetAvailableSlots) // Доступные слоты
	}
}
