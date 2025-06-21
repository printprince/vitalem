package service

import (
	"fmt"
	"strings"

	"NotificationService/internal/domain/models"
)

// MessageFormatter форматирует сообщения для разных каналов
type MessageFormatter struct{}

func NewMessageFormatter() *MessageFormatter {
	return &MessageFormatter{}
}

// FormatForTelegram форматирует сообщение для Telegram с Markdown
func (f *MessageFormatter) FormatForTelegram(notification *models.Notification) (string, bool) {
	switch notification.Type {
	case models.UserRegistered:
		return f.formatUserRegistered(notification), true
	case models.UserProfileUpdated:
		return f.formatUserProfileUpdated(notification), true
	case models.AppointmentBooked:
		return f.formatAppointmentBooked(notification), true
	case models.AppointmentCanceled:
		return f.formatAppointmentCanceled(notification), true
	default:
		// Для неизвестных типов возвращаем обычное сообщение
		return notification.Message, false
	}
}

func (f *MessageFormatter) formatUserRegistered(notification *models.Notification) string {
	userMeta, err := notification.GetUserMetadata()
	if err != nil || userMeta == nil {
		return f.escapeMarkdown("🎉 Новый пользователь зарегистрировался в системе!")
	}

	var sb strings.Builder
	sb.WriteString("🎉 *Новый пользователь зарегистрировался\\!*\n\n")

	if userMeta.FullName != "" {
		sb.WriteString(fmt.Sprintf("👤 *Имя:* %s\n", f.escapeMarkdown(userMeta.FullName)))
	}

	if userMeta.Email != "" {
		sb.WriteString(fmt.Sprintf("📧 *Email:* `%s`\n", f.escapeMarkdown(userMeta.Email)))
	}

	if userMeta.Username != "" {
		sb.WriteString(fmt.Sprintf("🏷️ *Логин:* %s\n", f.escapeMarkdown(userMeta.Username)))
	}

	if userMeta.Role != "" {
		roleEmoji := f.getRoleEmoji(userMeta.Role)
		sb.WriteString(fmt.Sprintf("%s *Роль:* %s\n", roleEmoji, f.escapeMarkdown(userMeta.Role)))
	}

	sb.WriteString(fmt.Sprintf("\n🕐 *Время регистрации:* %s", f.escapeMarkdown(notification.CreatedAt.Format("02.01.2006 15:04"))))

	return sb.String()
}

func (f *MessageFormatter) formatUserProfileUpdated(notification *models.Notification) string {
	userMeta, err := notification.GetUserMetadata()
	if err != nil || userMeta == nil {
		return f.escapeMarkdown("✏️ Пользователь обновил свой профиль")
	}

	var sb strings.Builder
	sb.WriteString("✏️ *Профиль пользователя обновлен*\n\n")

	if userMeta.FullName != "" {
		sb.WriteString(fmt.Sprintf("👤 *Пользователь:* %s\n", f.escapeMarkdown(userMeta.FullName)))
	}

	if userMeta.Email != "" {
		sb.WriteString(fmt.Sprintf("📧 *Email:* `%s`\n", f.escapeMarkdown(userMeta.Email)))
	}

	sb.WriteString(fmt.Sprintf("\n🕐 *Время обновления:* %s", f.escapeMarkdown(notification.CreatedAt.Format("02.01.2006 15:04"))))

	return sb.String()
}

func (f *MessageFormatter) formatAppointmentBooked(notification *models.Notification) string {
	appointmentMeta, err := notification.GetAppointmentMetadata()
	if err != nil || appointmentMeta == nil {
		return f.escapeMarkdown("📅 Новая запись к врачу")
	}

	var sb strings.Builder
	sb.WriteString("📅 *Новая запись к врачу\\!*\n\n")

	if appointmentMeta.PatientName != "" {
		sb.WriteString(fmt.Sprintf("👤 *Пациент:* %s\n", f.escapeMarkdown(appointmentMeta.PatientName)))
	}

	if appointmentMeta.DoctorName != "" {
		sb.WriteString(fmt.Sprintf("👨‍⚕️ *Врач:* %s\n", f.escapeMarkdown(appointmentMeta.DoctorName)))
	}

	if appointmentMeta.Specialty != "" {
		sb.WriteString(fmt.Sprintf("🏥 *Специальность:* %s\n", f.escapeMarkdown(appointmentMeta.Specialty)))
	}

	sb.WriteString(fmt.Sprintf("🕐 *Дата и время:* %s\n", f.escapeMarkdown(appointmentMeta.DateTime.Format("02.01.2006 15:04"))))

	if appointmentMeta.Duration > 0 {
		sb.WriteString(fmt.Sprintf("⏱️ *Длительность:* %d мин\n", appointmentMeta.Duration))
	}

	return sb.String()
}

func (f *MessageFormatter) formatAppointmentCanceled(notification *models.Notification) string {
	appointmentMeta, err := notification.GetAppointmentMetadata()
	if err != nil || appointmentMeta == nil {
		return f.escapeMarkdown("❌ Запись к врачу отменена")
	}

	var sb strings.Builder
	sb.WriteString("❌ *Запись к врачу отменена*\n\n")

	if appointmentMeta.PatientName != "" {
		sb.WriteString(fmt.Sprintf("👤 *Пациент:* %s\n", f.escapeMarkdown(appointmentMeta.PatientName)))
	}

	if appointmentMeta.DoctorName != "" {
		sb.WriteString(fmt.Sprintf("👨‍⚕️ *Врач:* %s\n", f.escapeMarkdown(appointmentMeta.DoctorName)))
	}

	sb.WriteString(fmt.Sprintf("🕐 *Дата и время:* %s\n", f.escapeMarkdown(appointmentMeta.DateTime.Format("02.01.2006 15:04"))))

	return sb.String()
}

// getRoleEmoji возвращает эмодзи для роли
func (f *MessageFormatter) getRoleEmoji(role string) string {
	switch strings.ToLower(role) {
	case "doctor", "врач":
		return "👨‍⚕️"
	case "patient", "пациент":
		return "🤒"
	case "admin", "администратор":
		return "👑"
	case "nurse", "медсестра":
		return "👩‍⚕️"
	default:
		return "👤"
	}
}

// escapeMarkdown экранирует специальные символы для MarkdownV2
func (f *MessageFormatter) escapeMarkdown(text string) string {
	// Символы, которые нужно экранировать в MarkdownV2
	specialChars := []string{"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"}

	result := text
	for _, char := range specialChars {
		result = strings.ReplaceAll(result, char, "\\"+char)
	}

	return result
}
