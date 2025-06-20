package middleware

import (
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// min возвращает минимальное из двух чисел
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// JWTMiddleware - Middleware для проверки JWT токена
func JWTMiddleware(jwtSecret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Получаем заголовок Authorization и проверяем его наличие
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "Missing Authorization header")
			}

			// Проверяем формат заголовка
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid Authorization header format")
			}

			// Получаем токен из заголовка
			tokenString := parts[1]

			// Debug: логируем информацию о токене
			c.Logger().Infof("DEBUG JWT: Validating token")
			c.Logger().Infof("DEBUG JWT: Token preview: %s...", tokenString[:min(20, len(tokenString))])
			c.Logger().Infof("DEBUG JWT: Secret length: %d", len(jwtSecret))
			c.Logger().Infof("DEBUG JWT: Secret preview: %s...", jwtSecret[:min(8, len(jwtSecret))])

			// Парсим и проверяем токен
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				// Проверяем алгоритм токена
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, echo.NewHTTPError(http.StatusUnauthorized, "Invalid token signing method")
				}

				// Возвращаем секретный ключ для проверки подписи
				return []byte(jwtSecret), nil
			})

			// Проверяем наличие ошибок и валидность токена
			if err != nil {
				c.Logger().Errorf("DEBUG JWT: Parse error: %v", err)
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid or expired token")
			}

			if !token.Valid {
				c.Logger().Errorf("DEBUG JWT: Token is not valid")
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid or expired token")
			}

			// Извлекаем данные из токена и добавляем их в контекст
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token claims")
			}

			// Получаем ID пользователя и преобразуем его в UUID
			var userID uuid.UUID

			// Проверяем тип ID пользователя в токене
			switch id := claims["user_id"].(type) {
			case string:
				// Если строка, парсим её как UUID
				userID, err = uuid.Parse(id)
				if err != nil {
					return echo.NewHTTPError(http.StatusUnauthorized, "Invalid user_id format in token")
				}
			default:
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid user_id type in token")
			}

			// Сохраняем данные в контексте
			c.Set("user_id", userID)
			c.Set("role", claims["role"])
			c.Set("email", claims["email"])

			// Вызываем следующий обработчик
			return next(c)
		}
	}
}
