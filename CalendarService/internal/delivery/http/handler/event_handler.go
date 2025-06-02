package handler

import (
	"fmt"
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

// CreateSchedule - POST /schedule - автоматическое создание графика
func (h *EventHandler) CreateSchedule(c echo.Context) error {
	ctx := c.Request().Context()

	// Получаем doctorID из контекста
	doctorID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return utils.JSONError(c, http.StatusUnauthorized, "unauthorized")
	}

	type ScheduleRequest struct {
		StartDate       string `json:"start_date"`       // "2024-06-03"
		EndDate         string `json:"end_date"`         // "2024-06-07"
		WorkDays        []int  `json:"work_days"`        // [1,2,3,4,5] (Пн-Пт)
		StartTime       string `json:"start_time"`       // "09:00"
		EndTime         string `json:"end_time"`         // "17:00"
		SlotDuration    int    `json:"slot_duration"`    // 30 (минут)
		BreakStart      string `json:"break_start"`      // "12:00"
		BreakEnd        string `json:"break_end"`        // "13:00"
		Title           string `json:"title"`            // "Консультация терапевта"
		Description     string `json:"description"`      // "Прием пациентов"
		AppointmentType string `json:"appointment_type"` // "offline" или "online"
	}

	var req ScheduleRequest
	if err := c.Bind(&req); err != nil {
		return utils.JSONError(c, http.StatusBadRequest, "invalid request body")
	}

	// Парсим даты
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return utils.JSONError(c, http.StatusBadRequest, "invalid start_date format (use YYYY-MM-DD)")
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return utils.JSONError(c, http.StatusBadRequest, "invalid end_date format (use YYYY-MM-DD)")
	}

	// Валидация
	if req.SlotDuration <= 0 || req.SlotDuration > 240 {
		return utils.JSONError(c, http.StatusBadRequest, "slot_duration must be between 1 and 240 minutes")
	}

	slotsCreated := 0

	// Генерируем слоты для каждого дня
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		weekday := int(d.Weekday())
		if weekday == 0 {
			weekday = 7
		} // Воскресенье = 7

		// Проверяем, рабочий ли день
		isWorkDay := false
		for _, wd := range req.WorkDays {
			if wd == weekday {
				isWorkDay = true
				break
			}
		}

		if !isWorkDay {
			continue
		}

		// Создаем слоты для этого дня
		daySlots, err := h.generateDaySlots(d, req, doctorID)
		if err != nil {
			return utils.JSONError(c, http.StatusBadRequest, err.Error())
		}

		// Сохраняем слоты в БД
		for _, slot := range daySlots {
			if err := h.service.CreateEvent(ctx, slot); err != nil {
				h.logger.Error("Failed to create slot:", err)
				return utils.JSONError(c, http.StatusInternalServerError, "failed to create slot")
			}
			slotsCreated++
		}
	}

	return utils.JSONSuccess(c, map[string]interface{}{
		"message":       "schedule created successfully",
		"slots_created": slotsCreated,
	})
}

