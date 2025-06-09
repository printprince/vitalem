package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/printprince/vitalem/identity_service/internal/models"
	"github.com/printprince/vitalem/identity_service/internal/repository"

	"github.com/dgrijalva/jwt-go"
	"github.com/printprince/vitalem/logger_service/pkg/logger"
	"golang.org/x/crypto/bcrypt"
)

// TokenClaims - структура данных для JWT токена
// Вся инфа, которую мы засовываем в токен и потом можем получить из него
type TokenClaims struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	ExpiresAt int64  `json:"exp"`
}

// AuthService - сервис аутентификации и авторизации
// Содержит бизнес-логику для работы с пользователями, JWT и хешированием
type AuthService struct {
	userRepository *repository.UserRepository
	messageService MessageService
	jwtSecret      string
	jwtExpire      int
	logger         *logger.Client
}

// NewAuthService - фабрика для создания сервиса аутентификации
// Принимает репозиторий пользователей и настройки JWT
func NewAuthService(userRepository *repository.UserRepository, jwtSecret string, jwtExpire int) *AuthService {
	return &AuthService{
		userRepository: userRepository,
		jwtSecret:      jwtSecret,
		jwtExpire:      jwtExpire,
	}
}

// SetMessageService - установка сервиса сообщений
// Опционально - используется для публикации событий создания пользователя
func (s *AuthService) SetMessageService(messageService MessageService) {
	s.messageService = messageService
}

// SetLogger - устанавливает клиент логирования
func (s *AuthService) SetLogger(logger *logger.Client) {
	s.logger = logger
}

// Register - регистрация нового пользователя
// Создает хеш пароля, записывает в БД и отправляет событие в RabbitMQ
// В случае успеха тригерит создание начального профиля пользователя
func (s *AuthService) Register(email, password, role string) error {
	// Проверяем, не существует ли уже пользователь с таким email
	// Защита от дублей в базе
	existingUser, err := s.userRepository.FindByEmail(email)
	if err == nil && existingUser != nil {
		if s.logger != nil {
			s.logger.Warn("Попытка повторной регистрации", map[string]interface{}{
				"email": email,
			})
		}
		return errors.New("User with this email already exists")
	}

	// Хешируем пароль с помощью bcrypt
	// Используем высокий cost для усиления защиты
	// (по дефолту 10, но можно поднять до 14 для критичных систем)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		if s.logger != nil {
			s.logger.Error("Ошибка хеширования пароля", map[string]interface{}{
				"error": err.Error(),
			})
		}
		return err
	}

	// Создаем нового пользователя
	// ID генерится автоматически в BeforeCreate хуке
	user := &models.Users{
		Email:          email,
		HashedPassword: string(hashedPassword),
		Role:           role,
	}

	// Сохраняем пользователя в базу данных
	if err := s.userRepository.Create(user); err != nil {
		if s.logger != nil {
			s.logger.Error("Ошибка создания пользователя", map[string]interface{}{
				"email": email,
				"error": err.Error(),
			})
		}
		return err
	}

	// Если у нас есть сервис сообщений, отправляем событие создания пользователя
	// Это триггерит создание записей в других сервисах (профили пациента/врача)
	if s.messageService != nil {
		if s.logger != nil {
			s.logger.Info("Отправляем событие создания пользователя", map[string]interface{}{
				"userID": user.ID.String(),
				"email":  user.Email,
				"role":   user.Role,
			})
		}

		// Создаем событие с указателем на структуру, как требует интерфейс
		event := &models.UserCreatedEvent{
			UserID: user.ID.String(),
			Email:  user.Email,
			Role:   user.Role,
		}

		// Создаем контекст для передачи в метод публикации
		ctx := context.Background()

		if err := s.messageService.PublishUserCreated(ctx, event); err != nil {
			// Логируем ошибку, но не фейлим всю операцию
			// Это нормально для систем с eventual consistency
			if s.logger != nil {
				s.logger.Error("Ошибка публикации события создания пользователя", map[string]interface{}{
					"user_id": user.ID,
					"email":   user.Email,
					"error":   err.Error(),
				})
			}
			// Не возвращаем ошибку, т.к. пользователь уже создан
			// В худшем случае профиль придется создать вручную
		} else {
			if s.logger != nil {
				s.logger.Info("Событие создания пользователя успешно отправлено", map[string]interface{}{
					"userID": user.ID.String(),
					"email":  user.Email,
					"role":   user.Role,
				})
			}
		}
	} else {
		if s.logger != nil {
			s.logger.Warn("MessageService не установлен, событие не отправлено", map[string]interface{}{
				"userID": user.ID.String(),
				"email":  user.Email,
				"role":   user.Role,
			})
		}
	}

	if s.logger != nil {
		s.logger.Info("Пользователь успешно зарегистрирован", map[string]interface{}{
			"email": email,
			"role":  role,
		})
	}

	return nil
}

