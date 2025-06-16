package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/printprince/vitalem/appointment_service/internal/models"
	"github.com/printprince/vitalem/appointment_service/internal/service"
	"github.com/printprince/vitalem/logger_service/pkg/logger"
)

// AppointmentHandler - обработчик HTTP запросов
type AppointmentHandler struct {
	service service.AppointmentService
	logger  *logger.Client
}

// NewAppointmentHandler - создание нового обработчика
func NewAppointmentHandler(service service.AppointmentService) *AppointmentHandler {
	return &AppointmentHandler{service: service}
}

// SetLogger - устанавливает логгер для хендлера
func (h *AppointmentHandler) SetLogger(loggerClient *logger.Client) {
	h.logger = loggerClient
}

// logInfo - вспомогательный метод для информационного логирования
func (h *AppointmentHandler) logInfo(message string, metadata map[string]interface{}) {
	if h.logger != nil {
		h.logger.Info(message, metadata)
	}
}

// logError - вспомогательный метод для логирования ошибок
func (h *AppointmentHandler) logError(message string, metadata map[string]interface{}) {
	if h.logger != nil {
		h.logger.Error(message, metadata)
	}
}

// === SCHEDULE ENDPOINTS ===

// CreateSchedule - POST /api/doctor/schedules
func (h *AppointmentHandler) CreateSchedule(c echo.Context) error {
	// Получаем user_id из JWT контекста
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		h.logError("Invalid user ID in token", map[string]interface{}{
			"endpoint": "CreateSchedule",
		})
		return c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error:   "Invalid user ID in token",
		})
	}

	var req models.CreateScheduleRequest
	if err := c.Bind(&req); err != nil {
		h.logError("Invalid request body", map[string]interface{}{
			"endpoint": "CreateSchedule",
			"userID":   userID.String(),
			"error":    err.Error(),
		})
		return c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	// Валидация входящих данных
	if err := c.Validate(&req); err != nil {
		h.logError("Validation error", map[string]interface{}{
			"endpoint": "CreateSchedule",
			"userID":   userID.String(),
			"error":    err.Error(),
		})
		return c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	h.logInfo("Creating schedule", map[string]interface{}{
		"endpoint":     "CreateSchedule",
		"userID":       userID.String(),
		"scheduleName": req.Name,
	})

	response, err := h.service.CreateSchedule(userID, &req)
	if err != nil {
		h.logError("Failed to create schedule", map[string]interface{}{
			"endpoint": "CreateSchedule",
			"userID":   userID.String(),
			"error":    err.Error(),
		})
		return c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	h.logInfo("Schedule created successfully", map[string]interface{}{
		"endpoint":   "CreateSchedule",
		"userID":     userID.String(),
		"scheduleID": response.ID.String(),
	})

	return c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Data:    response,
	})
}

// GetDoctorSchedules - GET /api/doctor/schedules
func (h *AppointmentHandler) GetDoctorSchedules(c echo.Context) error {
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		h.logError("Invalid user ID in token", map[string]interface{}{
			"endpoint": "GetDoctorSchedules",
		})
		return c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error:   "Invalid user ID in token",
		})
	}

	h.logInfo("Getting doctor schedules", map[string]interface{}{
		"endpoint": "GetDoctorSchedules",
		"userID":   userID.String(),
	})

	schedules, err := h.service.GetDoctorSchedules(userID)
	if err != nil {
		h.logError("Failed to get doctor schedules", map[string]interface{}{
			"endpoint": "GetDoctorSchedules",
			"userID":   userID.String(),
			"error":    err.Error(),
		})
		return c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	h.logInfo("Doctor schedules retrieved successfully", map[string]interface{}{
		"endpoint":      "GetDoctorSchedules",
		"userID":        userID.String(),
		"scheduleCount": len(schedules),
	})

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

	// Проверяем, есть ли тело запроса
	var req models.ToggleScheduleRequest
	var hasRequestBody bool

	// Читаем содержимое запроса
	if err := c.Bind(&req); err != nil {
		// Если ошибка парсинга, это не обязательно означает отсутствие тела
		return c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	// Проверяем, был ли передан Content-Length или Content-Type
	contentLength := c.Request().Header.Get("Content-Length")
	hasRequestBody = contentLength != "" && contentLength != "0"

	response, err := h.service.ToggleSchedule(userID, scheduleID, &req, hasRequestBody)
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

	response, err := h.service.GenerateSlots(userID, scheduleID, &req)
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

// DeleteScheduleSlots - DELETE /api/doctor/schedules/:id/slots
func (h *AppointmentHandler) DeleteScheduleSlots(c echo.Context) error {
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

	if err := h.service.DeleteScheduleSlots(userID, scheduleID); err != nil {
		return c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    "Schedule slots deleted successfully",
	})
}

// GetGeneratedSlots - GET /api/doctor/schedules/:id/generated-slots
func (h *AppointmentHandler) GetGeneratedSlots(c echo.Context) error {
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		h.logError("Failed to get user ID from context", map[string]interface{}{
			"endpoint": "GetGeneratedSlots",
		})
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}

	scheduleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.logError("Invalid schedule ID format", map[string]interface{}{
			"endpoint":   "GetGeneratedSlots",
			"scheduleID": c.Param("id"),
			"userID":     userID.String(),
			"error":      err.Error(),
		})
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid schedule ID format")
	}

	// Получаем параметры запроса
	startDate := c.QueryParam("start_date")
	endDate := c.QueryParam("end_date")

	// Валидируем параметры
	if startDate == "" || endDate == "" {
		h.logError("Missing required query parameters", map[string]interface{}{
			"endpoint":   "GetGeneratedSlots",
			"userID":     userID.String(),
			"scheduleID": scheduleID.String(),
			"startDate":  startDate,
			"endDate":    endDate,
		})
		return echo.NewHTTPError(http.StatusBadRequest, "start_date and end_date parameters are required")
	}

	// Простая валидация формата даты
	if len(startDate) != 10 || len(endDate) != 10 {
		h.logError("Invalid date format", map[string]interface{}{
			"endpoint":   "GetGeneratedSlots",
			"userID":     userID.String(),
			"scheduleID": scheduleID.String(),
			"startDate":  startDate,
			"endDate":    endDate,
		})
		return echo.NewHTTPError(http.StatusBadRequest, "Date format should be YYYY-MM-DD")
	}

	h.logInfo("Getting generated slots", map[string]interface{}{
		"endpoint":   "GetGeneratedSlots",
		"userID":     userID.String(),
		"scheduleID": scheduleID.String(),
		"startDate":  startDate,
		"endDate":    endDate,
	})

	response, err := h.service.GetGeneratedSlots(userID, scheduleID, startDate, endDate)
	if err != nil {
		h.logError("Failed to get generated slots", map[string]interface{}{
			"endpoint":   "GetGeneratedSlots",
			"userID":     userID.String(),
			"scheduleID": scheduleID.String(),
			"startDate":  startDate,
			"endDate":    endDate,
			"error":      err.Error(),
		})
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	h.logInfo("Generated slots retrieved successfully", map[string]interface{}{
		"endpoint":       "GetGeneratedSlots",
		"userID":         userID.String(),
		"scheduleID":     scheduleID.String(),
		"totalSlots":     response.Summary.TotalSlots,
		"availableSlots": response.Summary.AvailableSlots,
		"bookedSlots":    response.Summary.BookedSlots,
	})

	return c.JSON(http.StatusOK, &models.APIResponse{
		Success: true,
		Data:    response,
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

// CancelAppointment - POST /api/patient/appointments/:id/cancel
func (h *AppointmentHandler) CancelAppointment(c echo.Context) error {
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error:   "Invalid user ID in token",
		})
	}

	appointmentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid appointment ID",
		})
	}

	if err := h.service.CancelAppointment(userID, appointmentID); err != nil {
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

// GetDoctorAppointments - GET /api/doctor/appointments
func (h *AppointmentHandler) GetDoctorAppointments(c echo.Context) error {
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error:   "Invalid user ID in token",
		})
	}

	appointments, err := h.service.GetDoctorAppointments(userID)
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

