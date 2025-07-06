package docs

// ===============================
// SWAGGER MODELS
// ===============================

// UserInfo represents basic user information from JWT
// swagger:model UserInfo
type UserInfo struct {
	// User ID from users table (UUID)
	// in: string
	UserID string `json:"user_id" example:"123e4567-e89b-12d3-a456-426614174000"`

	// User role (admin, doctor, patient)
	// in: string
	Role string `json:"role" example:"patient"`

	// User email
	// in: string
	Email string `json:"email" example:"user@example.com"`
}

// ===============================
// ПАЦИЕНТЫ (PATIENTS)
// ===============================

// PatientCreateRequest represents request for creating/updating patient
// swagger:model PatientCreateRequest
type PatientCreateRequest struct {
	// First name
	// in: string
	FirstName string `json:"first_name" example:"Иван"`

	// Middle name
	// in: string
	MiddleName string `json:"middle_name" example:"Иванович"`

	// Last name
	// in: string
	LastName string `json:"last_name" example:"Иванов"`

	// Email address
	// in: string
	Email string `json:"email" example:"ivan@example.com"`

	// Phone number
	// in: string
	Phone string `json:"phone" example:"+7700123456"`

	// Address
	// in: string
	Address string `json:"address" example:"г. Алматы, ул. Абая 1"`

	// Avatar URL
	// in: string
	AvatarURL string `json:"avatar_url" example:"https://example.com/avatar.jpg"`

	// Individual Identification Number
	// in: string
	IIN string `json:"iin" example:"123456789012"`

	// Date of birth
	// in: string
	DateOfBirth string `json:"date_of_birth" example:"1990-01-01"`

	// Gender (male/female/other)
	// in: string
	Gender string `json:"gender" example:"male"`

	// Height in cm
	// in: number
	Height float64 `json:"height" example:"175.5"`

	// Weight in kg
	// in: number
	Weight float64 `json:"weight" example:"70.2"`

	// Physical activity level
	// in: string
	PhysActivity string `json:"phys_activity" example:"moderate"`

	// List of diagnoses
	// in: array
	Diagnoses []string `json:"diagnoses" example:"['Гипертония', 'Диабет']"`

	// Additional diagnoses
	// in: array
	AdditionalDiagnoses []string `json:"additional_diagnoses"`

	// List of allergens
	// in: array
	Allergens []string `json:"allergens" example:"['Пенициллин', 'Молоко']"`

	// Additional allergens
	// in: array
	AdditionalAllergens []string `json:"additional_allergens"`

	// Diet preferences
	// in: array
	Diet []string `json:"diet" example:"['Безглютеновая', 'Низкосолевая']"`

	// Additional diets
	// in: array
	AdditionalDiets []string `json:"additional_diets"`
}

// PatientResponse represents patient data response
// swagger:model PatientResponse
type PatientResponse struct {
	// Patient ID (primary key in patients table)
	// in: string
	ID string `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`

	// User ID (foreign key to users table)
	// in: string
	UserID string `json:"user_id" example:"123e4567-e89b-12d3-a456-426614174001"`

	// First name
	// in: string
	FirstName string `json:"first_name" example:"Иван"`

	// Middle name
	// in: string
	MiddleName string `json:"middle_name" example:"Иванович"`

	// Last name
	// in: string
	LastName string `json:"last_name" example:"Иванов"`

	// Email address
	// in: string
	Email string `json:"email" example:"ivan@example.com"`

	// Phone number
	// in: string
	Phone string `json:"phone" example:"+7700123456"`

	// Address
	// in: string
	Address string `json:"address" example:"г. Алматы, ул. Абая 1"`

	// Avatar URL
	// in: string
	AvatarURL string `json:"avatar_url" example:"https://example.com/avatar.jpg"`

	// Individual Identification Number
	// in: string
	IIN string `json:"iin" example:"123456789012"`

	// Date of birth
	// in: string
	DateOfBirth string `json:"date_of_birth" example:"1990-01-01T00:00:00Z"`

	// Gender (male/female/other)
	// in: string
	Gender string `json:"gender" example:"male"`

	// Height in cm
	// in: number
	Height float64 `json:"height" example:"175.5"`

	// Weight in kg
	// in: number
	Weight float64 `json:"weight" example:"70.2"`

	// Physical activity level
	// in: string
	PhysActivity string `json:"phys_activity" example:"moderate"`

	// List of diagnoses
	// in: array
	Diagnoses []string `json:"diagnoses"`

	// Additional diagnoses
	// in: array
	AdditionalDiagnoses []string `json:"additional_diagnoses"`

	// List of allergens
	// in: array
	Allergens []string `json:"allergens"`

	// Additional allergens
	// in: array
	AdditionalAllergens []string `json:"additional_allergens"`

	// Diet preferences
	// in: array
	Diet []string `json:"diet"`

	// Additional diets
	// in: array
	AdditionalDiets []string `json:"additional_diets"`

	// Creation timestamp
	// in: string
	CreatedAt string `json:"created_at" example:"2024-01-01T12:00:00Z"`

	// Update timestamp
	// in: string
	UpdatedAt string `json:"updated_at" example:"2024-01-01T12:00:00Z"`
}