// Login - аутентификация пользователя по email и паролю
// Проверяет существование юзера, валидирует пароль и генерит JWT токен
// Возвращает ошибку если данные неверные или произошел сбой в БД
func (s *AuthService) Login(email, password string) (string, error) {
	// Проверяем, существует ли пользователь с таким email
	// Кейс-сенситив поиск по уникальному индексу
	user, err := s.userRepository.FindByEmail(email)
	if err != nil {
		if s.logger != nil {
			s.logger.Error("Ошибка при поиске пользователя", map[string]interface{}{
				"email": email,
				"error": err.Error(),
			})
		}
		return "", errors.New("Invalid email")
	}

	// Проверяем пароль через bcrypt
	// bcrypt.CompareHashAndPassword - тяжелая операция по CPU
	// Защищает от брутфорса и тайминг-атак
	err = bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(password))
	if err != nil {
		if s.logger != nil {
			s.logger.Warn("Неудачная попытка входа: неверный пароль", map[string]interface{}{
				"email": email,
			})
		}
		return "", errors.New("Invalid password")
	}

	// Генерируем токен после успешной проверки
	// Используем UUID для идентификации пользователя
	token, err := s.generateToken(user.ID.String(), user.Email, user.Role)
	if err != nil {
		if s.logger != nil {
			s.logger.Error("Ошибка генерации токена", map[string]interface{}{
				"email": email,
				"error": err.Error(),
			})
		}
		return "", err
	}

	if s.logger != nil {
		s.logger.Info("Успешный вход пользователя", map[string]interface{}{
			"email": email,
			"role":  user.Role,
		})
	}

	return token, nil
}

// ValidateToken - проверка и декодирование JWT токена
// Полная валидация подписи, срока действия и формата
// Возвращает данные из токена или ошибку если токен невалидный
func (s *AuthService) ValidateToken(tokenString string) (*TokenClaims, error) {
	// Парсим токен
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Проверяем алгоритм подписи
		// Защита от атак с подменой алгоритма (none -> HS256)
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Возвращаем секретный ключ для проверки подписи
		return []byte(s.jwtSecret), nil
	})

	// Проверяем ошибки парсинга
	if err != nil {
		return nil, err
	}

	// Проверяем валидность токена
	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	// Извлекаем claims из токена
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	// Проверяем срок действия токена
	exp, ok := claims["exp"].(float64)
	if !ok {
		return nil, errors.New("invalid expiration time")
	}

	if int64(exp) < time.Now().Unix() {
		return nil, errors.New("token expired")
	}

	// Извлекаем данные из токена
	userID, ok := claims["user_id"].(string)
	if !ok {
		return nil, errors.New("invalid user_id in token")
	}

	email, ok := claims["email"].(string)
	if !ok {
		return nil, errors.New("invalid email in token")
	}

	role, ok := claims["role"].(string)
	if !ok {
		return nil, errors.New("invalid role in token")
	}

	// Возвращаем данные из токена
	return &TokenClaims{
		UserID:    userID,
		Email:     email,
		Role:      role,
		ExpiresAt: int64(exp),
	}, nil
}

// GetJWTSecret - получение секретного ключа для JWT
// Используется в middleware для валидации токенов
func (s *AuthService) GetJWTSecret() string {
	return s.jwtSecret
}

// generateToken - внутренний метод для генерации JWT токена
// Создает signed JWT с полезной нагрузкой из ID, email и роли
func (s *AuthService) generateToken(userID, email, role string) (string, error) {
	// Задаем время истечения токена (текущее время + jwtExpire часов)
	expirationTime := time.Now().Add(time.Duration(s.jwtExpire) * time.Hour).Unix()

	// Создаем claims для токена
	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"role":    role,
		"exp":     expirationTime,
	}

	// Создаем новый токен с указанными claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Подписываем токен секретным ключом
	// HS256 - это симметричный алгоритм шифрования (один ключ для создания и проверки)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
