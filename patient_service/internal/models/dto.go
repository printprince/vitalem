package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Date кастомный тип для гибкого парсинга дат
type Date struct {
	time.Time
}

// UnmarshalJSON парсит дату из разных форматов
func (d *Date) UnmarshalJSON(data []byte) error {
	var dateStr string
	if err := json.Unmarshal(data, &dateStr); err != nil {
		return err
	}

	// Пробуем разные форматы
	formats := []string{
		"2006-01-02",           // 2004-10-27
		"2006-01-02T15:04:05Z", // 2004-10-27T00:00:00Z
		"02.01.2006",           // 27.10.2004
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			d.Time = t
			return nil
		}
	}

	return &time.ParseError{Value: dateStr, Layout: "supported formats: 2006-01-02, 2006-01-02T15:04:05Z, 02.01.2006"}
}

// PatientCreateRequest структура для создания пациента
type PatientCreateRequest struct {
	UserID     uuid.UUID `json:"user_id" validate:"required"`
	FirstName  string    `json:"first_name" validate:"required"`
	MiddleName string    `json:"middle_name"`
	LastName   string    `json:"last_name" validate:"required"`
	Email      string    `json:"email" validate:"required,email"`
	Phone      string    `json:"phone"`
	Address    string    `json:"address"`
	AvatarURL  string    `json:"avatar_url"`
}

// PatientResponse структура для ответа с данными пациента
type PatientResponse struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"user_id"`
	FirstName  string    `json:"first_name"`
	MiddleName string    `json:"middle_name"`
	LastName   string    `json:"last_name"`
	Email      string    `json:"email"`
	Phone      string    `json:"phone"`
	Address    string    `json:"address"`
	AvatarURL  string    `json:"avatar_url"`
	CreatedAt  string    `json:"created_at"`
	UpdatedAt  string    `json:"updated_at"`
}

// ToPatient конвертирует PatientCreateRequest в Patient
func (r *PatientCreateRequest) ToPatient() *Patient {
	return &Patient{
		UserID:     r.UserID,
		FirstName:  r.FirstName,
		MiddleName: r.MiddleName,
		LastName:   r.LastName,
		Email:      r.Email,
		Phone:      r.Phone,
		Address:    r.Address,
		AvatarURL:  r.AvatarURL,
	}
}

// ToPatientResponse конвертирует Patient в PatientResponse
func ToPatientResponse(p *Patient) *PatientResponse {
	return &PatientResponse{
		ID:         p.ID,
		UserID:     p.UserID,
		FirstName:  p.FirstName,
		MiddleName: p.MiddleName,
		LastName:   p.LastName,
		Email:      p.Email,
		Phone:      p.Phone,
		Address:    p.Address,
		AvatarURL:  p.AvatarURL,
		CreatedAt:  p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:  p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
