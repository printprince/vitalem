package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/printprince/vitalem/appointment_service/internal/models"
	"gorm.io/gorm"
)

// AppointmentRepository - интерфейс репозитория
type AppointmentRepository interface {
	// Schedules
	CreateSchedule(schedule *models.DoctorSchedule) error
	GetDoctorSchedules(doctorID uuid.UUID) ([]*models.DoctorSchedule, error)
	GetScheduleByID(id uuid.UUID) (*models.DoctorSchedule, error)
	UpdateSchedule(schedule *models.DoctorSchedule) error
	DeleteSchedule(id uuid.UUID) error
	// НОВЫЙ метод для удаления слотов расписания
	DeleteScheduleSlots(scheduleID uuid.UUID) error
	// НОВЫЙ метод для получения сгенерированных слотов
	GetScheduleSlots(scheduleID uuid.UUID, startDate, endDate time.Time) ([]*models.Appointment, error)
	// НОВЫЙ метод для обновления типа записи в слотах расписания
	UpdateScheduleAppointmentType(scheduleID uuid.UUID, appointmentType string) error

	// Appointments
	CreateAppointment(appointment *models.Appointment) error
	GetAppointmentByID(id uuid.UUID) (*models.Appointment, error)
	GetAvailableSlots(doctorID uuid.UUID, startDate, endDate time.Time) ([]*models.Appointment, error)
	GetDoctorAppointments(doctorID uuid.UUID, startDate, endDate time.Time) ([]*models.Appointment, error)
	GetPatientAppointments(patientID uuid.UUID, startDate, endDate time.Time) ([]*models.Appointment, error)
	UpdateAppointment(appointment *models.Appointment) error
	CheckSlotExists(doctorID uuid.UUID, startTime, endTime time.Time) (bool, error)

	// Exceptions
	CreateException(exception *models.ScheduleException) error
	GetDoctorExceptions(doctorID uuid.UUID, startDate, endDate time.Time) ([]*models.ScheduleException, error)
	DeleteException(id uuid.UUID) error
}

// appointmentRepository - реализация репозитория
type appointmentRepository struct {
	db *gorm.DB
}

// NewAppointmentRepository - создание нового репозитория
func NewAppointmentRepository(db *gorm.DB) AppointmentRepository {
	return &appointmentRepository{db: db}
}

// === SCHEDULES ===

func (r *appointmentRepository) CreateSchedule(schedule *models.DoctorSchedule) error {
	return r.db.Create(schedule).Error
}

func (r *appointmentRepository) GetDoctorSchedules(doctorID uuid.UUID) ([]*models.DoctorSchedule, error) {
	var schedules []*models.DoctorSchedule
	err := r.db.Where("doctor_id = ?", doctorID).
		Order("is_active DESC, created_at DESC").
		Find(&schedules).Error
	return schedules, err
}

func (r *appointmentRepository) GetScheduleByID(id uuid.UUID) (*models.DoctorSchedule, error) {
	var schedule models.DoctorSchedule
	err := r.db.Where("id = ?", id).First(&schedule).Error
	if err != nil {
		return nil, err
	}
	return &schedule, nil
}

func (r *appointmentRepository) UpdateSchedule(schedule *models.DoctorSchedule) error {
	return r.db.Save(schedule).Error
}

func (r *appointmentRepository) DeleteSchedule(id uuid.UUID) error {
	// Сначала удаляем все слоты этого расписания (только доступные)
	if err := r.DeleteScheduleSlots(id); err != nil {
		return err
	}

	// Затем физически удаляем расписание из базы данных
	return r.db.Delete(&models.DoctorSchedule{}, "id = ?", id).Error
}

// === APPOINTMENTS ===

func (r *appointmentRepository) CreateAppointment(appointment *models.Appointment) error {
	return r.db.Create(appointment).Error
}

