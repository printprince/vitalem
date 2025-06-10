package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/printprince/vitalem/logger_service/pkg/logger"
	"github.com/printprince/vitalem/patient_service/internal/models"
	"github.com/printprince/vitalem/patient_service/internal/service"
)

// PatientHandlers структура обработчиков для пациентов
type PatientHandlers struct {
	patientService service.PatientService
	logger         *logger.Client
}

// NewPatientHandlers создает новый экземпляр обработчиков для пациентов
func NewPatientHandlers(patientService service.PatientService, logger *logger.Client) *PatientHandlers {
	return &PatientHandlers{
		patientService: patientService,
		logger:         logger,
	}
}

// RegisterPublicRoutes регистрирует публичные маршруты
func (h *PatientHandlers) RegisterPublicRoutes(g *echo.Group) {
	// Оставил на будущее
}

// RegisterProtectedRoutes регистрирует защищенные маршруты
func (h *PatientHandlers) RegisterProtectedRoutes(g *echo.Group) {
	// Защищенные маршруты для управления профилями пациентов
	patients := g.Group("/patients")
	patients.POST("", h.CreatePatient)
	patients.PUT("/:id", h.UpdatePatient)
	patients.DELETE("/:id", h.DeletePatient)
	patients.GET("/:id", h.GetPatientByID)

	// Маршрут для получения всех пациентов (только для докторов)
	g.GET("/patients", h.GetAllPatients)

	// Маршрут для получения пациента по ID пользователя
	g.GET("/users/:userID/patient", h.GetPatientByUserID)

	// Маршрут для обновления профиля пациента (для второго этапа регистрации)
	g.PUT("/users/:userID/patient/profile", h.UpdatePatientProfile)

	// Тестовый маршрут для проверки токена и роли
	g.GET("/me", h.GetCurrentUserInfo)
}

// CreatePatient обработчик для создания пациента
func (h *PatientHandlers) CreatePatient(c echo.Context) error {
	// Получаем запрос
	var req models.PatientCreateRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Error("Invalid request format", map[string]interface{}{
			"error": err.Error(),
		})
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	// Валидация запроса
	if err := c.Validate(&req); err != nil {
		h.logger.Error("Validation error", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	// Создаем пациента
	patient, err := h.patientService.CreatePatient(c.Request().Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create patient", map[string]interface{}{
			"error": err.Error(),
		})
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to create patient")
	}

	// Возвращаем http ответ
	return c.JSON(http.StatusCreated, patient)
}

// GetPatientByID обработчик для получения пациента по ID
func (h *PatientHandlers) GetPatientByID(c echo.Context) error {
	// Получаем айди с контекста
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.logger.Error("Invalid patient ID", map[string]interface{}{
			"error": err.Error(),
			"id":    c.Param("id"),
		})
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid patient ID")
	}

	// Получаем пациента
	patient, err := h.patientService.GetPatientByID(c.Request().Context(), id)
	if err != nil {
		h.logger.Error("Failed to get patient", map[string]interface{}{
			"error": err.Error(),
		})
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to get patient")
	}

	if patient == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Patient not found")
	}

	// Возвращаем пациента
	return c.JSON(http.StatusOK, patient)
}

// GetPatientByUserID обработчик для получения пациента по ID пользователя
func (h *PatientHandlers) GetPatientByUserID(c echo.Context) error {
	// Получаем айди с контекста
	userID, err := uuid.Parse(c.Param("userID"))
	if err != nil {
		h.logger.Error("Invalid user ID", map[string]interface{}{
			"error": err.Error(),
			"id":    c.Param("userID"),
		})
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid user ID")
	}

	// Получаем пациента
	patient, err := h.patientService.GetPatientByUserID(c.Request().Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get patient by user ID", map[string]interface{}{
			"error":  err.Error(),
			"userID": userID,
		})
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to get patient")
	}

	if patient == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Patient not found for this user")
	}

	// Возвращаем пациента
	return c.JSON(http.StatusOK, patient)
}

