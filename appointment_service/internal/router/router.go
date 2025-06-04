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
	e.Use(middleware.Logger())
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
		doctor.POST("/schedules", handler.CreateSchedule)                   // Создать расписание
		doctor.GET("/schedules", handler.GetDoctorSchedules)                // Получить все расписания
		doctor.PUT("/schedules/:id", handler.UpdateSchedule)                // Обновить расписание
		doctor.DELETE("/schedules/:id", handler.DeleteSchedule)             // Удалить расписание
		doctor.PATCH("/schedules/:id/toggle", handler.ToggleSchedule)       // Активировать/деактивировать
		doctor.POST("/schedules/:id/generate-slots", handler.GenerateSlots) // Генерация слотов

		// Doctor's appointments - врач видит свои записи
		doctor.GET("/appointments", handler.GetDoctorAppointments)

		// Exception management - только врачи могут создавать исключения
		doctor.POST("/exceptions", handler.AddException)
		doctor.GET("/exceptions", handler.GetDoctorExceptions)
	}

	// Patient routes (только для пациентов)
	patient := api.Group("/patient")
	patient.Use(utilsMiddleware.RequirePatient())
	{
		// Patient's appointments - пациент видит свои записи
		patient.GET("/appointments", handler.GetPatientAppointments)
	}

	// Public routes (для всех авторизованных пользователей - врачей и пациентов)
	public := api.Group("")
	public.Use(utilsMiddleware.RequireDoctorOrPatient())
	{
		// View available slots - все могут видеть доступные слоты
		public.GET("/doctors/:id/available-slots", handler.GetAvailableSlots)

		// Appointment management - пациенты бронируют, все могут отменять
		public.POST("/appointments/:id/book", handler.BookAppointment)
		public.POST("/appointments/:id/cancel", handler.CancelAppointment)
	}
}
