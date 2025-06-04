package service

import (
	"errors"
	"fmt"
	"time"

	"appointment_service/internal/models"
	"appointment_service/internal/repository"

	"github.com/google/uuid"
)

// AppointmentService - интерфейс сервиса
type AppointmentService interface {
	// Schedules
	CreateSchedule(doctorID uuid.UUID, req *models.CreateScheduleRequest) (*models.ScheduleResponse, error)
	GetDoctorSchedules(doctorID uuid.UUID) ([]*models.ScheduleResponse, error)
	UpdateSchedule(doctorID, scheduleID uuid.UUID, req *models.UpdateScheduleRequest) (*models.ScheduleResponse, error)
	DeleteSchedule(doctorID, scheduleID uuid.UUID) error
	ToggleSchedule(doctorID, scheduleID uuid.UUID, req *models.ToggleScheduleRequest) (*models.ScheduleResponse, error)
	GenerateSlots(doctorID, scheduleID uuid.UUID, req *models.GenerateSlotsRequest) error

	// Appointments
	GetAvailableSlots(doctorID uuid.UUID, date string) ([]*models.AvailableSlot, error)
	BookAppointment(patientID, appointmentID uuid.UUID, req *models.BookAppointmentRequest) (*models.AppointmentResponse, error)
	CancelAppointment(appointmentID uuid.UUID) error
	GetDoctorAppointments(doctorID uuid.UUID, date string) ([]*models.AppointmentResponse, error)
	GetPatientAppointments(patientID uuid.UUID, date string) ([]*models.AppointmentResponse, error)

	// Exceptions
	AddException(doctorID uuid.UUID, req *models.AddExceptionRequest) (*models.ExceptionResponse, error)
	GetDoctorExceptions(doctorID uuid.UUID, startDate, endDate string) ([]*models.ExceptionResponse, error)
}

// appointmentService - реализация сервиса
type appointmentService struct {
	repo repository.AppointmentRepository
}

// NewAppointmentService - создание нового сервиса
func NewAppointmentService(repo repository.AppointmentRepository) AppointmentService {
	return &appointmentService{repo: repo}
}

// === SCHEDULES ===

func (s *appointmentService) CreateSchedule(doctorID uuid.UUID, req *models.CreateScheduleRequest) (*models.ScheduleResponse, error) {
	// Если это основное расписание, деактивируем другие основные
	if req.IsDefault {
		schedules, _ := s.repo.GetDoctorSchedules(doctorID)
		for _, schedule := range schedules {
			if schedule.IsDefault {
				schedule.IsDefault = false
				s.repo.UpdateSchedule(schedule)
			}
		}
	}

	schedule := &models.DoctorSchedule{
		DoctorID:     doctorID,
		Name:         req.Name,
		WorkDays:     req.WorkDays,
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
		BreakStart:   req.BreakStart,
		BreakEnd:     req.BreakEnd,
		SlotDuration: req.SlotDuration,
		SlotTitle:    req.SlotTitle,
		IsActive:     true,
		IsDefault:    req.IsDefault,
	}

	if err := s.repo.CreateSchedule(schedule); err != nil {
		return nil, fmt.Errorf("failed to create schedule: %w", err)
	}

	return s.scheduleToResponse(schedule), nil
}

func (s *appointmentService) GetDoctorSchedules(doctorID uuid.UUID) ([]*models.ScheduleResponse, error) {
	schedules, err := s.repo.GetDoctorSchedules(doctorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get schedules: %w", err)
	}

	responses := make([]*models.ScheduleResponse, len(schedules))
	for i, schedule := range schedules {
		responses[i] = s.scheduleToResponse(schedule)
	}

	return responses, nil
}

func (s *appointmentService) GenerateSlots(doctorID, scheduleID uuid.UUID, req *models.GenerateSlotsRequest) error {
	schedule, err := s.repo.GetScheduleByID(scheduleID)
	if err != nil {
		return fmt.Errorf("schedule not found: %w", err)
	}

	if schedule.DoctorID != doctorID {
		return errors.New("schedule doesn't belong to this doctor")
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return fmt.Errorf("invalid start date: %w", err)
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return fmt.Errorf("invalid end date: %w", err)
	}

	// Получаем исключения для периода
	exceptions, _ := s.repo.GetDoctorExceptions(doctorID, startDate, endDate)
	exceptionMap := make(map[string]*models.ScheduleException)
	for _, ex := range exceptions {
		dateStr := ex.Date.Format("2006-01-02")
		exceptionMap[dateStr] = ex
	}

	// Генерируем слоты
	for date := startDate; !date.After(endDate); date = date.AddDate(0, 0, 1) {
		dateStr := date.Format("2006-01-02")
		weekday := int(date.Weekday())
		if weekday == 0 {
			weekday = 7 // Воскресенье = 7
		}

		// Проверяем исключения
		if exception, exists := exceptionMap[dateStr]; exists {
			if exception.Type == "day_off" {
				continue // Пропускаем выходной
			}
			// Для кастомных часов используем их вместо обычного расписания
			if exception.Type == "custom_hours" && exception.CustomStartTime != nil && exception.CustomEndTime != nil {
				s.generateSlotsForDay(date, *exception.CustomStartTime, *exception.CustomEndTime, nil, nil, schedule)
				continue
			}
		}

		// Проверяем рабочие дни
		isWorkDay := false
		for _, workDay := range schedule.WorkDays {
			if workDay == weekday {
				isWorkDay = true
				break
			}
		}

		if !isWorkDay {
			continue
		}

		// Генерируем слоты для обычного дня
		s.generateSlotsForDay(date, schedule.StartTime, schedule.EndTime, schedule.BreakStart, schedule.BreakEnd, schedule)
	}

	return nil
}