// GetAllPatients обработчик для получения всех пациентов
func (h *PatientHandlers) GetAllPatients(c echo.Context) error {
	// Проверяем роль пользователя (должна быть "doctor")
	role, ok := c.Get("role").(string)
	if !ok || role != "doctor" {
		h.logger.Error("Unauthorized access to patients list", map[string]interface{}{
			"role": role,
		})
		return echo.NewHTTPError(http.StatusForbidden, "Only doctors can access the patients list")
	}

	patients, err := h.patientService.GetAllPatients(c.Request().Context())
	if err != nil {
		h.logger.Error("Failed to get all patients", map[string]interface{}{
			"error": err.Error(),
		})
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get patients")
	}

	return c.JSON(http.StatusOK, patients)
}

// UpdatePatient обработчик для обновления пациента
func (h *PatientHandlers) UpdatePatient(c echo.Context) error {
	// Получаем id
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.logger.Error("Invalid patient ID", map[string]interface{}{
			"error": err.Error(),
			"id":    c.Param("id"),
		})
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid patient ID")
	}

	// Получаем запрос
	var req models.PatientCreateRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Error("Invalid request format", map[string]interface{}{
			"error": err.Error(),
		})
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	// Валидация запроса
	if err := c.Validate(&req); err != nil {
		h.logger.Error("Validation error", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	// Обновляем пациента
	patient, err := h.patientService.UpdatePatient(c.Request().Context(), id, &req)
	if err != nil {
		h.logger.Error("Failed to update patient", map[string]interface{}{
			"error": err.Error(),
			"id":    c.Param("id"),
		})
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to update patient")
	}

	if patient == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Patient not found")
	}

	// Возвращаем http ответ
	return c.JSON(http.StatusOK, patient)
}

// UpdatePatientProfile обработчик для обновления профиля пациента (для второго этапа регистрации)
func (h *PatientHandlers) UpdatePatientProfile(c echo.Context) error {
	// Получаем user ID из URL
	userID, err := uuid.Parse(c.Param("userID"))
	if err != nil {
		h.logger.Error("Invalid user ID", map[string]interface{}{
			"error": err.Error(),
			"id":    c.Param("userID"),
		})
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid user ID")
	}

	// Получаем данные профиля из запроса
	var req models.PatientCreateRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Error("Invalid request format", map[string]interface{}{
			"error": err.Error(),
		})
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	// Для обновления профиля НЕ проводим строгую валидацию (убираем c.Validate)
	// Позволяем частичные обновления

	// Обновляем или создаем профиль пациента
	patient, err := h.patientService.UpdatePatientProfile(c.Request().Context(), userID, &req)
	if err != nil {
		h.logger.Error("Failed to update patient profile", map[string]interface{}{
			"error":  err.Error(),
			"userID": userID,
		})
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to update patient profile")
	}

	// Возвращаем обновленный профиль
	return c.JSON(http.StatusOK, patient)
}

// DeletePatient обработчик для удаления пациента
func (h *PatientHandlers) DeletePatient(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.logger.Error("Invalid patient ID", map[string]interface{}{
			"error": err.Error(),
			"id":    c.Param("id"),
		})
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid patient ID")
	}

	if err := h.patientService.DeletePatient(c.Request().Context(), id); err != nil {
		h.logger.Error("Failed to delete patient", map[string]interface{}{
			"error": err.Error(),
			"id":    c.Param("id"),
		})
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to delete patient")
	}

	return c.NoContent(http.StatusNoContent)
}

// GetCurrentUserInfo возвращает информацию о текущем пользователе из JWT токена
func (h *PatientHandlers) GetCurrentUserInfo(c echo.Context) error {
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "User ID not found in context")
	}

	role, ok := c.Get("role").(string)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "Role not found in context")
	}

	email, _ := c.Get("email").(string)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"user_id": userID,
		"role":    role,
		"email":   email,
	})
}
