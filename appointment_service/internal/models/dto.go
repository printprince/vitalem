package models

import (
	"time"

	"github.com/google/uuid"
)

// === SCHEDULE DTOs ===

// CreateScheduleRequest - создание расписания врача
type CreateScheduleRequest struct {
	Name         string  `json:"name" validate:"required,min=1,max=255"`                     // "Основное расписание"
	WorkDays     []int   `json:"work_days" validate:"required,min=1,max=7,dive,min=1,max=7"` // [1,2,3,4,5]
	StartTime    string  `json:"start_time" validate:"required,len=5"`                       // "09:00"
	EndTime      string  `json:"end_time" validate:"required,len=5"`                         // "18:00"
	BreakStart   *string `json:"break_start,omitempty" validate:"omitempty,len=5"`           // "12:00"
	BreakEnd     *string `json:"break_end,omitempty" validate:"omitempty,len=5"`             // "13:00"
	SlotDuration int     `json:"slot_duration" validate:"required,min=15,max=180"`           // 30
	SlotTitle    string  `json:"slot_title" validate:"max=255"`                              // "Консультация"
	IsDefault    bool    `json:"is_default"`                                                 // Основное расписание
}

// ScheduleResponse - ответ с расписанием
type ScheduleResponse struct {
	ID           uuid.UUID `json:"id"`
	DoctorID     uuid.UUID `json:"doctor_id"`
	Name         string    `json:"name"`
	WorkDays     []int     `json:"work_days"`
	StartTime    string    `json:"start_time"`
	EndTime      string    `json:"end_time"`
	BreakStart   *string   `json:"break_start,omitempty"`
	BreakEnd     *string   `json:"break_end,omitempty"`
	SlotDuration int       `json:"slot_duration"`
	SlotTitle    string    `json:"slot_title"`
	IsActive     bool      `json:"is_active"`
	IsDefault    bool      `json:"is_default"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// GenerateSlotsRequest - генерация слотов
type GenerateSlotsRequest struct {
	StartDate string `json:"start_date" validate:"required,len=10"` // "2024-06-01"
	EndDate   string `json:"end_date" validate:"required,len=10"`   // "2024-06-30"
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

	PatientNotes string `json:"patient_notes"`
	DoctorNotes  string `json:"doctor_notes"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AvailableSlot - доступный слот для пациента
type AvailableSlot struct {
	ID        uuid.UUID `json:"id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Duration  int       `json:"duration_minutes"`
	Title     string    `json:"title"`
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