// Генерирует слоты для одного дня
func (h *EventHandler) generateDaySlots(date time.Time, req struct {
	StartDate       string `json:"start_date"`
	EndDate         string `json:"end_date"`
	WorkDays        []int  `json:"work_days"`
	StartTime       string `json:"start_time"`
	EndTime         string `json:"end_time"`
	SlotDuration    int    `json:"slot_duration"`
	BreakStart      string `json:"break_start"`
	BreakEnd        string `json:"break_end"`
	Title           string `json:"title"`
	Description     string `json:"description"`
	AppointmentType string `json:"appointment_type"`
}, doctorID uuid.UUID) ([]*models.Event, error) {
	var slots []*models.Event

	// Парсим время начала и конца работы
	startTime, err := time.Parse("15:04", req.StartTime)
	if err != nil {
		return nil, fmt.Errorf("invalid start_time format (use HH:MM)")
	}

	endTime, err := time.Parse("15:04", req.EndTime)
	if err != nil {
		return nil, fmt.Errorf("invalid end_time format (use HH:MM)")
	}

	// Парсим время обеда (если указано)
	var breakStart, breakEnd time.Time
	hasBreak := req.BreakStart != "" && req.BreakEnd != ""
	if hasBreak {
		breakStart, err = time.Parse("15:04", req.BreakStart)
		if err != nil {
			return nil, fmt.Errorf("invalid break_start format (use HH:MM)")
		}
		breakEnd, err = time.Parse("15:04", req.BreakEnd)
		if err != nil {
			return nil, fmt.Errorf("invalid break_end format (use HH:MM)")
		}
	}

	// Создаем начальное время для данного дня
	currentTime := time.Date(date.Year(), date.Month(), date.Day(),
		startTime.Hour(), startTime.Minute(), 0, 0, date.Location())

	endDateTime := time.Date(date.Year(), date.Month(), date.Day(),
		endTime.Hour(), endTime.Minute(), 0, 0, date.Location())

	slotDuration := time.Duration(req.SlotDuration) * time.Minute

	// Генерируем слоты
	for currentTime.Before(endDateTime) {
		slotEnd := currentTime.Add(slotDuration)

		// Проверяем, не попадает ли слот на время обеда
		if hasBreak {
			breakStartDateTime := time.Date(date.Year(), date.Month(), date.Day(),
				breakStart.Hour(), breakStart.Minute(), 0, 0, date.Location())
			breakEndDateTime := time.Date(date.Year(), date.Month(), date.Day(),
				breakEnd.Hour(), breakEnd.Minute(), 0, 0, date.Location())

			// Если слот пересекается с обедом, пропускаем
			if currentTime.Before(breakEndDateTime) && slotEnd.After(breakStartDateTime) {
				// Если мы в обеде, переходим к концу обеда
				if currentTime.Before(breakEndDateTime) {
					currentTime = breakEndDateTime
					continue
				}
			}
		}

		// Если слот выходит за рабочее время, прекращаем
		if slotEnd.After(endDateTime) {
			break
		}

		// Создаем слот
		slot := &models.Event{
			ID:              uuid.New(),
			Title:           req.Title,
			Description:     req.Description,
			StartTime:       currentTime,
			EndTime:         slotEnd,
			SpecialistID:    doctorID,
			Status:          "available",
			AppointmentType: req.AppointmentType,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		slots = append(slots, slot)
		currentTime = slotEnd
	}

	return slots, nil
}

// GetAvailableSlots - GET /specialists/:specialist_id/slots - для пациентов
func (h *EventHandler) GetAvailableSlots(c echo.Context) error {
	ctx := c.Request().Context()

	// Получаем ID специалиста из URL
	specialistIDStr := c.Param("specialist_id")
	specialistID, err := uuid.Parse(specialistIDStr)
	if err != nil {
		return utils.JSONError(c, http.StatusBadRequest, "invalid specialist_id")
	}

	// Дополнительные фильтры из query params
	status := c.QueryParam("status")        // available, booked, canceled
	appointmentType := c.QueryParam("type") // online, offline
	dateFrom := c.QueryParam("from")        // 2024-06-03
	dateTo := c.QueryParam("to")            // 2024-06-07

	// Получаем все события специалиста
	events, err := h.service.GetEventsBySpecialist(ctx, specialistID)
	if err != nil {
		return utils.JSONError(c, http.StatusInternalServerError, "failed to get events")
	}

	// Фильтруем результаты
	var filteredEvents []*models.Event
	for _, event := range events {
		// Фильтр по статусу (по умолчанию только available)
		if status == "" && event.Status != "available" {
			continue
		}
		if status != "" && event.Status != status {
			continue
		}

		// Фильтр по типу приема
		if appointmentType != "" && event.AppointmentType != appointmentType {
			continue
		}

		// Фильтр по дате
		if dateFrom != "" {
			fromDate, parseErr := time.Parse("2006-01-02", dateFrom)
			if parseErr == nil && event.StartTime.Before(fromDate) {
				continue
			}
		}

		if dateTo != "" {
			toDate, parseErr := time.Parse("2006-01-02", dateTo)
			if parseErr == nil && event.StartTime.After(toDate.AddDate(0, 0, 1)) {
				continue
			}
		}

		filteredEvents = append(filteredEvents, event)
	}

	// Сортируем по времени начала
	for i := 0; i < len(filteredEvents)-1; i++ {
		for j := i + 1; j < len(filteredEvents); j++ {
			if filteredEvents[i].StartTime.After(filteredEvents[j].StartTime) {
				filteredEvents[i], filteredEvents[j] = filteredEvents[j], filteredEvents[i]
			}
		}
	}

	return utils.JSONSuccess(c, map[string]interface{}{
		"specialist_id": specialistID,
		"total_slots":   len(filteredEvents),
		"slots":         filteredEvents,
	})
}

// GetDoctorInfo - GET /specialists/:specialist_id/info - информация о враче + его слоты
func (h *EventHandler) GetDoctorInfo(c echo.Context) error {
	ctx := c.Request().Context()

	// Получаем ID специалиста из URL
	specialistIDStr := c.Param("specialist_id")
	specialistID, err := uuid.Parse(specialistIDStr)
	if err != nil {
		return utils.JSONError(c, http.StatusBadRequest, "invalid specialist_id")
	}

	// Получаем слоты врача
	events, err := h.service.GetEventsBySpecialist(ctx, specialistID)
	if err != nil {
		return utils.JSONError(c, http.StatusInternalServerError, "failed to get doctor events")
	}

	// Фильтруем только доступные слоты
	var availableSlots []*models.Event
	for _, event := range events {
		if event.Status == "available" {
			availableSlots = append(availableSlots, event)
		}
	}

	// Сортируем по времени начала
	for i := 0; i < len(availableSlots)-1; i++ {
		for j := i + 1; j < len(availableSlots); j++ {
			if availableSlots[i].StartTime.After(availableSlots[j].StartTime) {
				availableSlots[i], availableSlots[j] = availableSlots[j], availableSlots[i]
			}
		}
	}

	// Считаем статистику слотов
	totalSlots := len(events)
	availableSlotsCount := len(availableSlots)
	bookedSlotsCount := 0
	for _, event := range events {
		if event.Status == "booked" {
			bookedSlotsCount++
		}
	}

	return utils.JSONSuccess(c, map[string]interface{}{
		"specialist_id":     specialistID,
		"total_slots":       totalSlots,
		"available_slots":   availableSlotsCount,
		"booked_slots":      bookedSlotsCount,
		"upcoming_slots":    availableSlots[:min(5, len(availableSlots))], // Ближайшие 5 слотов
		"appointment_types": getUniqueAppointmentTypes(events),
	})
}

// Вспомогательная функция для получения уникальных типов приемов
func getUniqueAppointmentTypes(events []*models.Event) []string {
	typeMap := make(map[string]bool)
	for _, event := range events {
		if event.AppointmentType != "" {
			typeMap[event.AppointmentType] = true
		}
	}

	var types []string
	for appointmentType := range typeMap {
		types = append(types, appointmentType)
	}
	return types
}

// Вспомогательная функция min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
