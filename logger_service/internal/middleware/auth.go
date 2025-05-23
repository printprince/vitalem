package middleware

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	"net/http"
	"strings"
)

// JWTMiddleware - Middleware для проверки JWT токена
// Используется для защиты маршрутов, требующих аутентификации и авторизации
// Он проверяет наличие и валидность JWT токена в заголовке Authorization
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
			// Парсим и проверяем токен
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, err := token.Method.(*jwt.SigningMethodHMAC); !err {
					return nil, echo.NewHTTPError(http.StatusUnauthorized, "Invalid token signing method")
				}
				// Возвращаем секретный ключ для проверки подписи
				return []byte(jwtSecret), nil
			})

			// Проверяем наличие ошибок и валидность токена
			if err != nil || !token.Valid {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid or expired token")
			}

			// Извлекаем данные из токена и добавляем их в контекст
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token claims")
			}

			// сохраняем ID, email и роль в контексте
			c.Set("user_id", uint(claims["user_id"].(float64)))
			c.Set("role", claims["role"])
			c.Set("email", claims["email"])

			// Вызываем следующий обработчик
			return next(c)
		}
	}
}
