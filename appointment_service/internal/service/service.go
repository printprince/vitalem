package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/printprince/vitalem/appointment_service/internal/models"
	"github.com/printprince/vitalem/appointment_service/internal/repository"
	"github.com/printprince/vitalem/logger_service/pkg/logger"
)

// AppointmentService - интерфейс сервиса
type AppointmentService interface {
	// Schedules
	CreateSchedule(doctorID uuid.UUID, req *models.CreateScheduleRequest) (*models.ScheduleResponse, error)
	GetDoctorSchedules(doctorID uuid.UUID) ([]*models.ScheduleResponse, error)
	UpdateSchedule(doctorID, scheduleID uuid.UUID, req *models.UpdateScheduleRequest) (*models.ScheduleResponse, error)
	DeleteSchedule(doctorID, scheduleID uuid.UUID) error
	ToggleSchedule(doctorID, scheduleID uuid.UUID, req *models.ToggleScheduleRequest, hasRequestBody bool) (*models.ScheduleResponse, error)
	GenerateSlots(doctorID, scheduleID uuid.UUID, req *models.GenerateSlotsRequest) (*models.GenerateSlotsResponse, error)
	GetGeneratedSlots(doctorID, scheduleID uuid.UUID, startDate, endDate string) (*models.GeneratedSlotsResponse, error)

	// Appointments
	GetAvailableSlots(doctorID uuid.UUID, date string) ([]*models.AvailableSlot, error)
	BookAppointment(patientID, appointmentID uuid.UUID, req *models.BookAppointmentRequest) (*models.AppointmentResponse, error)
	CancelAppointment(patientID, appointmentID uuid.UUID) error
	GetDoctorAppointments(doctorID uuid.UUID) ([]*models.AppointmentResponse, error)
	GetDoctorAppointmentByID(doctorID, appointmentID uuid.UUID) (*models.AppointmentResponse, error)
	GetPatientAppointments(patientID uuid.UUID) ([]*models.AppointmentResponse, error)
	GetPatientAppointmentByID(patientID, appointmentID uuid.UUID) (*models.AppointmentResponse, error)

	// Exceptions
	AddException(doctorID uuid.UUID, req *models.AddExceptionRequest) (*models.ExceptionResponse, error)
	GetDoctorExceptions(doctorID uuid.UUID, startDate, endDate string) ([]*models.ExceptionResponse, error)

	// New method for forcing clean slots of schedule
	DeleteScheduleSlots(doctorID, scheduleID uuid.UUID) error
}

// appointmentService - реализация сервиса
type appointmentService struct {
	repo   repository.AppointmentRepository
	logger *logger.Client
}

// NewAppointmentService - создание нового сервиса
func NewAppointmentService(repo repository.AppointmentRepository, loggerClient *logger.Client) AppointmentService {
	return &appointmentService{
		repo:   repo,
		logger: loggerClient,
	}
}

// SetLogger - устанавливает логгер для сервиса (deprecated, используйте NewAppointmentService)
func (s *appointmentService) SetLogger(loggerClient *logger.Client) {
	s.logger = loggerClient
}

// logInfo - вспомогательный метод для информационного логирования
func (s *appointmentService) logInfo(message string, metadata map[string]interface{}) {
	if s.logger != nil {
		s.logger.Info(message, metadata)
	}
}

// logError - вспомогательный метод для логирования ошибок
func (s *appointmentService) logError(message string, metadata map[string]interface{}) {
	if s.logger != nil {
		s.logger.Error(message, metadata)
	}
}

// === SCHEDULES ===

func (s *appointmentService) CreateSchedule(doctorID uuid.UUID, req *models.CreateScheduleRequest) (*models.ScheduleResponse, error) {
	s.logInfo("Creating schedule for doctor", map[string]interface{}{
		"doctorID":     doctorID.String(),
		"scheduleName": req.Name,
	})

	// При создании нового расписания ВСЕГДА деактивируем все существующие
	// чтобы у врача было только одно активное расписание
	s.logInfo("Deactivating all existing schedules for single active schedule policy", map[string]interface{}{
		"doctorID": doctorID.String(),
	})

	if err := s.deactivateOtherSchedules(doctorID); err != nil {
		s.logError("Failed to deactivate existing schedules", map[string]interface{}{
			"doctorID": doctorID.String(),
			"error":    err.Error(),
		})
		return nil, fmt.Errorf("failed to deactivate existing schedules: %w", err)
	}

	schedule := &models.DoctorSchedule{
		DoctorID:          doctorID,
		Name:              req.Name,
		StartTime:         req.StartTime,
		EndTime:           req.EndTime,
		BreakStart:        req.BreakStart,
		BreakEnd:          req.BreakEnd,
		SlotDuration:      req.SlotDuration,
		SlotTitle:         req.SlotTitle,
		AppointmentFormat: req.AppointmentFormat,
		IsActive:          true, // Новое расписание всегда активно
	}

	// Устанавливаем рабочие дни через новый метод
	schedule.SetWorkDays(req.WorkDays)

	if err := s.repo.CreateSchedule(schedule); err != nil {
		s.logError("Failed to create schedule in repository", map[string]interface{}{
			"doctorID": doctorID.String(),
			"error":    err.Error(),
		})
		return nil, fmt.Errorf("failed to create schedule: %w", err)
	}

	s.logInfo("Schedule created successfully", map[string]interface{}{
		"doctorID":   doctorID.String(),
		"scheduleID": schedule.ID.String(),
		"isActive":   schedule.IsActive,
	})

	return s.scheduleToResponse(schedule), nil
}

