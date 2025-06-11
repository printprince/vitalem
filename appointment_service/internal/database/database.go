package database

import (
	"fmt"
	"log"

	"github.com/printprince/vitalem/appointment_service/internal/config"

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

	log.Println("Connected to database successfully")
	return db, nil
}

// RunMigrations - выполнение миграций
func RunMigrations(db *gorm.DB) error {
	log.Println("Running database migrations...")

	// Проверяем и исправляем структуру doctor_schedules если нужно
	if err := checkAndFixScheduleTable(db); err != nil {
		return fmt.Errorf("failed to check/fix schedule table: %w", err)
	}

	// Проверяем базовую связность с базой данных
	log.Println("Testing database connectivity...")
	var dbName string
	err := db.Raw("SELECT current_database()").Scan(&dbName).Error
	if err != nil {
		return fmt.Errorf("failed to test database connectivity: %w", err)
	}
	log.Printf("Database connectivity test successful: %s", dbName)

	// Создаем таблицы вручную
	log.Println("Creating tables...")

	// Создаем таблицу doctor_schedules
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS doctor_schedules (
			id UUID PRIMARY KEY,
			doctor_id UUID NOT NULL,
			name VARCHAR(255) NOT NULL,
			work_days_json TEXT NOT NULL,
			start_time VARCHAR(5) NOT NULL,
			end_time VARCHAR(5) NOT NULL,
			break_start VARCHAR(5),
			break_end VARCHAR(5),
			slot_duration BIGINT NOT NULL DEFAULT 30,
			slot_title VARCHAR(255),
			appointment_format VARCHAR(10) NOT NULL DEFAULT 'offline',
			is_active BOOLEAN DEFAULT true,
			created_at TIMESTAMP WITH TIME ZONE,
			updated_at TIMESTAMP WITH TIME ZONE
		)
	`).Error; err != nil {
		return fmt.Errorf("failed to create doctor_schedules table: %w", err)
	}

	// Создаем таблицу schedule_exceptions
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schedule_exceptions (
			id UUID PRIMARY KEY,
			doctor_id UUID NOT NULL,
			date DATE NOT NULL,
			type VARCHAR(20) NOT NULL,
			custom_start_time VARCHAR(5),
			custom_end_time VARCHAR(5),
			reason VARCHAR(255),
			created_at TIMESTAMP WITH TIME ZONE,
			updated_at TIMESTAMP WITH TIME ZONE
		)
	`).Error; err != nil {
		return fmt.Errorf("failed to create schedule_exceptions table: %w", err)
	}

	// Создаем таблицу appointments
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS appointments (
			id UUID PRIMARY KEY,
			start_time TIMESTAMP WITH TIME ZONE NOT NULL,
			end_time TIMESTAMP WITH TIME ZONE NOT NULL,
			doctor_id UUID NOT NULL,
			patient_id UUID,
			title VARCHAR(255) NOT NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'available',
			appointment_type VARCHAR(10) DEFAULT 'offline',
			meeting_link TEXT,
			meeting_id VARCHAR(100),
			patient_notes TEXT,
			doctor_notes TEXT,
			schedule_id UUID,
			created_at TIMESTAMP WITH TIME ZONE,
			updated_at TIMESTAMP WITH TIME ZONE
		)
	`).Error; err != nil {
		return fmt.Errorf("failed to create appointments table: %w", err)
	}

	// Создаем индексы
	if err := CreateIndexes(db); err != nil {
		log.Printf("Warning: Failed to create indexes: %v", err)
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// checkAndFixScheduleTable - проверяет и исправляет структуру таблицы doctor_schedules
func checkAndFixScheduleTable(db *gorm.DB) error {
	log.Println("Checking doctor_schedules table structure...")

	// Проверяем существует ли таблица
	var tableExists bool
	err := db.Raw("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = CURRENT_SCHEMA() AND table_name = $1)", "doctor_schedules").Scan(&tableExists).Error
	if err != nil {
		return fmt.Errorf("failed to check table existence: %w", err)
	}

	if !tableExists {
		log.Println("Table doctor_schedules doesn't exist yet, will be created by AutoMigrate")
		return nil
	}

	// Проверяем есть ли проблемный столбец work_days (integer[])
	var workDaysColumnExists bool
	err = db.Raw("SELECT EXISTS (SELECT FROM information_schema.columns WHERE table_schema = CURRENT_SCHEMA() AND table_name = $1 AND column_name = $2)", "doctor_schedules", "work_days").Scan(&workDaysColumnExists).Error
	if err != nil {
		return fmt.Errorf("failed to check work_days column: %w", err)
	}

	// Проверяем есть ли правильный столбец work_days_json
	var workDaysJsonColumnExists bool
	err = db.Raw("SELECT EXISTS (SELECT FROM information_schema.columns WHERE table_schema = CURRENT_SCHEMA() AND table_name = $1 AND column_name = $2)", "doctor_schedules", "work_days_json").Scan(&workDaysJsonColumnExists).Error
	if err != nil {
		return fmt.Errorf("failed to check work_days_json column: %w", err)
	}

	if workDaysColumnExists {
		log.Println(" CRITICAL: Found old work_days column (integer[]) in doctor_schedules table")
		log.Println("This column conflicts with the new structure and must be removed")
		log.Println("Please run the following SQL commands manually to fix this:")
		log.Println("docker exec -it vitalem_postgres psql -U vitalem_user -d vitalem_db")
		log.Println("")
		log.Println("-- If you want to preserve data:")
		log.Println("ALTER TABLE doctor_schedules ADD COLUMN work_days_json TEXT;")
		log.Println("UPDATE doctor_schedules SET work_days_json = (SELECT json_agg(unnest)::text FROM unnest(work_days)) WHERE work_days IS NOT NULL;")
		log.Println("UPDATE doctor_schedules SET work_days_json = '[]' WHERE work_days_json IS NULL;")
		log.Println("ALTER TABLE doctor_schedules ALTER COLUMN work_days_json SET NOT NULL;")
		log.Println("ALTER TABLE doctor_schedules DROP COLUMN work_days;")
		log.Println("")
		log.Println("-- Or if you want to recreate the table from scratch:")
		log.Println(" DROP TABLE doctor_schedules CASCADE;")
		log.Println("")

		return fmt.Errorf("table doctor_schedules contains incompatible work_days column - manual intervention required")
	}

	if !workDaysJsonColumnExists {
		log.Println("Table structure looks compatible, work_days_json will be created by AutoMigrate")
	} else {
		log.Println("Table structure is correct, work_days_json column exists")

		// Пробуем выполнить простой RAW SQL запрос
		log.Println("Testing raw SQL query on doctor_schedules...")
		var count int64
		err = db.Raw("SELECT COUNT(*) FROM doctor_schedules").Scan(&count).Error
		if err != nil {
			log.Printf("Raw SQL query failed: %v", err)
			return fmt.Errorf("raw SQL query failed: %w", err)
		}
		log.Printf("Raw SQL query successful, found %d records", count)
	}

	return nil
}

// CreateIndexes - создание индексов для оптимизации
func CreateIndexes(db *gorm.DB) error {
	log.Println("Creating database indexes...")

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
			log.Printf("Failed to create index: %s - %v", indexSQL, err)
			// Продолжаем выполнение, индексы не критичны
		}
	}

	log.Println("Database indexes created successfully")
	return nil
}