func (r *appointmentRepository) GetAppointmentByID(id uuid.UUID) (*models.Appointment, error) {
	var appointment models.Appointment
	err := r.db.Where("id = ?", id).First(&appointment).Error
	if err != nil {
		return nil, err
	}
	return &appointment, nil
}

func (r *appointmentRepository) GetAvailableSlots(doctorID uuid.UUID, startDate, endDate time.Time) ([]*models.Appointment, error) {
	var appointments []*models.Appointment
	err := r.db.Where("doctor_id = ? AND status = ? AND start_time >= ? AND end_time <= ?",
		doctorID, "available", startDate, endDate).
		Order("start_time ASC").
		Find(&appointments).Error
	return appointments, err
}

func (r *appointmentRepository) GetDoctorAppointments(doctorID uuid.UUID, startDate, endDate time.Time) ([]*models.Appointment, error) {
	var appointments []*models.Appointment
	err := r.db.Where("doctor_id = ? AND start_time >= ? AND end_time <= ?",
		doctorID, startDate, endDate).
		Order("start_time ASC").
		Find(&appointments).Error
	return appointments, err
}

func (r *appointmentRepository) GetPatientAppointments(patientID uuid.UUID, startDate, endDate time.Time) ([]*models.Appointment, error) {
	var appointments []*models.Appointment
	err := r.db.Where("patient_id = ? AND start_time >= ? AND end_time <= ?",
		patientID, startDate, endDate).
		Order("start_time ASC").
		Find(&appointments).Error
	return appointments, err
}

func (r *appointmentRepository) UpdateAppointment(appointment *models.Appointment) error {
	return r.db.Save(appointment).Error
}

// === EXCEPTIONS ===

func (r *appointmentRepository) CreateException(exception *models.ScheduleException) error {
	return r.db.Create(exception).Error
}

func (r *appointmentRepository) GetDoctorExceptions(doctorID uuid.UUID, startDate, endDate time.Time) ([]*models.ScheduleException, error) {
	var exceptions []*models.ScheduleException
	err := r.db.Where("doctor_id = ? AND date >= ? AND date <= ?",
		doctorID, startDate, endDate).
		Order("date ASC").
		Find(&exceptions).Error
	return exceptions, err
}

func (r *appointmentRepository) DeleteException(id uuid.UUID) error {
	return r.db.Delete(&models.ScheduleException{}, "id = ?", id).Error
}

func (r *appointmentRepository) CheckSlotExists(doctorID uuid.UUID, startTime, endTime time.Time) (bool, error) {
	var count int64
	err := r.db.Model(&models.Appointment{}).
		Where("doctor_id = ? AND ((start_time <= ? AND end_time > ?) OR (start_time < ? AND end_time >= ?) OR (start_time >= ? AND end_time <= ?))",
			doctorID, startTime, startTime, endTime, endTime, startTime, endTime).
		Count(&count).Error
	return count > 0, err
}

// === NEW METHODS ===

func (r *appointmentRepository) DeleteScheduleSlots(scheduleID uuid.UUID) error {
	// Удаляем только доступные слоты (не забронированные)
	// Забронированные слоты сохраняем для истории
	return r.db.Where("schedule_id = ? AND status = ?", scheduleID, "available").
		Delete(&models.Appointment{}).Error
}

func (r *appointmentRepository) GetScheduleSlots(scheduleID uuid.UUID, startDate, endDate time.Time) ([]*models.Appointment, error) {
	var slots []*models.Appointment
	err := r.db.Where("schedule_id = ? AND start_time >= ? AND end_time <= ?",
		scheduleID, startDate, endDate).
		Order("start_time ASC").
		Find(&slots).Error
	return slots, err
}

func (r *appointmentRepository) UpdateScheduleAppointmentType(scheduleID uuid.UUID, appointmentType string) error {
	return r.db.Model(&models.Appointment{}).
		Where("schedule_id = ? AND status = ?", scheduleID, "available").
		Update("appointment_type", appointmentType).Error
}
