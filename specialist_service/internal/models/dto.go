// Модель запроса для создания врача
package models

import (
	"github.com/google/uuid"
	"github.com/lib/pq"
)

// DoctorCreateRequest структура запроса для создания врача
type DoctorCreateRequest struct {
	UserID      uuid.UUID `json:"user_id" binding:"required"`
	FirstName   string    `json:"first_name" binding:"required"`
	MiddleName  string    `json:"middle_name"`
	LastName    string    `json:"last_name" binding:"required"`
	Description string    `json:"description"`
	Email       string    `json:"email" binding:"required,email"`
	Phone       string    `json:"phone" binding:"required"`
	Roles       []string  `json:"roles" binding:"required"`
}

// DoctorResponse структура ответа с данными врача
type DoctorResponse struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	FirstName   string    `json:"first_name"`
	MiddleName  string    `json:"middle_name"`
	LastName    string    `json:"last_name"`
	Description string    `json:"description"`
	Email       string    `json:"email"`
	Phone       string    `json:"phone"`
	Roles       []string  `json:"roles"`
}

// ToDoctor преобразует запрос в модель Doctor
func (r *DoctorCreateRequest) ToDoctor() *Doctor {
	return &Doctor{
		UserID:      r.UserID,
		FirstName:   r.FirstName,
		MiddleName:  r.MiddleName,
		LastName:    r.LastName,
		Description: r.Description,
		Email:       r.Email,
		Phone:       r.Phone,
		Roles:       pq.StringArray(r.Roles),
	}
}

// ToDoctorResponse преобразует модель Doctor в ответ
func (d *Doctor) ToDoctorResponse() *DoctorResponse {
	return &DoctorResponse{
		ID:          d.ID,
		UserID:      d.UserID,
		FirstName:   d.FirstName,
		MiddleName:  d.MiddleName,
		LastName:    d.LastName,
		Description: d.Description,
		Email:       d.Email,
		Phone:       d.Phone,
		Roles:       []string(d.Roles),
	}
}
