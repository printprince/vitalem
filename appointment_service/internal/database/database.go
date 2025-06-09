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

	// Проверяем и исправляем структуру doctor_schedules если нужно
	if err := checkAndFixScheduleTable(db); err != nil {
		return fmt.Errorf("failed to check/fix schedule table: %w", err)
	}

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

// checkAndFixScheduleTable - проверяет и исправляет структуру таблицы doctor_schedules
func checkAndFixScheduleTable(db *gorm.DB) error {
	log.Println("🔍 Checking doctor_schedules table structure...")

	// Проверяем существует ли таблица
	var tableExists bool
	err := db.Raw("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = CURRENT_SCHEMA() AND table_name = 'doctor_schedules')").Scan(&tableExists).Error
	if err != nil {
		return fmt.Errorf("failed to check table existence: %w", err)
	}

	if !tableExists {
		log.Println("📝 Table doctor_schedules doesn't exist yet, will be created by AutoMigrate")
		return nil
	}

	// Проверяем есть ли проблемный столбец work_days (integer[])
	var workDaysColumnExists bool
	err = db.Raw("SELECT EXISTS (SELECT FROM information_schema.columns WHERE table_schema = CURRENT_SCHEMA() AND table_name = 'doctor_schedules' AND column_name = 'work_days')").Scan(&workDaysColumnExists).Error
	if err != nil {
		return fmt.Errorf("failed to check work_days column: %w", err)
	}

	// Проверяем есть ли правильный столбец work_days_json
	var workDaysJsonColumnExists bool
	err = db.Raw("SELECT EXISTS (SELECT FROM information_schema.columns WHERE table_schema = CURRENT_SCHEMA() AND table_name = 'doctor_schedules' AND column_name = 'work_days_json')").Scan(&workDaysJsonColumnExists).Error
	if err != nil {
		return fmt.Errorf("failed to check work_days_json column: %w", err)
	}

	if workDaysColumnExists {
		log.Println("⚠️  CRITICAL: Found old work_days column (integer[]) in doctor_schedules table")
		log.Println("⚠️  This column conflicts with the new structure and must be removed")
		log.Println("⚠️  Please run the following SQL commands manually to fix this:")
		log.Println("⚠️  ")
		log.Println("⚠️  docker exec -it vitalem_postgres psql -U vitalem_user -d vitalem_db")
		log.Println("⚠️  ")
		log.Println("⚠️  -- If you want to preserve data:")
		log.Println("⚠️  ALTER TABLE doctor_schedules ADD COLUMN work_days_json TEXT;")
		log.Println("⚠️  UPDATE doctor_schedules SET work_days_json = (SELECT json_agg(unnest)::text FROM unnest(work_days)) WHERE work_days IS NOT NULL;")
		log.Println("⚠️  UPDATE doctor_schedules SET work_days_json = '[]' WHERE work_days_json IS NULL;")
		log.Println("⚠️  ALTER TABLE doctor_schedules ALTER COLUMN work_days_json SET NOT NULL;")
		log.Println("⚠️  ALTER TABLE doctor_schedules DROP COLUMN work_days;")
		log.Println("⚠️  ")
		log.Println("⚠️  -- Or if you want to recreate the table from scratch:")
		log.Println("⚠️  DROP TABLE doctor_schedules CASCADE;")
		log.Println("⚠️  ")

		return fmt.Errorf("table doctor_schedules contains incompatible work_days column - manual intervention required")
	}

	if !workDaysJsonColumnExists {
		log.Println("📝 Table structure looks compatible, work_days_json will be created by AutoMigrate")
	} else {
		log.Println("✅ Table structure is correct, work_days_json column exists")
	}

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

		// Уникальный индекс для предотвращения дублирования слотов
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_appointments_doctor_time_unique ON appointments(doctor_id, start_time, end_time);",

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
