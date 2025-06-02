package handler

import (
	"net/http"
	"time"

	"CalendarService/internal/domain/models"
	"CalendarService/internal/service"
	"CalendarService/internal/utils"
	"CalendarService/pkg/logger"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type EventHandler struct {
	service *service.CalendarService
	logger  *logger.Logger
}

func NewEventHandler(svc *service.CalendarService, log *logger.Logger) *EventHandler {
	return &EventHandler{
		service: svc,
		logger:  log,
	}
}

// CreateEvent - POST /api/events
func (h *EventHandler) CreateEvent(c echo.Context) error {
	ctx := c.Request().Context()
	event := new(models.Event)

	bindErr := c.Bind(event)
	if bindErr != nil {
		return utils.JSONError(c, http.StatusBadRequest, "invalid request body")
	}

	createErr := h.service.CreateEvent(ctx, event)
	if createErr != nil {
		return utils.JSONError(c, http.StatusInternalServerError, "failed to create event")
	}

	return utils.JSONSuccess(c, event)
}

// GetEventByID - GET /api/events/:id
func (h *EventHandler) GetEventByID(c echo.Context) error {
	ctx := c.Request().Context()
	idStr := c.Param("id")
	id, convErr := uuid.Parse(idStr)
	if convErr != nil {
		return utils.JSONError(c, http.StatusBadRequest, "invalid event ID")
	}

	event, getErr := h.service.GetEventByID(ctx, id)
	if getErr != nil {
		return utils.JSONError(c, http.StatusNotFound, "event not found")
	}

	return utils.JSONSuccess(c, event)
}

// GetEvents - GET /api/events
func (h *EventHandler) GetEvents(c echo.Context) error {
	ctx := c.Request().Context()
	specialistIDStr := c.QueryParam("specialist_id")

	var events []*models.Event
	var err error

	if specialistIDStr != "" {
		specialistID, convErr := uuid.Parse(specialistIDStr)
		if convErr != nil {
			return utils.JSONError(c, http.StatusBadRequest, "invalid specialist_id")
		}
		events, err = h.service.GetEventsBySpecialist(ctx, specialistID)
	} else {
		events, err = h.service.GetAllEvents(ctx)
	}

	if err != nil {
		return utils.JSONError(c, http.StatusInternalServerError, "failed to get events")
	}

	return utils.JSONSuccess(c, events)
}

type BookEventRequest struct {
	AppointmentType string `json:"appointment_type"`
}

func (h *EventHandler) BookEvent(c echo.Context) error {
	ctx := c.Request().Context()
	eventIDStr := c.Param("id")
	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		return utils.JSONError(c, http.StatusBadRequest, "invalid event ID")
	}

	patientIDVal := c.Get("user_id")
	patientID, ok := patientIDVal.(uuid.UUID)
	if !ok {
		return utils.JSONError(c, http.StatusUnauthorized, "unauthorized: invalid user ID")
	}

	patientEmailVal := c.Get("email")
	patientEmail, ok := patientEmailVal.(string)
	if !ok || patientEmail == "" {
		return utils.JSONError(c, http.StatusUnauthorized, "unauthorized: invalid email")
	}

	var req BookEventRequest
	if err := c.Bind(&req); err != nil {
		return utils.JSONError(c, http.StatusBadRequest, "invalid request body")
	}

	err = h.service.BookEvent(ctx, eventID, patientID, patientEmail, req.AppointmentType)
	if err != nil {
		h.logger.Error("Failed to book event:", err)
		return utils.JSONError(c, http.StatusBadRequest, err.Error())
	}

	return utils.JSONSuccess(c, map[string]string{"message": "event booked successfully"})
}

// CancelBooking - POST /api/events/:id/cancel
func (h *EventHandler) CancelBooking(c echo.Context) error {
	ctx := c.Request().Context()
	idStr := c.Param("id")
	eventID, convErr := uuid.Parse(idStr)
	if convErr != nil {
		return utils.JSONError(c, http.StatusBadRequest, "invalid event ID")
	}

	cancelErr := h.service.CancelBooking(ctx, eventID)
	if cancelErr != nil {
		return utils.JSONError(c, http.StatusBadRequest, cancelErr.Error())
	}

	return utils.JSONSuccess(c, map[string]string{"message": "booking canceled successfully"})
}

// Создание слотов (интервалов) для доктора
func (h *EventHandler) CreateSlots(c echo.Context) error {
	ctx := c.Request().Context()

	// Получаем doctorID из контекста (middleware JWT должен положить)
	doctorID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return utils.JSONError(c, http.StatusUnauthorized, "unauthorized")
	}

	// Структура запроса с интервалами
	type Slot struct {
		StartTime       string `json:"start_time"`
		EndTime         string `json:"end_time"`
		Title           string `json:"title,omitempty"`
		Description     string `json:"description,omitempty"`
		AppointmentType string `json:"appointment_type"` // новый параметр
	}

	var req struct {
		Slots []Slot `json:"slots"`
	}

	if err := c.Bind(&req); err != nil {
		h.logger.Error("Failed to bind slots request:", err)
		return utils.JSONError(c, http.StatusBadRequest, "invalid request body")
	}

	// Валидируем и конвертируем время, создаём события
	for _, slot := range req.Slots {
		start, err := time.Parse(time.RFC3339, slot.StartTime)
		if err != nil {
			return utils.JSONError(c, http.StatusBadRequest, "invalid start_time format")
		}
		end, err := time.Parse(time.RFC3339, slot.EndTime)
		if err != nil {
			return utils.JSONError(c, http.StatusBadRequest, "invalid end_time format")
		}

		event := &models.Event{
			ID:              uuid.New(),
			Title:           slot.Title,
			Description:     slot.Description,
			StartTime:       start,
			EndTime:         end,
			SpecialistID:    doctorID,
			Status:          "available",
			AppointmentType: slot.AppointmentType, // сохраняем тип приёма
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		err = h.service.CreateEvent(ctx, event)
		if err != nil {
			h.logger.Error("Failed to create event:", err)
			return utils.JSONError(c, http.StatusInternalServerError, "failed to create event")
		}
	}

	return utils.JSONSuccess(c, map[string]string{"message": "slots created successfully"})
}