// checkScheduleConflicts проверяет конфликты времени с существующими активными расписаниями
func (s *appointmentService) checkScheduleConflicts(doctorID uuid.UUID, req *models.CreateScheduleRequest) error {
	existingSchedules, err := s.repo.GetDoctorSchedules(doctorID)
	if err != nil {
		return fmt.Errorf("failed to get existing schedules: %w", err)
	}

	startTime, err := time.Parse("15:04", req.StartTime)
	if err != nil {
		return fmt.Errorf("invalid start time format: %w", err)
	}

	endTime, err := time.Parse("15:04", req.EndTime)
	if err != nil {
		return fmt.Errorf("invalid end time format: %w", err)
	}

	for _, existing := range existingSchedules {
		if !existing.IsActive {
			continue // пропускаем неактивные расписания
		}

		existingStart, _ := time.Parse("15:04", existing.StartTime)
		existingEnd, _ := time.Parse("15:04", existing.EndTime)

		// Проверяем пересечение рабочих дней
		if s.hasWorkDayConflict(req.WorkDays, existing.WorkDays()) {
			// Проверяем пересечение времени
			if s.hasTimeConflict(startTime, endTime, existingStart, existingEnd) {
				return fmt.Errorf("schedule conflicts with existing schedule '%s' on overlapping work days and times", existing.Name)
			}
		}
	}

	return nil
}

// hasWorkDayConflict проверяет есть ли пересечения в рабочих днях
func (s *appointmentService) hasWorkDayConflict(workDays1, workDays2 []int) bool {
	dayMap := make(map[int]bool)
	for _, day := range workDays1 {
		dayMap[day] = true
	}

	for _, day := range workDays2 {
		if dayMap[day] {
			return true
		}
	}

	return false
}

// hasTimeConflict проверяет пересекается ли время
func (s *appointmentService) hasTimeConflict(start1, end1, start2, end2 time.Time) bool {
	return start1.Before(end2) && end1.After(start2)
}

// deactivateOtherSchedules деактивирует все другие расписания врача (кроме указанного ID, если передан)
func (s *appointmentService) deactivateOtherSchedules(doctorID uuid.UUID, excludeIDs ...uuid.UUID) error {
	schedules, err := s.repo.GetDoctorSchedules(doctorID)
	if err != nil {
		return err
	}

	// Создаем мапу исключений для быстрого поиска
	excludeMap := make(map[uuid.UUID]bool)
	for _, id := range excludeIDs {
		excludeMap[id] = true
	}

	for _, schedule := range schedules {
		// Пропускаем расписания из списка исключений
		if excludeMap[schedule.ID] {
			continue
		}

		if schedule.IsActive {
			schedule.IsActive = false
			if err := s.repo.UpdateSchedule(schedule); err != nil {
				return fmt.Errorf("failed to deactivate schedule %s: %w", schedule.Name, err)
			}
			s.logInfo("Deactivated existing schedule", map[string]interface{}{
				"doctorID":   doctorID.String(),
				"scheduleID": schedule.ID.String(),
				"name":       schedule.Name,
			})
		}
	}

	return nil
}

func (s *appointmentService) GetDoctorSchedules(doctorID uuid.UUID) ([]*models.ScheduleResponse, error) {
	s.logInfo("Getting schedules for doctor", map[string]interface{}{
		"doctorID": doctorID.String(),
	})

	schedules, err := s.repo.GetDoctorSchedules(doctorID)
	if err != nil {
		s.logError("Failed to get schedules from repository", map[string]interface{}{
			"doctorID": doctorID.String(),
			"error":    err.Error(),
		})
		return nil, fmt.Errorf("failed to get schedules: %w", err)
	}

	responses := make([]*models.ScheduleResponse, len(schedules))
	for i, schedule := range schedules {
		responses[i] = s.scheduleToResponse(schedule)
	}

	s.logInfo("Schedules retrieved successfully", map[string]interface{}{
		"doctorID":      doctorID.String(),
		"scheduleCount": len(schedules),
	})

	return responses, nil
}

