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
		log.Println("🔄 Recreating doctor_schedules table with correct structure...")

		// Сохраняем данные во временную таблицу
		err = db.Exec(`
			CREATE TABLE doctor_schedules_backup AS 
			SELECT 
				id, 
				doctor_id, 
				name, 
				(SELECT json_agg(unnest)::text FROM unnest(work_days)) as work_days_json,
				start_time, 
				end_time, 
				break_start, 
				break_end, 
				slot_duration, 
				slot_title, 
				is_active, 
				is_default, 
				created_at, 
				updated_at 
			FROM doctor_schedules
		`).Error
		if err != nil {
			return fmt.Errorf("failed to backup data: %w", err)
		}

		// Удаляем старую таблицу
		err = db.Exec("DROP TABLE doctor_schedules").Error
		if err != nil {
			return fmt.Errorf("failed to drop old table: %w", err)
		}

		// Создаем новую таблицу с правильной структурой
		err = db.Exec(`
			CREATE TABLE doctor_schedules (
				id UUID PRIMARY KEY,
				doctor_id UUID NOT NULL,
				name VARCHAR(255) NOT NULL,
				work_days_json TEXT NOT NULL,
				start_time VARCHAR(5) NOT NULL,
				end_time VARCHAR(5) NOT NULL,
				break_start VARCHAR(5),
				break_end VARCHAR(5),
				slot_duration INTEGER NOT NULL DEFAULT 30,
				slot_title VARCHAR(255),
				is_active BOOLEAN DEFAULT true,
				is_default BOOLEAN DEFAULT false,
				created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
				updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
			)
		`).Error
		if err != nil {
			return fmt.Errorf("failed to create new table: %w", err)
		}

		// Восстанавливаем данные
		err = db.Exec(`
			INSERT INTO doctor_schedules 
			SELECT * FROM doctor_schedules_backup
		`).Error
		if err != nil {
			return fmt.Errorf("failed to restore data: %w", err)
		}

		// Удаляем временную таблицу
		err = db.Exec("DROP TABLE doctor_schedules_backup").Error
		if err != nil {
			log.Printf("⚠️ Failed to drop backup table: %v", err)
		}

		// Создаем индекс
		err = db.Exec("CREATE INDEX IF NOT EXISTS idx_schedules_doctor_active ON doctor_schedules(doctor_id, is_active)").Error
		if err != nil {
			log.Printf("⚠️ Failed to create index: %v", err)
		}

		log.Println("✅ Successfully recreated doctor_schedules table with work_days_json")
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