// GetDoctorAppointmentByID - GET /api/doctor/appointments/:id
func (h *AppointmentHandler) GetDoctorAppointmentByID(c echo.Context) error {
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		h.logError("Invalid user ID in token", map[string]interface{}{
			"endpoint": "GetDoctorAppointmentByID",
		})
		return c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error:   "Invalid user ID in token",
		})
	}

	appointmentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.logError("Invalid appointment ID", map[string]interface{}{
			"endpoint":      "GetDoctorAppointmentByID",
			"userID":        userID.String(),
			"appointmentID": c.Param("id"),
			"error":         err.Error(),
		})
		return c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid appointment ID",
		})
	}

	h.logInfo("Getting appointment by ID", map[string]interface{}{
		"endpoint":      "GetDoctorAppointmentByID",
		"userID":        userID.String(),
		"appointmentID": appointmentID.String(),
	})

	appointment, err := h.service.GetDoctorAppointmentByID(userID, appointmentID)
	if err != nil {
		h.logError("Failed to get appointment", map[string]interface{}{
			"endpoint":      "GetDoctorAppointmentByID",
			"userID":        userID.String(),
			"appointmentID": appointmentID.String(),
			"error":         err.Error(),
		})
		return c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	h.logInfo("Appointment retrieved successfully", map[string]interface{}{
		"endpoint":      "GetDoctorAppointmentByID",
		"userID":        userID.String(),
		"appointmentID": appointmentID.String(),
		"status":        appointment.Status,
	})

	return c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    appointment,
	})
}

// GetPatientAppointments - GET /api/patient/appointments
func (h *AppointmentHandler) GetPatientAppointments(c echo.Context) error {
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error:   "Invalid user ID in token",
		})
	}

	appointments, err := h.service.GetPatientAppointments(userID)
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

// GetPatientAppointmentByID - GET /api/patient/appointments/:id
func (h *AppointmentHandler) GetPatientAppointmentByID(c echo.Context) error {
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		h.logError("Invalid user ID in token", map[string]interface{}{
			"endpoint": "GetPatientAppointmentByID",
		})
		return c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error:   "Invalid user ID in token",
		})
	}

	appointmentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.logError("Invalid appointment ID", map[string]interface{}{
			"endpoint":      "GetPatientAppointmentByID",
			"userID":        userID.String(),
			"appointmentID": c.Param("id"),
			"error":         err.Error(),
		})
		return c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid appointment ID",
		})
	}

	h.logInfo("Getting appointment by ID", map[string]interface{}{
		"endpoint":      "GetPatientAppointmentByID",
		"userID":        userID.String(),
		"appointmentID": appointmentID.String(),
	})

	appointment, err := h.service.GetPatientAppointmentByID(userID, appointmentID)
	if err != nil {
		h.logError("Failed to get appointment", map[string]interface{}{
			"endpoint":      "GetPatientAppointmentByID",
			"userID":        userID.String(),
			"appointmentID": appointmentID.String(),
			"error":         err.Error(),
		})
		return c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	h.logInfo("Appointment retrieved successfully", map[string]interface{}{
		"endpoint":      "GetPatientAppointmentByID",
		"userID":        userID.String(),
		"appointmentID": appointmentID.String(),
		"status":        appointment.Status,
	})

	return c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    appointment,
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
