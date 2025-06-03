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

// PatientCreateRequest структура запроса для создания пациента
type PatientCreateRequest struct {
	UserID              uuid.UUID `json:"user_id" binding:"required"`
	IIN                 *string   `json:"iin"`
	Name                string    `json:"name"`
	Surname             string    `json:"surname"`
	DateOfBirth         Date      `json:"date_of_birth"`
	Gender              string    `json:"gender"`
	Email               string    `json:"email" binding:"required,email"`
	Phone               string    `json:"phone"`
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

// PatientResponse структура ответа с данными пациента
type PatientResponse struct {
	ID                  uuid.UUID `json:"id"`
	UserID              uuid.UUID `json:"user_id"`
	IIN                 *string   `json:"iin"`
	Name                string    `json:"name"`
	Surname             string    `json:"surname"`
	DateOfBirth         time.Time `json:"date_of_birth"`
	Gender              string    `json:"gender"`
	Email               string    `json:"email"`
	Phone               string    `json:"phone"`
	Height              float64   `json:"height"`
	Weight              float64   `json:"weight"`
	PhysActivity        string    `json:"phys_activity"`
	Diagnoses           []string  `json:"diagnoses"`
	AdditionalDiagnoses []string  `json:"additional_diagnoses"`
	Allergens           []string  `json:"allergens"`
	AdditionalAllergens []string  `json:"additional_allergens"`
	Diet                []string  `json:"diet"`
	AdditionalDiets     []string  `json:"additional_diets"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// ToPatient преобразует запрос в модель Patient
func (r *PatientCreateRequest) ToPatient() *Patient {
	return &Patient{
		UserID:              r.UserID,
		IIN:                 r.IIN,
		Name:                r.Name,
		Surname:             r.Surname,
		DateOfBirth:         r.DateOfBirth.Time,
		Gender:              r.Gender,
		Email:               r.Email,
		Phone:               r.Phone,
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

// ToPatientResponse преобразует модель Patient в ответ
func (p *Patient) ToPatientResponse() *PatientResponse {
	return &PatientResponse{
		ID:                  p.ID,
		UserID:              p.UserID,
		IIN:                 p.IIN,
		Name:                p.Name,
		Surname:             p.Surname,
		DateOfBirth:         p.DateOfBirth,
		Gender:              p.Gender,
		Email:               p.Email,
		Phone:               p.Phone,
		Height:              p.Height,
		Weight:              p.Weight,
		PhysActivity:        p.PhysActivity,
		Diagnoses:           []string(p.Diagnoses),
		AdditionalDiagnoses: []string(p.AdditionalDiagnoses),
		Allergens:           []string(p.Allergens),
		AdditionalAllergens: []string(p.AdditionalAllergens),
		Diet:                []string(p.Diet),
		AdditionalDiets:     []string(p.AdditionalDiets),
		CreatedAt:           p.CreatedAt,
		UpdatedAt:           p.UpdatedAt,
	}
}