// ===============================
// ВРАЧИ (DOCTORS)
// ===============================

// DoctorCreateRequest represents request for creating/updating doctor
// swagger:model DoctorCreateRequest
type DoctorCreateRequest struct {
	// First name
	// in: string
	FirstName string `json:"first_name" example:"Петр"`

	// Last name
	// in: string
	LastName string `json:"last_name" example:"Петров"`

	// Email address
	// in: string
	Email string `json:"email" example:"doctor@example.com"`

	// Phone number
	// in: string
	Phone string `json:"phone" example:"+7700123456"`

	// List of specializations
	// in: array
	Specializations []string `json:"specializations" example:"['Кардиолог', 'Терапевт']"`

	// Biography
	// in: string
	Biography string `json:"biography" example:"Опытный врач с 10-летним стажем"`

	// Avatar URL
	// in: string
	AvatarURL string `json:"avatar_url" example:"https://example.com/doctor.jpg"`

	// Consultation price
	// in: number
	ConsultationPrice float64 `json:"consultation_price" example:"15000"`

	// Years of experience
	// in: integer
	YearsOfExperience int `json:"years_of_experience" example:"10"`

	// Education information
	// in: string
	Education string `json:"education" example:"КазНМУ"`

	// Certificates list
	// in: array
	Certificates []string `json:"certificates"`

	// Working hours
	// in: string
	WorkingHours string `json:"working_hours" example:"9:00-18:00"`
}

// DoctorResponse represents doctor data response
// swagger:model DoctorResponse
type DoctorResponse struct {
	// Doctor ID (primary key in doctors table)
	// in: string
	ID string `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`

	// User ID (foreign key to users table)
	// in: string
	UserID string `json:"user_id" example:"123e4567-e89b-12d3-a456-426614174001"`

	// First name
	// in: string
	FirstName string `json:"first_name" example:"Петр"`

	// Last name
	// in: string
	LastName string `json:"last_name" example:"Петров"`

	// Email address
	// in: string
	Email string `json:"email" example:"doctor@example.com"`

	// Phone number
	// in: string
	Phone string `json:"phone" example:"+7700123456"`

	// List of specializations
	// in: array
	Specializations []string `json:"specializations"`

	// Biography
	// in: string
	Biography string `json:"biography" example:"Опытный врач с 10-летним стажем"`

	// Avatar URL
	// in: string
	AvatarURL string `json:"avatar_url" example:"https://example.com/doctor.jpg"`

	// Consultation price
	// in: number
	ConsultationPrice float64 `json:"consultation_price" example:"15000"`

	// Years of experience
	// in: integer
	YearsOfExperience int `json:"years_of_experience" example:"10"`

	// Education information
	// in: string
	Education string `json:"education" example:"КазНМУ"`

	// Certificates list
	// in: array
	Certificates []string `json:"certificates"`

	// Working hours
	// in: string
	WorkingHours string `json:"working_hours" example:"9:00-18:00"`

	// Creation timestamp
	// in: string
	CreatedAt string `json:"created_at" example:"2024-01-01T12:00:00Z"`

	// Update timestamp
	// in: string
	UpdatedAt string `json:"updated_at" example:"2024-01-01T12:00:00Z"`
}

// ===============================
// ЗАПИСИ (APPOINTMENTS)
// ===============================

