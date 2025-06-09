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

	// Ручная миграция для изменения структуры doctor_schedules
	if err := migrateScheduleWorkDays(db); err != nil {
		return fmt.Errorf("failed to migrate schedule work days: %w", err)
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

// migrateScheduleWorkDays - миграция для изменения work_days на work_days_json
func migrateScheduleWorkDays(db *gorm.DB) error {
	log.Println("🔄 Migrating doctor_schedules work_days to work_days_json...")

	// Проверяем существует ли таблица doctor_schedules
	var tableExists bool
	err := db.Raw("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = CURRENT_SCHEMA() AND table_name = 'doctor_schedules')").Scan(&tableExists).Error
	if err != nil {
		return fmt.Errorf("failed to check table existence: %w", err)
	}

	if !tableExists {
		log.Println("📝 Table doctor_schedules doesn't exist yet, skipping migration")
		return nil
	}

	// Проверяем существует ли старый столбец work_days
	var workDaysColumnExists bool
	err = db.Raw("SELECT EXISTS (SELECT FROM information_schema.columns WHERE table_schema = CURRENT_SCHEMA() AND table_name = 'doctor_schedules' AND column_name = 'work_days')").Scan(&workDaysColumnExists).Error
	if err != nil {
		return fmt.Errorf("failed to check work_days column: %w", err)
	}

	// Проверяем существует ли новый столбец work_days_json
	var workDaysJsonColumnExists bool
	err = db.Raw("SELECT EXISTS (SELECT FROM information_schema.columns WHERE table_schema = CURRENT_SCHEMA() AND table_name = 'doctor_schedules' AND column_name = 'work_days_json')").Scan(&workDaysJsonColumnExists).Error
	if err != nil {
		return fmt.Errorf("failed to check work_days_json column: %w", err)
	}

	if workDaysColumnExists && !workDaysJsonColumnExists {
		log.Println("🔄 Converting work_days integer[] to work_days_json text...")

		// 1. Добавляем новый столбец work_days_json
		err = db.Exec("ALTER TABLE doctor_schedules ADD COLUMN work_days_json TEXT").Error
		if err != nil {
			return fmt.Errorf("failed to add work_days_json column: %w", err)
		}

		// 2. Конвертируем данные из work_days в work_days_json
		err = db.Exec(`
			UPDATE doctor_schedules 
			SET work_days_json = (
				SELECT json_agg(unnest)::text 
				FROM unnest(work_days)
			) 
			WHERE work_days IS NOT NULL AND array_length(work_days, 1) > 0`).Error
		if err != nil {
			return fmt.Errorf("failed to convert work_days data: %w", err)
		}

		// 3. Устанавливаем значение по умолчанию для пустых массивов
		err = db.Exec("UPDATE doctor_schedules SET work_days_json = '[]' WHERE work_days_json IS NULL").Error
		if err != nil {
			return fmt.Errorf("failed to set default work_days_json values: %w", err)
		}

		// 4. Добавляем NOT NULL constraint
		err = db.Exec("ALTER TABLE doctor_schedules ALTER COLUMN work_days_json SET NOT NULL").Error
		if err != nil {
			return fmt.Errorf("failed to set NOT NULL constraint: %w", err)
		}

		// 5. Удаляем старый столбец work_days
		err = db.Exec("ALTER TABLE doctor_schedules DROP COLUMN work_days").Error
		if err != nil {
			return fmt.Errorf("failed to drop work_days column: %w", err)
		}

		log.Println("✅ Successfully migrated work_days to work_days_json")
	} else if workDaysJsonColumnExists && !workDaysColumnExists {
		log.Println("📝 Migration already completed - work_days_json column exists")
	} else if !workDaysColumnExists && !workDaysJsonColumnExists {
		log.Println("📝 Fresh installation - no migration needed")
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
