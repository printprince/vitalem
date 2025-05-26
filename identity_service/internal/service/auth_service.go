package service

import (
	"context"
	"errors"
	"fmt"
	"identity_service/internal/models"
	"identity_service/internal/repository"
	"log"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepository *repository.UserRepository
	messageService MessageService
	jwtSecret      string
	jwtExpire      int
}

// NewAuthService создает новый сервис аутентификации
func NewAuthService(userRepository *repository.UserRepository, jwtSecret string, jwtExpire int) *AuthService {
	return &AuthService{
		userRepository: userRepository,
		jwtSecret:      jwtSecret,
		jwtExpire:      jwtExpire,
	}
}

// SetMessageService устанавливает сервис сообщений
func (s *AuthService) SetMessageService(messageService MessageService) {
	s.messageService = messageService
}

// Register - сервис создания аккаунта
func (s *AuthService) Register(email, password, role string) error {
	// Проверка емайла на существование
	existing, _ := s.userRepository.FindByEmail(email)
	if existing != nil {
		return errors.New("User with this email already exists")
	}

	// Хешируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("could not hash password: %w", err)
	}

	// Создаем структуру и вкладываем туда данные юзера
	user := &models.Users{
		Email:          email,
		HashedPassword: string(hashedPassword),
		Role:           role,
		CreatedAt:      time.Now(),
	}

	// Создаем пользователя
	log.Printf("New user registered with email: %s, role: %s", email, role)
	if err := s.userRepository.Create(user); err != nil {
		return err
	}

	// Если сервис сообщений доступен, публикуем событие создания пользователя
	if s.messageService != nil {
		// Создаем событие
		event := &models.UserCreatedEvent{
			UserID: user.ID.String(),
			Email:  email,
			Role:   role,
		}

		// Публикуем событие
		ctx := context.Background()
		if err := s.messageService.PublishUserCreated(ctx, event); err != nil {
			log.Printf("Failed to publish user created event: %v", err)
			// Не возвращаем ошибку, так как пользователь уже создан
		}
	}

	return nil
}

// Login - сервис авторизации
func (s *AuthService) Login(email, password string) (string, error) {
	// Ищем емайл пользователя
	user, err := s.userRepository.FindByEmail(email)
	if err != nil {
		return "", errors.New("Invalid email")
	}

	// Проверяем пароль
	err = bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(password))
	if err != nil {
		return "", errors.New("Invalid password")
	}

	// Создаем claims с использованием нашей структуры
	claims := &models.TokenClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Duration(s.jwtExpire) * time.Hour).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
	}

	// Создаем токен с указанием claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Подписание токена
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", errors.New("could not generate token")
	}

	log.Printf("User %s logged in successfully", email)
	return tokenString, nil
}

// ValidateToken - сервис для валидации токена
func (s *AuthService) ValidateToken(tokenString string) (*models.TokenClaims, error) {
	// Создаем новый экземпляр TokenClaims
	claims := &models.TokenClaims{}

	// Парсим токен с указанными claims
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Проверяем используемый алгоритм
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Возвращаем ключ для проверки подписи
		return []byte(s.jwtSecret), nil
	})

	// Проверяем ошибки
	if err != nil {
		return nil, err
	}

	// Проверяем валидность токена
	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

func (s *AuthService) GetJWTSecret() string {
	return s.jwtSecret
}
