package models

import (
	"time"

	"github.com/google/uuid"
)

// NotificationChannel — канал доставки уведомления
type NotificationChannel string

const (
	ChannelEmail    NotificationChannel = "email"
	ChannelSMS      NotificationChannel = "sms"
	ChannelTelegram NotificationChannel = "telegram"
	ChannelPush     NotificationChannel = "push"
)

// NotificationType — тип события, которое вызвало уведомление
type NotificationType string

const (
	// 🏥 Medical/Appointment
	AppointmentBooked      NotificationType = "appointment.booked"
	AppointmentCanceled    NotificationType = "appointment.canceled"
	AppointmentReminder    NotificationType = "appointment.reminder"
	AppointmentConfirmed   NotificationType = "appointment.confirmed"
	AppointmentNew         NotificationType = "appointment.new"
	AppointmentRescheduled NotificationType = "appointment.rescheduled"

	// 👤 User/Profile
	UserRegistered      NotificationType = "user.registered"
	UserProfileUpdated  NotificationType = "user.profile.updated"
	UserPasswordChanged NotificationType = "user.password.changed"

	// 💊 Medical/Treatment
	PrescriptionIssued   NotificationType = "prescription.issued"
	PrescriptionExpiring NotificationType = "prescription.expiring"
	TreatmentStarted     NotificationType = "treatment.started"
	TreatmentCompleted   NotificationType = "treatment.completed"

	// 📋 Medical/Results
	TestResultsReady    NotificationType = "test.results.ready"
	LabResultsAvailable NotificationType = "lab.results.available"

	// 🔔 General/System
	SystemMaintenance NotificationType = "system.maintenance"
	SecurityAlert     NotificationType = "security.alert"
	PaymentProcessed  NotificationType = "payment.processed"
	PaymentFailed     NotificationType = "payment.failed"

	// 🧩 System/Logger
	SystemErrorOccurred   NotificationType = "system.error.occurred"
	GatewayServiceFailure NotificationType = "gateway.service.failure"
)

// DeliveryStatus — состояние доставки уведомления
type DeliveryStatus string

const (
	StatusPending DeliveryStatus = "pending"
	StatusSent    DeliveryStatus = "sent"
	StatusFailed  DeliveryStatus = "failed"
)

// Notification — основная сущность уведомления
type Notification struct {
	ID          uint                `gorm:"primaryKey" json:"id"`
	Type        NotificationType    `gorm:"type:varchar(100);not null" json:"type"`
	Channel     NotificationChannel `gorm:"type:varchar(50);not null" json:"channel"`
	RecipientID uuid.UUID           `gorm:"type:uuid;not null;index" json:"recipientId"`
	Recipient   string              `gorm:"type:varchar(255);not null" json:"recipient"`
	Message     string              `gorm:"type:text;not null" json:"message"`
	Status      DeliveryStatus      `gorm:"type:varchar(20);not null;default:'pending'" json:"status"`
	Attempts    int                 `gorm:"not null;default:0" json:"attempts"`
	LastError   *string             `gorm:"type:text" json:"lastError,omitempty"`
	CreatedAt   time.Time           `json:"createdAt"`
	SentAt      *time.Time          `json:"sentAt,omitempty"`
}

// TableName указывает GORM использовать имя таблицы "notifications"
func (Notification) TableName() string {
	return "notifications"
}