func (s *appointmentService) GenerateSlots(doctorID, scheduleID uuid.UUID, req *models.GenerateSlotsRequest) (*models.GenerateSlotsResponse, error) {
	s.logInfo("Starting slot generation", map[string]interface{}{
		"doctorID":   doctorID.String(),
		"scheduleID": scheduleID.String(),
		"startDate":  req.StartDate,
		"endDate":    req.EndDate,
	})

	schedule, err := s.repo.GetScheduleByID(scheduleID)
	if err != nil {
		s.logError("Schedule not found", map[string]interface{}{
			"doctorID":   doctorID.String(),
			"scheduleID": scheduleID.String(),
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("schedule not found: %w", err)
	}

	if schedule.DoctorID != doctorID {
		s.logError("Schedule ownership validation failed", map[string]interface{}{
			"doctorID":         doctorID.String(),
			"scheduleID":       scheduleID.String(),
			"scheduleDoctorID": schedule.DoctorID.String(),
		})
		return nil, errors.New("schedule doesn't belong to this doctor")
	}

	if !schedule.IsActive {
		s.logError("Cannot generate slots for inactive schedule", map[string]interface{}{
			"doctorID":   doctorID.String(),
			"scheduleID": scheduleID.String(),
		})
		return nil, errors.New("cannot generate slots for inactive schedule")
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		s.logError("Invalid start date format", map[string]interface{}{
			"doctorID":  doctorID.String(),
			"startDate": req.StartDate,
			"error":     err.Error(),
		})
		return nil, fmt.Errorf("invalid start date: %w", err)
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		s.logError("Invalid end date format", map[string]interface{}{
			"doctorID": doctorID.String(),
			"endDate":  req.EndDate,
			"error":    err.Error(),
		})
		return nil, fmt.Errorf("invalid end date: %w", err)
	}

	// Получаем исключения для периода
	exceptions, _ := s.repo.GetDoctorExceptions(doctorID, startDate, endDate)
	exceptionMap := make(map[string]*models.ScheduleException)
	for _, ex := range exceptions {
		dateStr := ex.Date.Format("2006-01-02")
		exceptionMap[dateStr] = ex
	}

	s.logInfo("Retrieved exceptions for period", map[string]interface{}{
		"doctorID":       doctorID.String(),
		"exceptionCount": len(exceptions),
	})

	// Шаг 1: Собираем ВСЕ потенциальные слоты, которые нужно создать
	var allSlotsToCreate []slotToCreate

	for date := startDate; !date.After(endDate); date = date.AddDate(0, 0, 1) {
		dateStr := date.Format("2006-01-02")
		weekday := int(date.Weekday())
		if weekday == 0 {
			weekday = 7 // Воскресенье = 7
		}

		// Проверяем исключения
		if exception, exists := exceptionMap[dateStr]; exists {
			if exception.Type == "day_off" {
				s.logInfo("Skipping day off", map[string]interface{}{
					"doctorID": doctorID.String(),
					"date":     dateStr,
					"reason":   exception.Reason,
				})
				continue // Пропускаем выходной
			}
			// Для кастомных часов используем их вместо обычного расписания
			if exception.Type == "custom_hours" && exception.CustomStartTime != nil && exception.CustomEndTime != nil {
				daySlots := s.generateSlotsForDayCheck(date, *exception.CustomStartTime, *exception.CustomEndTime, nil, nil, schedule)
				allSlotsToCreate = append(allSlotsToCreate, daySlots...)
				continue
			}
		}

		// Проверяем рабочие дни
		isWorkDay := false
		for _, workDay := range schedule.WorkDays() {
			if workDay == weekday {
				isWorkDay = true
				break
			}
		}

		if !isWorkDay {
			continue
		}

		// Собираем слоты для обычного дня
		daySlots := s.generateSlotsForDayCheck(date, schedule.StartTime, schedule.EndTime, schedule.BreakStart, schedule.BreakEnd, schedule)
		allSlotsToCreate = append(allSlotsToCreate, daySlots...)
	}

	// Шаг 2: Проверяем ВСЕ слоты на конфликты
	s.logInfo("Checking all slots for conflicts", map[string]interface{}{
		"doctorID":   doctorID.String(),
		"totalSlots": len(allSlotsToCreate),
	})

	for _, slot := range allSlotsToCreate {
		if s.slotExists(schedule.DoctorID, slot.startTime, slot.endTime) {
			s.logError("Slot conflict detected - aborting generation", map[string]interface{}{
				"doctorID":      doctorID.String(),
				"conflictStart": slot.startTime.Format("2006-01-02 15:04:05"),
				"conflictEnd":   slot.endTime.Format("2006-01-02 15:04:05"),
			})
			return nil, fmt.Errorf("cannot generate slots: slot from %s to %s already exists. Please choose a different time period or delete existing slots first",
				slot.startTime.Format("2006-01-02 15:04"), slot.endTime.Format("15:04"))
		}
	}

	// Шаг 3: Если все проверки прошли - создаем ВСЕ слоты
	s.logInfo("No conflicts found - creating all slots", map[string]interface{}{
		"doctorID":          doctorID.String(),
		"totalSlots":        len(allSlotsToCreate),
		"appointmentFormat": schedule.AppointmentFormat,
	})

	totalSlotsCreated := 0
	for _, slot := range allSlotsToCreate {
		// Определяем тип записи исходя из формата расписания
		// Слоты должны наследовать точно такой же формат как у расписания
		appointmentType := schedule.AppointmentFormat

		appointment := &models.Appointment{
			StartTime:       slot.startTime,
			EndTime:         slot.endTime,
			DoctorID:        schedule.DoctorID,
			Title:           schedule.SlotTitle,
			Status:          "available",
			AppointmentType: appointmentType,
			ScheduleID:      &schedule.ID,
		}

		if err := s.repo.CreateAppointment(appointment); err != nil {
			s.logError("Failed to create appointment slot", map[string]interface{}{
				"doctorID":  schedule.DoctorID.String(),
				"startTime": slot.startTime.Format("2006-01-02 15:04:05"),
				"endTime":   slot.endTime.Format("2006-01-02 15:04:05"),
				"error":     err.Error(),
			})
			return nil, fmt.Errorf("failed to create slot: %w", err)
		} else {
			totalSlotsCreated++
		}
	}

	s.logInfo("Slot generation completed successfully", map[string]interface{}{
		"doctorID":          doctorID.String(),
		"scheduleID":        scheduleID.String(),
		"totalSlotsCreated": totalSlotsCreated,
	})

	message := fmt.Sprintf("Генерация завершена успешно: создано %d слотов", totalSlotsCreated)

	return &models.GenerateSlotsResponse{
		SlotsCreated: totalSlotsCreated,
		Message:      message,
	}, nil
}

// Структура для хранения информации о слоте, который нужно создать
type slotToCreate struct {
	startTime time.Time
	endTime   time.Time
}

// generateSlotsForDayCheck - собирает слоты для дня без создания
func (s *appointmentService) generateSlotsForDayCheck(date time.Time, startTime, endTime string, breakStart, breakEnd *string, schedule *models.DoctorSchedule) []slotToCreate {
	location := time.Local
	var slots []slotToCreate

	// Парсим время начала и конца
	start, err := time.ParseInLocation("2006-01-02 15:04", date.Format("2006-01-02")+" "+startTime, location)
	if err != nil {
		s.logError("Failed to parse start time", map[string]interface{}{
			"doctorID":  schedule.DoctorID.String(),
			"date":      date.Format("2006-01-02"),
			"startTime": startTime,
			"error":     err.Error(),
		})
		return slots
	}

	end, err := time.ParseInLocation("2006-01-02 15:04", date.Format("2006-01-02")+" "+endTime, location)
	if err != nil {
		s.logError("Failed to parse end time", map[string]interface{}{
			"doctorID": schedule.DoctorID.String(),
			"date":     date.Format("2006-01-02"),
			"endTime":  endTime,
			"error":    err.Error(),
		})
		return slots
	}

	// Парсим перерыв, если есть
	var breakStartTime, breakEndTime *time.Time
	if breakStart != nil && breakEnd != nil {
		bStart, err := time.ParseInLocation("2006-01-02 15:04", date.Format("2006-01-02")+" "+*breakStart, location)
		if err == nil {
			breakStartTime = &bStart
		}

		bEnd, err := time.ParseInLocation("2006-01-02 15:04", date.Format("2006-01-02")+" "+*breakEnd, location)
		if err == nil {
			breakEndTime = &bEnd
		}
	}

	// Генерируем слоты
	slotDuration := time.Duration(schedule.SlotDuration) * time.Minute
	current := start

	for current.Add(slotDuration).Before(end) || current.Add(slotDuration).Equal(end) {
		slotEnd := current.Add(slotDuration)

		// Проверяем пересечение с перерывом
		if breakStartTime != nil && breakEndTime != nil {
			if current.Before(*breakEndTime) && slotEnd.After(*breakStartTime) {
				// Слот пересекается с перерывом, пропускаем
				current = *breakEndTime
				continue
			}
		}

		slots = append(slots, slotToCreate{
			startTime: current,
			endTime:   slotEnd,
		})

		current = slotEnd
	}

	return slots
}

// slotExists простая проверка существования слота
func (s *appointmentService) slotExists(doctorID uuid.UUID, startTime, endTime time.Time) bool {
	exists, err := s.repo.CheckSlotExists(doctorID, startTime, endTime)
	if err != nil {
		s.logError("Error checking slot existence", map[string]interface{}{
			"doctorID":  doctorID.String(),
			"startTime": startTime.Format("2006-01-02 15:04:05"),
			"endTime":   endTime.Format("2006-01-02 15:04:05"),
			"error":     err.Error(),
		})
		return true // В случае ошибки считаем что слот существует (безопасно)
	}
	return exists
}

// === APPOINTMENTS ===

func (s *appointmentService) GetAvailableSlots(doctorID uuid.UUID, date string) ([]*models.AvailableSlot, error) {
	startDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}
	endDate := startDate.AddDate(0, 0, 1)

	appointments, err := s.repo.GetAvailableSlots(doctorID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get available slots: %w", err)
	}

	slots := make([]*models.AvailableSlot, len(appointments))
	for i, appointment := range appointments {
		duration := int(appointment.EndTime.Sub(appointment.StartTime).Minutes())
		slots[i] = &models.AvailableSlot{
			ID:              appointment.ID,
			StartTime:       appointment.StartTime,
			EndTime:         appointment.EndTime,
			Duration:        duration,
			Title:           appointment.Title,
			AppointmentType: appointment.AppointmentType,
		}
	}

	return slots, nil
}

func (s *appointmentService) BookAppointment(patientID, appointmentID uuid.UUID, req *models.BookAppointmentRequest) (*models.AppointmentResponse, error) {
	appointment, err := s.repo.GetAppointmentByID(appointmentID)
	if err != nil {
		return nil, fmt.Errorf("appointment not found: %w", err)
	}

	if !appointment.IsAvailable() {
		return nil, errors.New("appointment is not available")
	}

	appointmentType := req.AppointmentType
	if appointmentType == "" {
		// Если тип не указан, выбираем по умолчанию в зависимости от слота
		if appointment.AppointmentType == "both" {
			appointmentType = "offline" // По умолчанию для "both" выбираем offline
		} else {
			appointmentType = appointment.AppointmentType // Используем тип слота
		}
	}

	// Проверяем совместимость запрашиваемого типа со слотом
	if !s.isAppointmentTypeCompatible(appointment.AppointmentType, appointmentType) {
		return nil, fmt.Errorf("appointment type '%s' is not compatible with slot type '%s'", appointmentType, appointment.AppointmentType)
	}

	appointment.Book(patientID, appointmentType, req.PatientNotes)

	if err := s.repo.UpdateAppointment(appointment); err != nil {
		return nil, fmt.Errorf("failed to book appointment: %w", err)
	}

	return s.appointmentToResponse(appointment), nil
}

// isAppointmentTypeCompatible проверяет совместимость типа записи со слотом
func (s *appointmentService) isAppointmentTypeCompatible(slotType, requestedType string) bool {
	// Если слот "both", то можно забронировать любой тип
	if slotType == "both" {
		return requestedType == "offline" || requestedType == "online"
	}

	// Для остальных случаев типы должны совпадать
	return slotType == requestedType
}

func (s *appointmentService) CancelAppointment(patientID, appointmentID uuid.UUID) error {
	appointment, err := s.repo.GetAppointmentByID(appointmentID)
	if err != nil {
		return fmt.Errorf("appointment not found: %w", err)
	}

	// Проверяем что запись принадлежит этому пациенту
	if appointment.PatientID == nil || *appointment.PatientID != patientID {
		return errors.New("appointment doesn't belong to this patient or is not booked")
	}

	// Проверяем что запись можно отменить (не уже отменена и не завершена)
	if appointment.Status != "booked" {
		return fmt.Errorf("cannot cancel appointment with status '%s'", appointment.Status)
	}

	appointment.Cancel()

	if err := s.repo.UpdateAppointment(appointment); err != nil {
		return fmt.Errorf("failed to cancel appointment: %w", err)
	}

	return nil
}

func (s *appointmentService) GetDoctorAppointments(doctorID uuid.UUID) ([]*models.AppointmentResponse, error) {
	appointments, err := s.repo.GetDoctorAppointments(doctorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get doctor appointments: %w", err)
	}

	responses := make([]*models.AppointmentResponse, len(appointments))
	for i, appointment := range appointments {
		responses[i] = s.appointmentToResponse(appointment)
	}

	return responses, nil
}

func (s *appointmentService) GetPatientAppointments(patientID uuid.UUID) ([]*models.AppointmentResponse, error) {
	appointments, err := s.repo.GetPatientAppointments(patientID)
	if err != nil {
		return nil, fmt.Errorf("failed to get patient appointments: %w", err)
	}

	responses := make([]*models.AppointmentResponse, len(appointments))
	for i, appointment := range appointments {
		responses[i] = s.appointmentToResponse(appointment)
	}

	return responses, nil
}

// === EXCEPTIONS ===

func (s *appointmentService) AddException(doctorID uuid.UUID, req *models.AddExceptionRequest) (*models.ExceptionResponse, error) {
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}

	exception := &models.ScheduleException{
		DoctorID:        doctorID,
		Date:            date,
		Type:            req.Type,
		CustomStartTime: req.CustomStartTime,
		CustomEndTime:   req.CustomEndTime,
		Reason:          req.Reason,
	}

	if err := s.repo.CreateException(exception); err != nil {
		return nil, fmt.Errorf("failed to create exception: %w", err)
	}

	return s.exceptionToResponse(exception), nil
}