// AppointmentResponse represents appointment data
// swagger:model AppointmentResponse
type AppointmentResponse struct {
	// Appointment ID
	// in: string
	ID string `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`

	// Doctor ID
	// in: string
	DoctorID string `json:"doctor_id" example:"123e4567-e89b-12d3-a456-426614174001"`

	// Patient ID (null if slot not booked)
	// in: string
	PatientID *string `json:"patient_id" example:"123e4567-e89b-12d3-a456-426614174002"`

	// Start time
	// in: string
	StartTime string `json:"start_time" example:"2024-01-15T10:00:00Z"`

	// End time
	// in: string
	EndTime string `json:"end_time" example:"2024-01-15T11:00:00Z"`

	// Status (available, booked, completed, cancelled)
	// in: string
	Status string `json:"status" example:"booked"`

	// Appointment type
	// in: string
	AppointmentType string `json:"appointment_type" example:"consultation"`

	// Patient notes
	// in: string
	PatientNotes string `json:"patient_notes" example:"Головная боль"`

	// Doctor notes
	// in: string
	DoctorNotes string `json:"doctor_notes" example:"Рекомендую отдых"`
}

// BookAppointmentRequest represents booking request
// swagger:model BookAppointmentRequest
type BookAppointmentRequest struct {
	// Appointment type
	// in: string
	AppointmentType string `json:"appointment_type" example:"consultation"`

	// Patient notes
	// in: string
	PatientNotes string `json:"patient_notes" example:"Головная боль"`
}

// ===============================
// РАСПИСАНИЕ (SCHEDULES)
// ===============================

// ScheduleCreateRequest represents schedule creation request
// swagger:model ScheduleCreateRequest
type ScheduleCreateRequest struct {
	// Day of week (0 = Sunday, 1 = Monday, etc.)
	// in: integer
	DayOfWeek int `json:"day_of_week" example:"1"`

	// Start time
	// in: string
	StartTime string `json:"start_time" example:"09:00"`

	// End time
	// in: string
	EndTime string `json:"end_time" example:"18:00"`

	// Slot duration in minutes
	// in: integer
	SlotDuration int `json:"slot_duration" example:"30"`

	// Break duration in minutes
	// in: integer
	BreakDuration int `json:"break_duration" example:"10"`

	// Is active
	// in: boolean
	IsActive bool `json:"is_active" example:"true"`
}

// ScheduleResponse represents schedule data
// swagger:model ScheduleResponse
type ScheduleResponse struct {
	// Schedule ID
	// in: string
	ID string `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`

	// Doctor ID
	// in: string
	DoctorID string `json:"doctor_id" example:"123e4567-e89b-12d3-a456-426614174001"`

	// Day of week (0 = Sunday, 1 = Monday, etc.)
	// in: integer
	DayOfWeek int `json:"day_of_week" example:"1"`

	// Start time
	// in: string
	StartTime string `json:"start_time" example:"09:00:00"`

	// End time
	// in: string
	EndTime string `json:"end_time" example:"18:00:00"`

	// Slot duration in minutes
	// in: integer
	SlotDuration int `json:"slot_duration" example:"30"`

	// Break duration in minutes
	// in: integer
	BreakDuration int `json:"break_duration" example:"10"`

	// Is active
	// in: boolean
	IsActive bool `json:"is_active" example:"true"`

	// Creation timestamp
	// in: string
	CreatedAt string `json:"created_at" example:"2024-01-01T12:00:00Z"`

	// Update timestamp
	// in: string
	UpdatedAt string `json:"updated_at" example:"2024-01-01T12:00:00Z"`
}

// ===============================
// ОБЩИЕ МОДЕЛИ
// ===============================

// ErrorResponse represents error response
// swagger:model ErrorResponse
type ErrorResponse struct {
	// Error message
	// in: string
	Error string `json:"error" example:"Неверные данные"`
}

// SuccessResponse represents success response
// swagger:model SuccessResponse
type SuccessResponse struct {
	// Success message
	// in: string
	Message string `json:"message" example:"Операция выполнена успешно"`
}

