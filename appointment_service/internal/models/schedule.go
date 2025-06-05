package models

import (
	"database/sql/driver"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// IntArray represents a slice of integers for PostgreSQL array
type IntArray []int

// Scan implements the Scanner interface for database/sql
func (ia *IntArray) Scan(value interface{}) error {
	if value == nil {
		*ia = nil
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return ia.scanBytes(v)
	case string:
		return ia.scanBytes([]byte(v))
	case []interface{}:
		// Handle interface{} slice from some drivers
		*ia = make(IntArray, len(v))
		for i, item := range v {
			if num, ok := item.(int64); ok {
				(*ia)[i] = int(num)
			} else if str, ok := item.(string); ok {
				if num, err := strconv.Atoi(str); err == nil {
					(*ia)[i] = num
				}
			}
		}
		return nil
	default:
		return fmt.Errorf("cannot scan %T into IntArray", value)
	}
}

func (ia *IntArray) scanBytes(src []byte) error {
	str := string(src)
	str = strings.Trim(str, "{}")
	if str == "" {
		*ia = IntArray{}
		return nil
	}

	parts := strings.Split(str, ",")
	*ia = make(IntArray, len(parts))
	for i, part := range parts {
		num, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil {
			return fmt.Errorf("invalid integer in array: %s", part)
		}
		(*ia)[i] = num
	}
	return nil
}

// Value implements the driver Valuer interface
func (ia IntArray) Value() (driver.Value, error) {
	if ia == nil {
		return nil, nil
	}

	if len(ia) == 0 {
		return "{}", nil
	}

	strs := make([]string, len(ia))
	for i, v := range ia {
		strs[i] = strconv.Itoa(v)
	}
	return "{" + strings.Join(strs, ",") + "}", nil
}

// DoctorSchedule - расписание работы врача
type DoctorSchedule struct {
	ID       uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	DoctorID uuid.UUID `gorm:"type:uuid;not null;index" json:"doctor_id"`
	Name     string    `gorm:"type:varchar(255);not null" json:"name"` // "Основное расписание"

	// Дни недели: [1,2,3,4,5] = Пн-Пт
	WorkDays IntArray `gorm:"type:integer[]" json:"work_days"`

	// Время работы
	StartTime string `gorm:"type:varchar(5);not null" json:"start_time"` // "09:00"
	EndTime   string `gorm:"type:varchar(5);not null" json:"end_time"`   // "18:00"

	// Перерыв (опционально)
	BreakStart *string `gorm:"type:varchar(5)" json:"break_start,omitempty"` // "12:00"
	BreakEnd   *string `gorm:"type:varchar(5)" json:"break_end,omitempty"`   // "13:00"

	// Настройки слотов
	SlotDuration int    `gorm:"not null;default:30" json:"slot_duration"` // 30 минут
	SlotTitle    string `gorm:"type:varchar(255)" json:"slot_title"`      // "Консультация"
	IsActive     bool   `gorm:"default:true" json:"is_active"`            // Активно ли расписание
	IsDefault    bool   `gorm:"default:false" json:"is_default"`          // Основное расписание

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (DoctorSchedule) TableName() string {
	return "doctor_schedules"
}

func (s *DoctorSchedule) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// ScheduleException - исключения в расписании (выходные, изменения)
type ScheduleException struct {
	ID       uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	DoctorID uuid.UUID `gorm:"type:uuid;not null;index" json:"doctor_id"`
	Date     time.Time `gorm:"type:date;not null;index" json:"date"`
	Type     string    `gorm:"type:varchar(20);not null" json:"type"` // "day_off", "custom_hours"

	// Для кастомных часов
	CustomStartTime *string `gorm:"type:varchar(5)" json:"custom_start_time,omitempty"`
	CustomEndTime   *string `gorm:"type:varchar(5)" json:"custom_end_time,omitempty"`

	Reason    string    `gorm:"type:varchar(255)" json:"reason"` // "Отпуск"
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (ScheduleException) TableName() string {
	return "schedule_exceptions"
}

func (e *ScheduleException) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}