func (s *appointmentService) GetDoctorExceptions(doctorID uuid.UUID, startDate, endDate string) ([]*models.ExceptionResponse, error) {
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start date: %w", err)
	}

	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end date: %w", err)
	}

	exceptions, err := s.repo.GetDoctorExceptions(doctorID, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to get exceptions: %w", err)
	}

	responses := make([]*models.ExceptionResponse, len(exceptions))
	for i, exception := range exceptions {
		responses[i] = s.exceptionToResponse(exception)
	}

	return responses, nil
}

// === HELPER METHODS ===

func (s *appointmentService) scheduleToResponse(schedule *models.DoctorSchedule) *models.ScheduleResponse {
	return &models.ScheduleResponse{
		ID:                schedule.ID,
		DoctorID:          schedule.DoctorID,
		Name:              schedule.Name,
		WorkDays:          schedule.WorkDays(),
		StartTime:         schedule.StartTime,
		EndTime:           schedule.EndTime,
		BreakStart:        schedule.BreakStart,
		BreakEnd:          schedule.BreakEnd,
		SlotDuration:      schedule.SlotDuration,
		SlotTitle:         schedule.SlotTitle,
		AppointmentFormat: schedule.AppointmentFormat,
		IsActive:          schedule.IsActive,
		CreatedAt:         schedule.CreatedAt,
		UpdatedAt:         schedule.UpdatedAt,
	}
}

