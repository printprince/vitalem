package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/printprince/vitalem/appointment_service/internal/models"
	"github.com/printprince/vitalem/appointment_service/internal/service"
)

// AppointmentHandler - обработчик HTTP запросов
type AppointmentHandler struct {
	service service.AppointmentService
}

// NewAppointmentHandler - создание нового обработчика
func NewAppointmentHandler(service service.AppointmentService) *AppointmentHandler {
	return &AppointmentHandler{service: service}
}

// === SCHEDULE ENDPOINTS ===

// CreateSchedule - POST /api/doctor/schedules
func (h *AppointmentHandler) CreateSchedule(c echo.Context) error {
	// Получаем user_id из JWT контекста
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error:   "Invalid user ID in token",
		})
	}

	var req models.CreateScheduleRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	// Валидация входящих данных
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	response, err := h.service.CreateSchedule(userID, &req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Data:    response,
	})
}

// GetDoctorSchedules - GET /api/doctor/schedules
func (h *AppointmentHandler) GetDoctorSchedules(c echo.Context) error {
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error:   "Invalid user ID in token",
		})
	}

	schedules, err := h.service.GetDoctorSchedules(userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    schedules,
	})
}

// UpdateSchedule - PUT /api/doctor/schedules/:id
func (h *AppointmentHandler) UpdateSchedule(c echo.Context) error {
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error:   "Invalid user ID in token",
		})
	}

	scheduleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid schedule ID",
		})
	}

	var req models.UpdateScheduleRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	// Валидация входящих данных
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	response, err := h.service.UpdateSchedule(userID, scheduleID, &req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    response,
	})
}

// DeleteSchedule - DELETE /api/doctor/schedules/:id
func (h *AppointmentHandler) DeleteSchedule(c echo.Context) error {
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error:   "Invalid user ID in token",
		})
	}

	scheduleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid schedule ID",
		})
	}

	if err := h.service.DeleteSchedule(userID, scheduleID); err != nil {
		return c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    "Schedule deleted successfully",
	})
}

// ToggleSchedule - PATCH /api/doctor/schedules/:id/toggle
func (h *AppointmentHandler) ToggleSchedule(c echo.Context) error {
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error:   "Invalid user ID in token",
		})
	}

	scheduleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid schedule ID",
		})
	}

	var req models.ToggleScheduleRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	response, err := h.service.ToggleSchedule(userID, scheduleID, &req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    response,
	})
}

// GenerateSlots - POST /api/doctor/schedules/:id/generate-slots
func (h *AppointmentHandler) GenerateSlots(c echo.Context) error {
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error:   "Invalid user ID in token",
		})
	}

	scheduleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid schedule ID",
		})
	}

	var req models.GenerateSlotsRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	// Валидация входящих данных
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	if err := h.service.GenerateSlots(userID, scheduleID, &req); err != nil {
		return c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    "Slots generated successfully",
	})
}

// === APPOINTMENT ENDPOINTS ===

// GetAvailableSlots - GET /api/doctors/:id/available-slots?date=2024-06-15
func (h *AppointmentHandler) GetAvailableSlots(c echo.Context) error {
	doctorID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid doctor ID",
		})
	}

	date := c.QueryParam("date")
	if date == "" {
		return c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Date parameter is required",
		})
	}

	slots, err := h.service.GetAvailableSlots(doctorID, date)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    slots,
	})
}

// BookAppointment - POST /api/appointments/:id/book
func (h *AppointmentHandler) BookAppointment(c echo.Context) error {
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error:   "Invalid user ID in token",
		})
	}

	// Проверяем роль - только пациенты могут бронировать
	role, _ := c.Get("role").(string)
	if role != "patient" {
		return c.JSON(http.StatusForbidden, models.APIResponse{
			Success: false,
			Error:   "Only patients can book appointments",
		})
	}

	appointmentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid appointment ID",
		})
	}

	var req models.BookAppointmentRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	// Валидация входящих данных
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	appointment, err := h.service.BookAppointment(userID, appointmentID, &req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    appointment,
	})
}

// CancelAppointment - POST /api/appointments/:id/cancel
func (h *AppointmentHandler) CancelAppointment(c echo.Context) error {
	appointmentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid appointment ID",
		})
	}

	if err := h.service.CancelAppointment(appointmentID); err != nil {
		return c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    "Appointment canceled successfully",
	})
}

// GetDoctorAppointments - GET /api/doctor/appointments?date=2024-06-15
func (h *AppointmentHandler) GetDoctorAppointments(c echo.Context) error {
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error:   "Invalid user ID in token",
		})
	}

	date := c.QueryParam("date")
	if date == "" {
		return c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Date parameter is required",
		})
	}

	appointments, err := h.service.GetDoctorAppointments(userID, date)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    appointments,
	})
}

// GetPatientAppointments - GET /api/patient/appointments?date=2024-06-15
func (h *AppointmentHandler) GetPatientAppointments(c echo.Context) error {
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error:   "Invalid user ID in token",
		})
	}

	date := c.QueryParam("date")
	if date == "" {
		return c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Date parameter is required",
		})
	}

	appointments, err := h.service.GetPatientAppointments(userID, date)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    appointments,
	})
}

// === EXCEPTION ENDPOINTS ===

// AddException - POST /api/doctor/exceptions
func (h *AppointmentHandler) AddException(c echo.Context) error {
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error:   "Invalid user ID in token",
		})
	}

	var req models.AddExceptionRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	// Валидация входящих данных
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	exception, err := h.service.AddException(userID, &req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Data:    exception,
	})
}

// GetDoctorExceptions - GET /api/doctor/exceptions?start_date=2024-06-01&end_date=2024-06-30
func (h *AppointmentHandler) GetDoctorExceptions(c echo.Context) error {
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error:   "Invalid user ID in token",
		})
	}

	startDate := c.QueryParam("start_date")
	endDate := c.QueryParam("end_date")

	if startDate == "" || endDate == "" {
		return c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "start_date and end_date parameters are required",
		})
	}

	exceptions, err := h.service.GetDoctorExceptions(userID, startDate, endDate)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    exceptions,
	})
}

// === HEALTH CHECK ===

// HealthCheck - GET /health
func (h *AppointmentHandler) HealthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    "Appointment Service is running",
	})
}
