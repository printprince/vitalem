package router

import (
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

	// Основная API группа
	api := e.Group("/api")
	api.Use(utilsMiddleware.JWTMiddleware(jwtSecret))

	// Doctor routes (только для врачей)
	doctor := api.Group("/doctor")
	doctor.Use(utilsMiddleware.RequireDoctor())
	{
		// Schedule management - только врачи могут управлять расписанием
		doctor.POST("/schedules", handler.CreateSchedule)                       // Создать расписание
		doctor.GET("/schedules", handler.GetDoctorSchedules)                    // Получить все расписания
		doctor.PUT("/schedules/:id", handler.UpdateSchedule)                    // Обновить расписание
		doctor.DELETE("/schedules/:id", handler.DeleteSchedule)                 // Удалить расписание
		doctor.PATCH("/schedules/:id/toggle", handler.ToggleSchedule)           // Активировать/деактивировать
		doctor.POST("/schedules/:id/generate-slots", handler.GenerateSlots)     // Генерация слотов
		doctor.DELETE("/schedules/:id/slots", handler.DeleteScheduleSlots)      // Удалить слоты расписания
		doctor.GET("/schedules/:id/generated-slots", handler.GetGeneratedSlots) // Получить детали сгенерированных слотов

		// Doctor's appointments - врач видит свои записи
		doctor.GET("/appointments", handler.GetDoctorAppointments)        // Все записи врача
		doctor.GET("/appointments/:id", handler.GetDoctorAppointmentByID) // Конкретная запись по ID

		// Exception management - только врачи могут создавать исключения
		doctor.POST("/exceptions", handler.AddException)
		doctor.GET("/exceptions", handler.GetDoctorExceptions)
	}

	// Patient routes (только для пациентов)
	patient := api.Group("/patient")
	patient.Use(utilsMiddleware.RequirePatient())
	{
		// Patient's appointments - пациент видит свои записи
		patient.GET("/appointments", handler.GetPatientAppointments)        // Все записи пациента
		patient.GET("/appointments/:id", handler.GetPatientAppointmentByID) // Конкретная запись по ID

		// Patient can cancel their own appointments
		patient.POST("/appointments/:id/cancel", handler.CancelAppointment)
	}

	// Public routes (для всех авторизованных пользователей - врачей и пациентов)
	public := api.Group("")
	public.Use(utilsMiddleware.RequireDoctorOrPatient())
	{
		// View available slots - все могут видеть доступные слоты
		public.GET("/doctors/:id/available-slots", handler.GetAvailableSlots)

		// Appointment booking - пациенты бронируют
		public.POST("/appointments/:id/book", handler.BookAppointment)
	}
}