func (s *appointmentService) appointmentToResponse(appointment *models.Appointment) *models.AppointmentResponse {
	return &models.AppointmentResponse{
		ID:              appointment.ID,
		StartTime:       appointment.StartTime,
		EndTime:         appointment.EndTime,
		DoctorID:        appointment.DoctorID,
		PatientID:       appointment.PatientID,
		Title:           appointment.Title,
		Status:          appointment.Status,
		AppointmentType: appointment.AppointmentType,
		MeetingLink:     appointment.MeetingLink,
		MeetingID:       appointment.MeetingID,
		PatientNotes:    appointment.PatientNotes,
		DoctorNotes:     appointment.DoctorNotes,
		CreatedAt:       appointment.CreatedAt,
		UpdatedAt:       appointment.UpdatedAt,
	}
}

func (s *appointmentService) exceptionToResponse(exception *models.ScheduleException) *models.ExceptionResponse {
	return &models.ExceptionResponse{
		ID:              exception.ID,
		DoctorID:        exception.DoctorID,
		Date:            exception.Date,
		Type:            exception.Type,
		CustomStartTime: exception.CustomStartTime,
		CustomEndTime:   exception.CustomEndTime,
		Reason:          exception.Reason,
		CreatedAt:       exception.CreatedAt,
	}
}

