package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Appointment - запись к врачу
type Appointment struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	StartTime time.Time `gorm:"not null;index" json:"start_time"`
	EndTime   time.Time `gorm:"not null;index" json:"end_time"`

	// Участники
	DoctorID  uuid.UUID  `gorm:"type:uuid;not null;index" json:"doctor_id"`
	PatientID *uuid.UUID `gorm:"type:uuid;index" json:"patient_id,omitempty"`

	// Информация о записи
	Title           string `gorm:"type:varchar(255);not null" json:"title"`                     // "Консультация терапевта"
	Status          string `gorm:"type:varchar(20);not null;default:'available'" json:"status"` // available, booked, completed, canceled
	AppointmentType string `gorm:"type:varchar(10);default:'offline'" json:"appointment_type"`  // offline, online

	// Ссылка на онлайн встречу (только для онлайн записей)
	MeetingLink *string `gorm:"type:text" json:"meeting_link,omitempty"`
	MeetingID   *string `gorm:"type:varchar(100)" json:"meeting_id,omitempty"`

	// Заметки
	PatientNotes string `gorm:"type:text" json:"patient_notes"` // Жалобы пациента
	DoctorNotes  string `gorm:"type:text" json:"doctor_notes"`  // Заметки врача

	// Связь с расписанием
	ScheduleID *uuid.UUID `gorm:"type:uuid;index" json:"schedule_id,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Appointment) TableName() string {
	return "appointments"
}

func (a *Appointment) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

// Методы для работы с записью
func (a *Appointment) IsAvailable() bool {
	return a.Status == "available" && a.PatientID == nil
}

func (a *Appointment) Book(patientID uuid.UUID, appointmentType, notes string) {
	a.PatientID = &patientID
	a.Status = "booked"
	if appointmentType != "" {
		a.AppointmentType = appointmentType
	}
	if notes != "" {
		a.PatientNotes = notes
	}

	// Если это онлайн запись, генерируем заглушку ссылки
	if appointmentType == "online" {
		meetingID := fmt.Sprintf("vitalem-%s", a.ID.String()[:8])
		meetingLink := fmt.Sprintf("https://meet.vitalem.kz/room/%s", meetingID)
		a.MeetingID = &meetingID
		a.MeetingLink = &meetingLink
	}

	a.UpdatedAt = time.Now()
}

func (a *Appointment) Cancel() {
	a.Status = "canceled"
	a.PatientID = nil
	// Очищаем ссылку на встречу при отмене
	a.MeetingLink = nil
	a.MeetingID = nil
	a.UpdatedAt = time.Now()
}

func (a *Appointment) Complete(doctorNotes string) {
	a.Status = "completed"
	if doctorNotes != "" {
		a.DoctorNotes = doctorNotes
	}
	a.UpdatedAt = time.Now()
}