func (s *appointmentService) generateSlotsForDay(date time.Time, startTime, endTime string, breakStart, breakEnd *string, schedule *models.DoctorSchedule) error {
	location := time.Local

	// Парсим время начала и конца
	start, err := time.ParseInLocation("2006-01-02 15:04", date.Format("2006-01-02")+" "+startTime, location)
	if err != nil {
		return err
	}

	end, err := time.ParseInLocation("2006-01-02 15:04", date.Format("2006-01-02")+" "+endTime, location)
	if err != nil {
		return err
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

		// Создаем слот
		appointment := &models.Appointment{
			StartTime:  current,
			EndTime:    slotEnd,
			DoctorID:   schedule.DoctorID,
			Title:      schedule.SlotTitle,
			Status:     "available",
			ScheduleID: &schedule.ID,
		}

		s.repo.CreateAppointment(appointment)
		current = slotEnd
	}

	return nil
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
		appointmentType = "offline"
	}

	appointment.Book(patientID, appointmentType, req.PatientNotes)

	if err := s.repo.UpdateAppointment(appointment); err != nil {
		return nil, fmt.Errorf("failed to book appointment: %w", err)
	}

	return s.appointmentToResponse(appointment), nil
}

func (s *appointmentService) CancelAppointment(appointmentID uuid.UUID) error {
	appointment, err := s.repo.GetAppointmentByID(appointmentID)
	if err != nil {
		return fmt.Errorf("appointment not found: %w", err)
	}

	appointment.Cancel()

	if err := s.repo.UpdateAppointment(appointment); err != nil {
		return fmt.Errorf("failed to cancel appointment: %w", err)
	}

	return nil
}

func (s *appointmentService) GetDoctorAppointments(doctorID uuid.UUID, date string) ([]*models.AppointmentResponse, error) {
	startDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}
	endDate := startDate.AddDate(0, 0, 1)

	appointments, err := s.repo.GetDoctorAppointments(doctorID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get doctor appointments: %w", err)
	}

	responses := make([]*models.AppointmentResponse, len(appointments))
	for i, appointment := range appointments {
		responses[i] = s.appointmentToResponse(appointment)
	}

	return responses, nil
}

func (s *appointmentService) GetPatientAppointments(patientID uuid.UUID, date string) ([]*models.AppointmentResponse, error) {
	startDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}
	endDate := startDate.AddDate(0, 0, 1)

	appointments, err := s.repo.GetPatientAppointments(patientID, startDate, endDate)
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
		ID:           schedule.ID,
		DoctorID:     schedule.DoctorID,
		Name:         schedule.Name,
		WorkDays:     schedule.WorkDays,
		StartTime:    schedule.StartTime,
		EndTime:      schedule.EndTime,
		BreakStart:   schedule.BreakStart,
		BreakEnd:     schedule.BreakEnd,
		SlotDuration: schedule.SlotDuration,
		SlotTitle:    schedule.SlotTitle,
		IsActive:     schedule.IsActive,
		IsDefault:    schedule.IsDefault,
		CreatedAt:    schedule.CreatedAt,
		UpdatedAt:    schedule.UpdatedAt,
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

	// Обновляем только переданные поля
	if req.Name != nil {
		schedule.Name = *req.Name
	}
	if req.WorkDays != nil {
		schedule.WorkDays = *req.WorkDays
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
	if req.IsDefault != nil {
		// Если устанавливаем как основное, деактивируем другие основные
		if *req.IsDefault {
			schedules, _ := s.repo.GetDoctorSchedules(doctorID)
			for _, otherSchedule := range schedules {
				if otherSchedule.IsDefault && otherSchedule.ID != scheduleID {
					otherSchedule.IsDefault = false
					s.repo.UpdateSchedule(otherSchedule)
				}
			}
		}
		schedule.IsDefault = *req.IsDefault
	}

	if err := s.repo.UpdateSchedule(schedule); err != nil {
		return nil, fmt.Errorf("failed to update schedule: %w", err)
	}

	return s.scheduleToResponse(schedule), nil
}

func (s *appointmentService) DeleteSchedule(doctorID, scheduleID uuid.UUID) error {
	schedule, err := s.repo.GetScheduleByID(scheduleID)
	if err != nil {
		return fmt.Errorf("schedule not found: %w", err)
	}

	if schedule.DoctorID != doctorID {
		return errors.New("schedule doesn't belong to this doctor")
	}

	if err := s.repo.DeleteSchedule(scheduleID); err != nil {
		return fmt.Errorf("failed to delete schedule: %w", err)
	}

	return nil
}

func (s *appointmentService) ToggleSchedule(doctorID, scheduleID uuid.UUID, req *models.ToggleScheduleRequest) (*models.ScheduleResponse, error) {
	schedule, err := s.repo.GetScheduleByID(scheduleID)
	if err != nil {
		return nil, fmt.Errorf("schedule not found: %w", err)
	}

	if schedule.DoctorID != doctorID {
		return nil, errors.New("schedule doesn't belong to this doctor")
	}

	schedule.IsActive = req.IsActive

	if err := s.repo.UpdateSchedule(schedule); err != nil {
		return nil, fmt.Errorf("failed to toggle schedule: %w", err)
	}

	return s.scheduleToResponse(schedule), nil
}
