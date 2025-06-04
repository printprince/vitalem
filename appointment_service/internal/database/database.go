package database

import (
	"fmt"
	"log"

	"github.com/printprince/vitalem/appointment_service/internal/config"
	"github.com/printprince/vitalem/appointment_service/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ConnectDB - подключение к базе данных
func ConnectDB(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("✅ Connected to database successfully")
	return db, nil
}

// RunMigrations - выполнение миграций
func RunMigrations(db *gorm.DB) error {
	log.Println("🔄 Running database migrations...")

	// Автоматическая миграция моделей
	err := db.AutoMigrate(
		&models.DoctorSchedule{},
		&models.ScheduleException{},
		&models.Appointment{},
	)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("✅ Database migrations completed successfully")
	return nil
}

// CreateIndexes - создание индексов для оптимизации
func CreateIndexes(db *gorm.DB) error {
	log.Println("🔄 Creating database indexes...")

	// Создание индексов для appointments таблицы
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_appointments_doctor_start_time ON appointments(doctor_id, start_time);",
		"CREATE INDEX IF NOT EXISTS idx_appointments_status_start_time ON appointments(status, start_time);",
		"CREATE INDEX IF NOT EXISTS idx_appointments_patient_start_time ON appointments(patient_id, start_time) WHERE patient_id IS NOT NULL;",

		// Индексы для schedule_exceptions
		"CREATE INDEX IF NOT EXISTS idx_exceptions_doctor_date ON schedule_exceptions(doctor_id, date);",

		// Индексы для doctor_schedules
		"CREATE INDEX IF NOT EXISTS idx_schedules_doctor_active ON doctor_schedules(doctor_id, is_active);",
	}

	for _, indexSQL := range indexes {
		if err := db.Exec(indexSQL).Error; err != nil {
			log.Printf("⚠️ Failed to create index: %s - %v", indexSQL, err)
			// Продолжаем выполнение, индексы не критичны
		}
	}

	log.Println("✅ Database indexes created successfully")
	return nil
}