// ListResponse represents paginated list response
// swagger:model ListResponse
type ListResponse struct {
	// Total count
	// in: integer
	Total int `json:"total" example:"100"`

	// Items count in current page
	// in: integer
	Count int `json:"count" example:"10"`

	// Current page
	// in: integer
	Page int `json:"page" example:"1"`

	// Data array
	// in: array
	Data interface{} `json:"data"`
}

// ===============================
// ФАЙЛЫ (FILES)
// ===============================

// FileResponse represents file metadata response
// swagger:model FileResponse
type FileResponse struct {
	// File ID
	// in: string
	ID string `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`

	// Original filename
	// in: string
	Filename string `json:"filename" example:"document.pdf"`

	// File size in bytes
	// in: integer
	Size int64 `json:"size" example:"1048576"`

	// MIME content type
	// in: string
	ContentType string `json:"content_type" example:"application/pdf"`

	// File description
	// in: string
	Description string `json:"description" example:"Medical document"`

	// Is file public
	// in: boolean
	IsPublic bool `json:"is_public" example:"false"`

	// Upload timestamp
	// in: string
	UploadedAt string `json:"uploaded_at" example:"2024-01-01T12:00:00Z"`

	// User ID who uploaded the file
	// in: string
	UserID string `json:"user_id" example:"123e4567-e89b-12d3-a456-426614174001"`

	// Public download URL (if public)
	// in: string
	PublicURL *string `json:"public_url,omitempty" example:"https://api.example.com/public/123e4567"`
}

// FileUploadResponse represents file upload response
// swagger:model FileUploadResponse
type FileUploadResponse struct {
	// File ID
	// in: string
	ID string `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`

	// Original filename
	// in: string
	Filename string `json:"filename" example:"document.pdf"`

	// Success message
	// in: string
	Message string `json:"message" example:"Файл успешно загружен"`

	// Public download URL (if made public)
	// in: string
	PublicURL *string `json:"public_url,omitempty" example:"https://api.example.com/public/123e4567"`
}

// FileVisibilityRequest represents file visibility toggle request
// swagger:model FileVisibilityRequest
type FileVisibilityRequest struct {
	// Make file public (true) or private (false)
	// in: boolean
	IsPublic bool `json:"is_public" example:"true"`
}

// FileVisibilityResponse represents file visibility response
// swagger:model FileVisibilityResponse
type FileVisibilityResponse struct {
	// Success message
	// in: string
	Message string `json:"message" example:"Видимость файла изменена"`

	// Current visibility status
	// in: boolean
	IsPublic bool `json:"is_public" example:"true"`

	// Public URL if made public
	// in: string
	PublicURL *string `json:"public_url,omitempty" example:"https://api.example.com/public/123e4567"`
}

// ===============================
// УВЕДОМЛЕНИЯ (NOTIFICATIONS)
// ===============================