func (s *appointmentService) UpdateSchedule(doctorID, scheduleID uuid.UUID, req *models.UpdateScheduleRequest) (*models.ScheduleResponse, error) {
	schedule, err := s.repo.GetScheduleByID(scheduleID)
	if err != nil {
		return nil, fmt.Errorf("schedule not found: %w", err)
	}

	if schedule.DoctorID != doctorID {
		return nil, errors.New("schedule doesn't belong to this doctor")
	}

	// Сохраняем оригинальные значения для проверки конфликтов
	originalSchedule := *schedule

	// Обновляем только переданные поля
	if req.Name != nil {
		schedule.Name = *req.Name
	}
	if req.WorkDays != nil {
		schedule.SetWorkDays(*req.WorkDays)
	}
	if req.StartTime != nil {
		schedule.StartTime = *req.StartTime
	}
	if req.EndTime != nil {
		schedule.EndTime = *req.EndTime
	}
	if req.BreakStart != nil {
		schedule.BreakStart = req.BreakStart
	}
	if req.BreakEnd != nil {
		schedule.BreakEnd = req.BreakEnd
	}
	if req.SlotDuration != nil {
		schedule.SlotDuration = *req.SlotDuration
	}
	if req.SlotTitle != nil {
		schedule.SlotTitle = *req.SlotTitle
	}
	if req.AppointmentFormat != nil {
		schedule.AppointmentFormat = *req.AppointmentFormat
	}

	// Проверяем изменение типа записи для обновления слотов
	appointmentFormatChanged := req.AppointmentFormat != nil && originalSchedule.AppointmentFormat != schedule.AppointmentFormat

	// Проверяем конфликты если расписание активно и изменились критичные поля
	if schedule.IsActive {
		timeChanged := originalSchedule.StartTime != schedule.StartTime || originalSchedule.EndTime != schedule.EndTime
		daysChanged := !s.workDaysEqual(originalSchedule.WorkDays(), schedule.WorkDays())

		if timeChanged || daysChanged {
			s.logInfo("Checking conflicts after schedule update", map[string]interface{}{
				"doctorID":    doctorID.String(),
				"scheduleID":  scheduleID.String(),
				"timeChanged": timeChanged,
				"daysChanged": daysChanged,
			})

			if err := s.checkScheduleConflictsForExisting(doctorID, schedule); err != nil {
				s.logError("Cannot update schedule due to conflicts", map[string]interface{}{
					"doctorID":   doctorID.String(),
					"scheduleID": scheduleID.String(),
					"error":      err.Error(),
				})
				return nil, err
			}
		}
	}

	if err := s.repo.UpdateSchedule(schedule); err != nil {
		return nil, fmt.Errorf("failed to update schedule: %w", err)
	}

	// Если изменился формат записи - обновляем все доступные слоты
	if appointmentFormatChanged {
		s.logInfo("Appointment format changed - updating available slots", map[string]interface{}{
			"doctorID":             doctorID.String(),
			"scheduleID":           scheduleID.String(),
			"oldAppointmentFormat": originalSchedule.AppointmentFormat,
			"newAppointmentFormat": schedule.AppointmentFormat,
		})

		if err := s.repo.UpdateScheduleAppointmentType(scheduleID, schedule.AppointmentFormat); err != nil {
			s.logError("Failed to update appointment type in slots", map[string]interface{}{
				"doctorID":             doctorID.String(),
				"scheduleID":           scheduleID.String(),
				"newAppointmentFormat": schedule.AppointmentFormat,
				"error":                err.Error(),
			})
			// Возвращаем ошибку, но расписание уже сохранено - это важно отметить
			return nil, fmt.Errorf("schedule updated but failed to update slots appointment type: %w", err)
		}

		s.logInfo("Available slots appointment type updated successfully", map[string]interface{}{
			"doctorID":             doctorID.String(),
			"scheduleID":           scheduleID.String(),
			"newAppointmentFormat": schedule.AppointmentFormat,
		})
	}

	return s.scheduleToResponse(schedule), nil
}

// workDaysEqual сравнивает два массива рабочих дней
func (s *appointmentService) workDaysEqual(days1, days2 []int) bool {
	if len(days1) != len(days2) {
		return false
	}

	// Создаем мапы для сравнения
	map1 := make(map[int]bool)
	map2 := make(map[int]bool)

	for _, day := range days1 {
		map1[day] = true
	}
	for _, day := range days2 {
		map2[day] = true
	}

	for day := range map1 {
		if !map2[day] {
			return false
		}
	}

	return true
}

func (s *appointmentService) DeleteSchedule(doctorID, scheduleID uuid.UUID) error {
	schedule, err := s.repo.GetScheduleByID(scheduleID)
	if err != nil {
		return fmt.Errorf("schedule not found: %w", err)
	}

	if schedule.DoctorID != doctorID {
		return errors.New("schedule doesn't belong to this doctor")
	}

	s.logInfo("Starting schedule deletion", map[string]interface{}{
		"doctorID":   doctorID.String(),
		"scheduleID": scheduleID.String(),
		"name":       schedule.Name,
	})

	if err := s.repo.DeleteSchedule(scheduleID); err != nil {
		s.logError("Failed to delete schedule", map[string]interface{}{
			"doctorID":   doctorID.String(),
			"scheduleID": scheduleID.String(),
			"error":      err.Error(),
		})
		return fmt.Errorf("failed to delete schedule: %w", err)
	}

	s.logInfo("Schedule and available slots deleted successfully", map[string]interface{}{
		"doctorID":   doctorID.String(),
		"scheduleID": scheduleID.String(),
		"note":       "Booked appointments are preserved for history",
	})

	return nil
}

