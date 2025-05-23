package handlers

import (
	"identity_service/internal/middleware"
	"identity_service/internal/service"
	"net/http"
	"strings"

	"github.com/printprince/vitalem/logger_service/pkg/logger"

	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	authService *service.AuthService
	logger      *logger.Client
}

func NewAuthHandler(authService *service.AuthService, logger *logger.Client) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger,
	}
}

type loginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}
type registerRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Role     string `json:"role" validate:"required,oneof=user, doctor"`
}

type authResponse struct {
	Token string `json:"token"`
}

func (h *AuthHandler) Login(c echo.Context) error {
	// Создаем запрос на вход
	var req loginRequest
	if err := c.Bind(&req); err != nil {
		if h.logger != nil {
			h.logger.Error("Invalid login request format", map[string]interface{}{
				"error": err.Error(),
			})
		}
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	// Пробуем залогиниться
	token, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		if h.logger != nil {
			h.logger.Error("Login failed", map[string]interface{}{
				"email": req.Email,
				"error": err.Error(),
			})
		}

		if err.Error() == "invalid email or password" {
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid email or password")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if h.logger != nil {
		h.logger.Info("User logged in", map[string]interface{}{
			"email": req.Email,
		})
	}

	// Возвращаем токен в ответе
	return c.JSON(http.StatusOK, authResponse{Token: token})
}

func (h *AuthHandler) Register(c echo.Context) error {
	// Создаем запрос на регистрацию
	var req registerRequest
	if err := c.Bind(&req); err != nil {
		if h.logger != nil {
			h.logger.Error("Invalid register request format", map[string]interface{}{
				"error": err.Error(),
			})
		}
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	// Пробуем зарегистрироваться
	err := h.authService.Register(req.Email, req.Password, req.Role)
	if err != nil {
		if h.logger != nil {
			h.logger.Error("Registration failed", map[string]interface{}{
				"email": req.Email,
				"role":  req.Role,
				"error": err.Error(),
			})
		}

		if err.Error() == "User with this email already exists" {
			return echo.NewHTTPError(http.StatusBadRequest, "User with this email already exists")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if h.logger != nil {
		h.logger.Info("User registered", map[string]interface{}{
			"email": req.Email,
			"role":  req.Role,
		})
	}

	// Возвращаем ответ с успешной регистрацией
	return c.JSON(http.StatusCreated, map[string]string{"message": "User created successfully"})
}

func (h *AuthHandler) ValidateToken(c echo.Context) error {
	// Проверяем наличие заголовка Authorization
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		if h.logger != nil {
			h.logger.Warn("Missing Authorization header", nil)
		}
		return echo.NewHTTPError(http.StatusUnauthorized, "Missing Authorization header")
	}

	// Проверяем формат заголовка
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		if h.logger != nil {
			h.logger.Warn("Invalid Authorization header format", nil)
		}
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid Authorization header format, should be 'Bearer [token]'")
	}

	// Получаем токен из заголовка
	tokenString := parts[1]

	// Парсим и проверяем токен
	claims, err := h.authService.ValidateToken(tokenString)
	if err != nil {
		if h.logger != nil {
			h.logger.Error("Invalid token", map[string]interface{}{
				"error": err.Error(),
			})
		}
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token: "+err.Error())
	}

	if h.logger != nil {
		h.logger.Info("Token validated", map[string]interface{}{
			"user_id": claims.UserID,
			"email":   claims.Email,
			"role":    claims.Role,
		})
	}

	// Возвращаем данные из токена
	return c.JSON(http.StatusOK, map[string]interface{}{
		"valid":   true,
		"user_id": claims.UserID,
		"email":   claims.Email,
		"role":    claims.Role,
		"expire":  claims.ExpiresAt,
	})
}

func (h *AuthHandler) GetUser(c echo.Context) error {
	userID := c.Get("user_id")

	if h.logger != nil {
		h.logger.Info("User data requested", map[string]interface{}{
			"user_id": userID,
			"email":   c.Get("email"),
			"role":    c.Get("role"),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"user_id": userID,
		"email":   c.Get("email"),
		"role":    c.Get("role"),
	})
}

func RegisterRoutes(e *echo.Echo, authService *service.AuthService, logger *logger.Client) {
	handler := NewAuthHandler(authService, logger)

	// Публичные маршруты
	// Все маршруты, которые не требуют аутентификации
	e.POST("/login", handler.Login)
	e.POST("/register", handler.Register)
	e.GET("/validate-token", handler.ValidateToken)

	// Защищенные маршруты
	// Маршруты, которые требуют аутентификации
	protected := e.Group("/protected")
	protected.Use(middleware.JWTMiddleware(authService.GetJWTSecret()))
	protected.GET("/user", handler.GetUser)
}
