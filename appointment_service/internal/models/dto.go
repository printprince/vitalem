package models

import (
	"time"

	"github.com/google/uuid"
)

// === SCHEDULE DTOs ===

// CreateScheduleRequest - создание расписания врача
type CreateScheduleRequest struct {
	Name              string  `json:"name" validate:"required,min=1,max=255"`                           // "Основное расписание"
	WorkDays          []int   `json:"work_days" validate:"required,min=1,max=7,dive,min=1,max=7"`       // [1,2,3,4,5]
	StartTime         string  `json:"start_time" validate:"required,len=5"`                             // "09:00"
	EndTime           string  `json:"end_time" validate:"required,len=5"`                               // "18:00"
	BreakStart        *string `json:"break_start,omitempty" validate:"omitempty,len=5"`                 // "12:00"
	BreakEnd          *string `json:"break_end,omitempty" validate:"omitempty,len=5"`                   // "13:00"
	SlotDuration      int64   `json:"slot_duration" validate:"required,min=15,max=180"`                 // 30
	SlotTitle         string  `json:"slot_title" validate:"max=255"`                                    // "Консультация"
	AppointmentFormat string  `json:"appointment_format" validate:"required,oneof=offline online both"` // "offline", "online", "both"
}

// ScheduleResponse - ответ с расписанием
type ScheduleResponse struct {
	ID                uuid.UUID `json:"id"`
	DoctorID          uuid.UUID `json:"doctor_id"`
	Name              string    `json:"name"`
	WorkDays          []int     `json:"work_days"`
	StartTime         string    `json:"start_time"`
	EndTime           string    `json:"end_time"`
	BreakStart        *string   `json:"break_start,omitempty"`
	BreakEnd          *string   `json:"break_end,omitempty"`
	SlotDuration      int64     `json:"slot_duration"`
	SlotTitle         string    `json:"slot_title"`
	AppointmentFormat string    `json:"appointment_format"`
	IsActive          bool      `json:"is_active"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// UpdateScheduleRequest - обновление расписания врача
type UpdateScheduleRequest struct {
	Name              *string `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	WorkDays          *[]int  `json:"work_days,omitempty" validate:"omitempty,min=1,max=7,dive,min=1,max=7"`
	StartTime         *string `json:"start_time,omitempty" validate:"omitempty,len=5"`
	EndTime           *string `json:"end_time,omitempty" validate:"omitempty,len=5"`
	BreakStart        *string `json:"break_start,omitempty" validate:"omitempty,len=5"`
	BreakEnd          *string `json:"break_end,omitempty" validate:"omitempty,len=5"`
	SlotDuration      *int64  `json:"slot_duration,omitempty" validate:"omitempty,min=15,max=180"`
	SlotTitle         *string `json:"slot_title,omitempty" validate:"omitempty,max=255"`
	AppointmentFormat *string `json:"appointment_format,omitempty" validate:"omitempty,oneof=offline online both"`
}

// ToggleScheduleRequest - активация/деактивация расписания
type ToggleScheduleRequest struct {
	IsActive bool `json:"is_active"`
}

// GenerateSlotsRequest - генерация слотов
type GenerateSlotsRequest struct {
	StartDate string `json:"start_date" validate:"required,len=10"` // "2024-06-01"
	EndDate   string `json:"end_date" validate:"required,len=10"`   // "2024-06-30"
}

// GenerateSlotsResponse - ответ генерации слотов
type GenerateSlotsResponse struct {
	SlotsCreated int    `json:"slots_created"`
	Message      string `json:"message"`
}

// === APPOINTMENT DTOs ===

// BookAppointmentRequest - бронирование записи
type BookAppointmentRequest struct {
	AppointmentType string `json:"appointment_type" validate:"omitempty,oneof=offline online"` // "offline", "online"
	PatientNotes    string `json:"patient_notes" validate:"max=1000"`                          // "Болит голова"
}