func (s *appointmentService) ToggleSchedule(doctorID, scheduleID uuid.UUID, req *models.ToggleScheduleRequest, hasRequestBody bool) (*models.ScheduleResponse, error) {
	schedule, err := s.repo.GetScheduleByID(scheduleID)
	if err != nil {
		return nil, fmt.Errorf("schedule not found: %w", err)
	}

	if schedule.DoctorID != doctorID {
		return nil, errors.New("schedule doesn't belong to this doctor")
	}

	// Определяем желаемое состояние
	var targetIsActive bool
	var actionType string

	if hasRequestBody {
		// Если есть тело запроса - используем переданное значение
		targetIsActive = req.IsActive
		if req.IsActive && !schedule.IsActive {
			actionType = "ACTIVATING_BY_REQUEST"
		} else if !req.IsActive && schedule.IsActive {
			actionType = "DEACTIVATING_BY_REQUEST"
		} else if req.IsActive && schedule.IsActive {
			actionType = "ALREADY_ACTIVE_BY_REQUEST"
		} else {
			actionType = "ALREADY_INACTIVE_BY_REQUEST"
		}
	} else {
		// Если нет тела запроса - переключаем на противоположное
		targetIsActive = !schedule.IsActive
		if targetIsActive {
			actionType = "AUTO_ACTIVATING"
		} else {
			actionType = "AUTO_DEACTIVATING"
		}
	}

	// Детальное логирование для отладки
	s.logInfo("ToggleSchedule request received", map[string]interface{}{
		"doctorID":          doctorID.String(),
		"scheduleID":        scheduleID.String(),
		"scheduleName":      schedule.Name,
		"currentIsActive":   schedule.IsActive,
		"hasRequestBody":    hasRequestBody,
		"requestedIsActive": req.IsActive,
		"targetIsActive":    targetIsActive,
		"actionType":        actionType,
	})

	// Если активируем расписание - деактивируем все остальные
	if targetIsActive && !schedule.IsActive {
		s.logInfo("Activating schedule - deactivating all other schedules", map[string]interface{}{
			"doctorID":   doctorID.String(),
			"scheduleID": scheduleID.String(),
			"name":       schedule.Name,
		})

		// Деактивируем все другие активные расписания
		if err := s.deactivateOtherSchedules(doctorID, scheduleID); err != nil {
			s.logError("Failed to deactivate other schedules", map[string]interface{}{
				"doctorID":   doctorID.String(),
				"scheduleID": scheduleID.String(),
				"error":      err.Error(),
			})
			return nil, fmt.Errorf("failed to deactivate other schedules: %w", err)
		}
	}

	schedule.IsActive = targetIsActive

	if err := s.repo.UpdateSchedule(schedule); err != nil {
		return nil, fmt.Errorf("failed to toggle schedule: %w", err)
	}

	s.logInfo("Schedule toggled successfully", map[string]interface{}{
		"doctorID":      doctorID.String(),
		"scheduleID":    scheduleID.String(),
		"finalIsActive": schedule.IsActive,
		"operation": func() string {
			if schedule.IsActive {
				return "ACTIVATED"
			} else {
				return "DEACTIVATED"
			}
		}(),
	})

	return s.scheduleToResponse(schedule), nil
}

// checkScheduleConflictsForExisting проверяет конфликты для существующего расписания
func (s *appointmentService) checkScheduleConflictsForExisting(doctorID uuid.UUID, schedule *models.DoctorSchedule) error {
	existingSchedules, err := s.repo.GetDoctorSchedules(doctorID)
	if err != nil {
		return fmt.Errorf("failed to get existing schedules: %w", err)
	}

	startTime, err := time.Parse("15:04", schedule.StartTime)
	if err != nil {
		return fmt.Errorf("invalid start time format: %w", err)
	}

	endTime, err := time.Parse("15:04", schedule.EndTime)
	if err != nil {
		return fmt.Errorf("invalid end time format: %w", err)
	}

	for _, existing := range existingSchedules {
		// Пропускаем само расписание и неактивные расписания
		if existing.ID == schedule.ID || !existing.IsActive {
			continue
		}

		existingStart, _ := time.Parse("15:04", existing.StartTime)
		existingEnd, _ := time.Parse("15:04", existing.EndTime)

		// Проверяем пересечение рабочих дней
		if s.hasWorkDayConflict(schedule.WorkDays(), existing.WorkDays()) {
			// Проверяем пересечение времени
			if s.hasTimeConflict(startTime, endTime, existingStart, existingEnd) {
				return fmt.Errorf("schedule conflicts with active schedule '%s' on overlapping work days and times", existing.Name)
			}
		}
	}

	return nil
}

// New method for forcing clean slots of schedule
func (s *appointmentService) DeleteScheduleSlots(doctorID, scheduleID uuid.UUID) error {
	schedule, err := s.repo.GetScheduleByID(scheduleID)
	if err != nil {
		return fmt.Errorf("schedule not found: %w", err)
	}

	if schedule.DoctorID != doctorID {
		return errors.New("schedule doesn't belong to this doctor")
	}

	if err := s.repo.DeleteScheduleSlots(scheduleID); err != nil {
		return fmt.Errorf("failed to delete schedule slots: %w", err)
	}

	return nil
}

