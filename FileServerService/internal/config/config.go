package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

func (db *DBConfig) DSN() string {
	return "host=" + db.Host +
		" port=" + db.Port +
		" user=" + db.User +
		" password=" + db.Password +
		" dbname=" + db.Name +
		" sslmode=disable"
}

type MinIOConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	UseSSL    bool
}

type Config struct {
	Env        string // 👈 добавили это поле
	ServerPort string
	DB         DBConfig
	MinIO      MinIOConfig

	JWTSecret string

	// Опциональные таймауты
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// Load загружает конфигурацию из .env или переменных окружения
func Load() *Config {
	// Загрузить .env (если есть)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	useSSL := false
	if val := os.Getenv("MINIO_USE_SSL"); val != "" {
		b, err := strconv.ParseBool(val)
		if err == nil {
			useSSL = b
		}
	}

	readTimeout := 10 * time.Second
	writeTimeout := 10 * time.Second
	idleTimeout := 60 * time.Second

	// Можно добавить чтение таймаутов из окружения по желанию

	return &Config{
		Env:        getEnv("APP_ENV", "development"), // 👈 добавлено
		ServerPort: getEnv("SERVER_PORT", "8080"),
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "12345"),
			Name:     getEnv("DB_NAME", "fileserver_db"),
		},
		MinIO: MinIOConfig{
			Endpoint:  getEnv("MINIO_ENDPOINT", "localhost:9000"),
			AccessKey: getEnv("MINIO_ACCESS_KEY", "minioadmin"),
			SecretKey: getEnv("MINIO_SECRET_KEY", "minioadmin"),
			UseSSL:    useSSL,
		},

		JWTSecret: getEnv("JWT_SECRET", "4324pkh23sk4jh342alhdlfl2sdjf"),

		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
