package service

import (
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
	"identity_service/internal/models"
	"identity_service/internal/repository"
	"log"
	"time"
)

type AuthService struct {
	userRepository *repository.UserRepository
	jwtSecret      string
	jwtExpire      int
}

func NewAuthService(userRepository *repository.UserRepository, jwtSecret string, jwtExpire int) *AuthService {
	return &AuthService{
		userRepository: userRepository,
		jwtSecret:      jwtSecret,
		jwtExpire:      jwtExpire,
	}
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
	return s.userRepository.Create(user)
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