func (s *appointmentService) GetGeneratedSlots(doctorID, scheduleID uuid.UUID, startDate, endDate string) (*models.GeneratedSlotsResponse, error) {
	s.logInfo("Getting generated slots", map[string]interface{}{
		"doctorID":   doctorID.String(),
		"scheduleID": scheduleID.String(),
		"startDate":  startDate,
		"endDate":    endDate,
	})

	// Проверяем расписание и права доступа
	schedule, err := s.repo.GetScheduleByID(scheduleID)
	if err != nil {
		s.logError("Schedule not found", map[string]interface{}{
			"doctorID":   doctorID.String(),
			"scheduleID": scheduleID.String(),
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("schedule not found: %w", err)
	}

	if schedule.DoctorID != doctorID {
		s.logError("Schedule ownership validation failed", map[string]interface{}{
			"doctorID":         doctorID.String(),
			"scheduleID":       scheduleID.String(),
			"scheduleDoctorID": schedule.DoctorID.String(),
		})
		return nil, errors.New("schedule doesn't belong to this doctor")
	}

	// Парсим даты
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		s.logError("Invalid start date format", map[string]interface{}{
			"doctorID":  doctorID.String(),
			"startDate": startDate,
			"error":     err.Error(),
		})
		return nil, fmt.Errorf("invalid start date: %w", err)
	}

	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		s.logError("Invalid end date format", map[string]interface{}{
			"doctorID": doctorID.String(),
			"endDate":  endDate,
			"error":    err.Error(),
		})
		return nil, fmt.Errorf("invalid end date: %w", err)
	}

	// Добавляем время к концу дня для правильного поиска
	endWithTime := end.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	// Получаем все слоты для расписания в заданном периоде
	appointments, err := s.repo.GetScheduleSlots(scheduleID, start, endWithTime)
	if err != nil {
		s.logError("Failed to get schedule slots from repository", map[string]interface{}{
			"doctorID":   doctorID.String(),
			"scheduleID": scheduleID.String(),
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to get schedule slots: %w", err)
	}

	// Преобразуем слоты в детальную информацию
	slots := make([]models.GeneratedSlotDetail, len(appointments))
	var availableCount, bookedCount, canceledCount int

	for i, appointment := range appointments {
		duration := int(appointment.EndTime.Sub(appointment.StartTime).Minutes())

		var bookedAt *time.Time
		if appointment.Status == "booked" && !appointment.UpdatedAt.IsZero() {
			bookedAt = &appointment.UpdatedAt
		}

		slots[i] = models.GeneratedSlotDetail{
			ID:              appointment.ID,
			StartTime:       appointment.StartTime,
			EndTime:         appointment.EndTime,
			Duration:        duration,
			Status:          appointment.Status,
			AppointmentType: appointment.AppointmentType,
			Title:           appointment.Title,
			PatientID:       appointment.PatientID,
			PatientNotes:    appointment.PatientNotes,
			BookedAt:        bookedAt,
		}

		// Подсчет статистики
		switch appointment.Status {
		case "available":
			availableCount++
		case "booked":
			bookedCount++
		case "canceled":
			canceledCount++
		}
	}

	// Подготавливаем метаданные расписания
	scheduleMetadata := models.ScheduleMetadata{
		ID:                schedule.ID,
		Name:              schedule.Name,
		WorkDays:          schedule.WorkDays(),
		StartTime:         schedule.StartTime,
		EndTime:           schedule.EndTime,
		BreakStart:        schedule.BreakStart,
		BreakEnd:          schedule.BreakEnd,
		SlotDuration:      schedule.SlotDuration,
		SlotTitle:         schedule.SlotTitle,
		AppointmentFormat: schedule.AppointmentFormat,
		IsActive:          schedule.IsActive,
	}

	// Подсчитываем количество дней
	days := int(end.Sub(start).Hours()/24) + 1

	// Подготавливаем период
	period := models.Period{
		StartDate: startDate,
		EndDate:   endDate,
		Days:      days,
	}

	// Подготавливаем сводку
	summary := models.SlotsSummary{
		TotalSlots:     len(slots),
		AvailableSlots: availableCount,
		BookedSlots:    bookedCount,
		CanceledSlots:  canceledCount,
	}

	s.logInfo("Generated slots retrieved successfully", map[string]interface{}{
		"doctorID":       doctorID.String(),
		"scheduleID":     scheduleID.String(),
		"totalSlots":     len(slots),
		"availableSlots": availableCount,
		"bookedSlots":    bookedCount,
		"canceledSlots":  canceledCount,
	})

	return &models.GeneratedSlotsResponse{
		Schedule: scheduleMetadata,
		Period:   period,
		Slots:    slots,
		Summary:  summary,
	}, nil
}

func (s *appointmentService) GetDoctorAppointmentByID(doctorID, appointmentID uuid.UUID) (*models.AppointmentResponse, error) {
	appointment, err := s.repo.GetAppointmentByID(appointmentID)
	if err != nil {
		return nil, fmt.Errorf("appointment not found: %w", err)
	}

	if appointment.DoctorID != doctorID {
		return nil, errors.New("appointment doesn't belong to this doctor")
	}

	return s.appointmentToResponse(appointment), nil
}

func (s *appointmentService) GetPatientAppointmentByID(patientID, appointmentID uuid.UUID) (*models.AppointmentResponse, error) {
	appointment, err := s.repo.GetAppointmentByID(appointmentID)
	if err != nil {
		return nil, fmt.Errorf("appointment not found: %w", err)
	}

	if appointment.PatientID == nil || *appointment.PatientID != patientID {
		return nil, errors.New("appointment doesn't belong to this patient")
	}

	return s.appointmentToResponse(appointment), nil
}