// NotificationCreateRequest represents notification creation request
// swagger:model NotificationCreateRequest
type NotificationCreateRequest struct {
	// Recipient ID
	// in: string
	RecipientID string `json:"recipient_id" example:"123e4567-e89b-12d3-a456-426614174000"`

	// Notification type (appointment, reminder, system, etc.)
	// in: string
	Type string `json:"type" example:"appointment"`

	// Notification title
	// in: string
	Title string `json:"title" example:"Напоминание о записи"`

	// Notification message
	// in: string
	Message string `json:"message" example:"У вас запись к врачу завтра в 10:00"`

	// Delivery channel (email, sms, telegram, push)
	// in: string
	Channel string `json:"channel" example:"email"`

	// Metadata for additional information
	// in: object
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// NotificationResponse represents notification data
// swagger:model NotificationResponse
type NotificationResponse struct {
	// Notification ID
	// in: integer
	ID int `json:"id" example:"123"`

	// Recipient ID
	// in: string
	RecipientID string `json:"recipient_id" example:"123e4567-e89b-12d3-a456-426614174000"`

	// Notification type
	// in: string
	Type string `json:"type" example:"appointment"`

	// Notification title
	// in: string
	Title string `json:"title" example:"Напоминание о записи"`

	// Notification message
	// in: string
	Message string `json:"message" example:"У вас запись к врачу завтра в 10:00"`

	// Delivery channel
	// in: string
	Channel string `json:"channel" example:"email"`

	// Notification status (pending, sent, failed, delivered)
	// in: string
	Status string `json:"status" example:"sent"`

	// Metadata
	// in: object
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// Creation timestamp
	// in: string
	CreatedAt string `json:"created_at" example:"2024-01-01T12:00:00Z"`

	// Sent timestamp
	// in: string
	SentAt *string `json:"sent_at,omitempty" example:"2024-01-01T12:05:00Z"`
}

// ===============================
// АВТОРИЗАЦИЯ И ПОЛЬЗОВАТЕЛИ
// ===============================

// LoginRequest represents login request
// swagger:model LoginRequest
type LoginRequest struct {
	// Email address
	// in: string
	Email string `json:"email" example:"user@example.com"`

	// Password
	// in: string
	Password string `json:"password" example:"password123"`
}

// LoginResponse represents login response
// swagger:model LoginResponse
type LoginResponse struct {
	// JWT access token
	// in: string
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIs..."`

	// User information
	// in: object
	User UserInfo `json:"user"`

	// Token expiration time in hours
	// in: integer
	ExpiresIn int `json:"expires_in" example:"168"`
}

// RegisterRequest represents registration request
// swagger:model RegisterRequest
type RegisterRequest struct {
	// Email address
	// in: string
	Email string `json:"email" example:"user@example.com"`

	// Password
	// in: string
	Password string `json:"password" example:"password123"`

	// User role (patient, doctor)
	// in: string
	Role string `json:"role" example:"patient"`
}

// RegisterResponse represents registration response
// swagger:model RegisterResponse
type RegisterResponse struct {
	// Created user ID
	// in: string
	UserID string `json:"user_id" example:"123e4567-e89b-12d3-a456-426614174000"`

	// Success message
	// in: string
	Message string `json:"message" example:"Пользователь зарегистрирован. Завершите настройку профиля."`

	// Next step URL for profile completion
	// in: string
	NextStep string `json:"next_step" example:"/users/{userID}/patient"`
}

// ValidateTokenResponse represents token validation response
// swagger:model ValidateTokenResponse
type ValidateTokenResponse struct {
	// Is token valid
	// in: boolean
	Valid bool `json:"valid" example:"true"`

	// User ID from token
	// in: string
	UserID string `json:"user_id" example:"123e4567-e89b-12d3-a456-426614174000"`

	// User role from token
	// in: string
	Role string `json:"role" example:"patient"`

	// Token expiration timestamp
	// in: string
	ExpiresAt string `json:"expires_at" example:"2024-01-08T12:00:00Z"`
}

// ===============================
// РАСПИСАНИЕ ИСКЛЮЧЕНИЯ
// ===============================

// ExceptionCreateRequest represents schedule exception creation request
// swagger:model ExceptionCreateRequest
type ExceptionCreateRequest struct {
	// Exception date
	// in: string
	Date string `json:"date" example:"2024-01-15"`

	// Exception type (day_off, custom_hours, special_event)
	// in: string
	Type string `json:"type" example:"day_off"`

	// Reason for exception
	// in: string
	Reason string `json:"reason" example:"Отпуск"`

	// Custom start time (for custom_hours type)
	// in: string
	CustomStartTime *string `json:"custom_start_time,omitempty" example:"11:00"`

	// Custom end time (for custom_hours type)
	// in: string
	CustomEndTime *string `json:"custom_end_time,omitempty" example:"15:00"`
}

// ExceptionResponse represents schedule exception
// swagger:model ExceptionResponse
type ExceptionResponse struct {
	// Exception ID
	// in: string
	ID string `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`

	// Doctor ID
	// in: string
	DoctorID string `json:"doctor_id" example:"123e4567-e89b-12d3-a456-426614174001"`

	// Exception date
	// in: string
	Date string `json:"date" example:"2024-01-15"`

	// Exception type
	// in: string
	Type string `json:"type" example:"day_off"`

	// Reason for exception
	// in: string
	Reason string `json:"reason" example:"Отпуск"`

	// Custom start time
	// in: string
	CustomStartTime *string `json:"custom_start_time,omitempty" example:"11:00:00"`

	// Custom end time
	// in: string
	CustomEndTime *string `json:"custom_end_time,omitempty" example:"15:00:00"`

	// Creation timestamp
	// in: string
	CreatedAt string `json:"created_at" example:"2024-01-01T12:00:00Z"`
}