// AppointmentResponse - ответ с записью
type AppointmentResponse struct {
	ID        uuid.UUID `json:"id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`

	DoctorID  uuid.UUID  `json:"doctor_id"`
	PatientID *uuid.UUID `json:"patient_id,omitempty"`

	Title           string `json:"title"`
	Status          string `json:"status"`
	AppointmentType string `json:"appointment_type"`

	// Онлайн встреча (только для онлайн записей)
	MeetingLink *string `json:"meeting_link,omitempty"`
	MeetingID   *string `json:"meeting_id,omitempty"`

	PatientNotes string `json:"patient_notes"`
	DoctorNotes  string `json:"doctor_notes"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AvailableSlot - доступный слот для пациента
type AvailableSlot struct {
	ID              uuid.UUID `json:"id"`
	StartTime       time.Time `json:"start_time"`
	EndTime         time.Time `json:"end_time"`
	Duration        int       `json:"duration_minutes"`
	Title           string    `json:"title"`
	AppointmentType string    `json:"appointment_type"` // offline, online
}

// === EXCEPTION DTOs ===

// AddExceptionRequest - добавление исключения
type AddExceptionRequest struct {
	Date            string  `json:"date" validate:"required,len=10"`                        // "2024-06-15"
	Type            string  `json:"type" validate:"required,oneof=day_off custom_hours"`    // "day_off", "custom_hours"
	CustomStartTime *string `json:"custom_start_time,omitempty" validate:"omitempty,len=5"` // "10:00"
	CustomEndTime   *string `json:"custom_end_time,omitempty" validate:"omitempty,len=5"`   // "16:00"
	Reason          string  `json:"reason" validate:"max=255"`                              // "Отпуск"
}

// ExceptionResponse - ответ с исключением
type ExceptionResponse struct {
	ID              uuid.UUID `json:"id"`
	DoctorID        uuid.UUID `json:"doctor_id"`
	Date            time.Time `json:"date"`
	Type            string    `json:"type"`
	CustomStartTime *string   `json:"custom_start_time,omitempty"`
	CustomEndTime   *string   `json:"custom_end_time,omitempty"`
	Reason          string    `json:"reason"`
	CreatedAt       time.Time `json:"created_at"`
}

// === COMMON RESPONSES ===

// APIResponse - стандартный ответ API
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// PaginatedResponse - ответ с пагинацией
type PaginatedResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Total   int64       `json:"total"`
	Page    int         `json:"page"`
	Limit   int         `json:"limit"`
}

// GeneratedSlotsRequest - запрос для получения сгенерированных слотов
type GeneratedSlotsRequest struct {
	StartDate string `json:"start_date" validate:"required,len=10"` // "2024-06-01"
	EndDate   string `json:"end_date" validate:"required,len=10"`   // "2024-06-30"
}

// GeneratedSlotDetail - детальная информация о сгенерированном слоте
type GeneratedSlotDetail struct {
	ID              uuid.UUID `json:"id"`
	StartTime       time.Time `json:"start_time"`
	EndTime         time.Time `json:"end_time"`
	Duration        int       `json:"duration_minutes"`
	Status          string    `json:"status"`           // "available", "booked", "canceled"
	AppointmentType string    `json:"appointment_type"` // "offline", "online"
	Title           string    `json:"title"`

	// Информация о пациенте, если слот забронирован
	PatientID    *uuid.UUID `json:"patient_id,omitempty"`
	PatientNotes string     `json:"patient_notes,omitempty"`
	BookedAt     *time.Time `json:"booked_at,omitempty"`
}

// ScheduleMetadata - метаданные расписания для сгенерированных слотов
type ScheduleMetadata struct {
	ID                uuid.UUID `json:"id"`
	Name              string    `json:"name"`
	WorkDays          []int     `json:"work_days"`
	StartTime         string    `json:"start_time"`
	EndTime           string    `json:"end_time"`
	BreakStart        *string   `json:"break_start,omitempty"`
	BreakEnd          *string   `json:"break_end,omitempty"`
	SlotDuration      int64     `json:"slot_duration"`
	SlotTitle         string    `json:"slot_title"`
	AppointmentFormat string    `json:"appointment_format"`
	IsActive          bool      `json:"is_active"`
}

// GeneratedSlotsResponse - ответ с детальной информацией о сгенерированных слотах
type GeneratedSlotsResponse struct {
	Schedule ScheduleMetadata      `json:"schedule"`
	Period   Period                `json:"period"`
	Slots    []GeneratedSlotDetail `json:"slots"`
	Summary  SlotsSummary          `json:"summary"`
}

// Period - период генерации слотов
type Period struct {
	StartDate string `json:"start_date"` // "2024-06-01"
	EndDate   string `json:"end_date"`   // "2024-06-30"
	Days      int    `json:"days"`       // Количество дней в периоде
}

// SlotsSummary - сводка по слотам
type SlotsSummary struct {
	TotalSlots     int `json:"total_slots"`
	AvailableSlots int `json:"available_slots"`
	BookedSlots    int `json:"booked_slots"`
	CanceledSlots  int `json:"canceled_slots"`
}
