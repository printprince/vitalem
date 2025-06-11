package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
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
	UserID              uuid.UUID `json:"user_id" validate:"required"`
	FirstName           string    `json:"first_name" validate:"required"`
	MiddleName          string    `json:"middle_name"`
	LastName            string    `json:"last_name" validate:"required"`
	Email               string    `json:"email" validate:"required,email"`
	Phone               string    `json:"phone"`
	Address             string    `json:"address"`
	AvatarURL           string    `json:"avatar_url"`
	IIN                 *string   `json:"iin"`
	DateOfBirth         Date      `json:"date_of_birth"`
	Gender              string    `json:"gender"`
	Height              float64   `json:"height"`
	Weight              float64   `json:"weight"`
	PhysActivity        string    `json:"phys_activity"`
	Diagnoses           []string  `json:"diagnoses"`
	AdditionalDiagnoses []string  `json:"additional_diagnoses"`
	Allergens           []string  `json:"allergens"`
	AdditionalAllergens []string  `json:"additional_allergens"`
	Diet                []string  `json:"diet"`
	AdditionalDiets     []string  `json:"additional_diets"`
}

// PatientResponse структура для ответа с данными пациента
type PatientResponse struct {
	ID                  uuid.UUID `json:"id"`
	UserID              uuid.UUID `json:"user_id"`
	FirstName           string    `json:"first_name"`
	MiddleName          string    `json:"middle_name"`
	LastName            string    `json:"last_name"`
	Email               string    `json:"email"`
	Phone               string    `json:"phone"`
	Address             string    `json:"address"`
	AvatarURL           string    `json:"avatar_url"`
	IIN                 *string   `json:"iin"`
	DateOfBirth         time.Time `json:"date_of_birth"`
	Gender              string    `json:"gender"`
	Height              float64   `json:"height"`
	Weight              float64   `json:"weight"`
	PhysActivity        string    `json:"phys_activity"`
	Diagnoses           []string  `json:"diagnoses"`
	AdditionalDiagnoses []string  `json:"additional_diagnoses"`
	Allergens           []string  `json:"allergens"`
	AdditionalAllergens []string  `json:"additional_allergens"`
	Diet                []string  `json:"diet"`
	AdditionalDiets     []string  `json:"additional_diets"`
	CreatedAt           string    `json:"created_at"`
	UpdatedAt           string    `json:"updated_at"`
}

// ToPatient конвертирует PatientCreateRequest в Patient
func (r *PatientCreateRequest) ToPatient() *Patient {
	return &Patient{
		UserID:              r.UserID,
		FirstName:           r.FirstName,
		MiddleName:          r.MiddleName,
		LastName:            r.LastName,
		Email:               r.Email,
		Phone:               r.Phone,
		Address:             r.Address,
		AvatarURL:           r.AvatarURL,
		IIN:                 r.IIN,
		DateOfBirth:         r.DateOfBirth.Time,
		Gender:              r.Gender,
		Height:              r.Height,
		Weight:              r.Weight,
		PhysActivity:        r.PhysActivity,
		Diagnoses:           pq.StringArray(r.Diagnoses),
		AdditionalDiagnoses: pq.StringArray(r.AdditionalDiagnoses),
		Allergens:           pq.StringArray(r.Allergens),
		AdditionalAllergens: pq.StringArray(r.AdditionalAllergens),
		Diet:                pq.StringArray(r.Diet),
		AdditionalDiets:     pq.StringArray(r.AdditionalDiets),
	}
}

// ToPatientResponse конвертирует Patient в PatientResponse
func ToPatientResponse(p *Patient) *PatientResponse {
	return &PatientResponse{
		ID:                  p.ID,
		UserID:              p.UserID,
		FirstName:           p.FirstName,
		MiddleName:          p.MiddleName,
		LastName:            p.LastName,
		Email:               p.Email,
		Phone:               p.Phone,
		Address:             p.Address,
		AvatarURL:           p.AvatarURL,
		IIN:                 p.IIN,
		DateOfBirth:         p.DateOfBirth,
		Gender:              p.Gender,
		Height:              p.Height,
		Weight:              p.Weight,
		PhysActivity:        p.PhysActivity,
		Diagnoses:           []string(p.Diagnoses),
		AdditionalDiagnoses: []string(p.AdditionalDiagnoses),
		Allergens:           []string(p.Allergens),
		AdditionalAllergens: []string(p.AdditionalAllergens),
		Diet:                []string(p.Diet),
		AdditionalDiets:     []string(p.AdditionalDiets),
		CreatedAt:           p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:           p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
