package middleware

import (
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// JWTMiddleware - Продвинутый мидлварь для авторизации запросов через JWT
// Валидирует токены и выдергивает из них нужную инфу (юзер айди, роль, емейл)
// Прокидывает эти данные в контекст реквеста для дальнейшего использования
// в обработчиках. Позволяет разрулить доступы к ручкам в зависимости от роли.
func JWTMiddleware(jwtSecret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Выдёргиваем хедер Auth из запроса
			// Без него вообще не пускаем дальше, сразу 401
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "Missing Authorization header")
			}

			// Валидируем формат хедера - должен быть "Bearer <token>"
			// Без этого префикса тоже не пропускаем, чтобы не словить неожиданных багов
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid Authorization header format")
			}

			// Вытаскиваем сам токен из хедера
			tokenString := parts[1]

			// Парсим и валидируем JWT токен
			// Кастомная функция проверки подписи - фулпруф от подмены алгоритма
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				// Убеждаемся, что алгоритм токена именно HMAC
				// Хак-протекшн от атак с подменой алгоритма (none -> HS256)
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, echo.NewHTTPError(http.StatusUnauthorized, "Invalid token signing method")
				}

				// Возвращаем наш секретный ключ для проверки подписи
				return []byte(jwtSecret), nil
			})

			// Если токен невалидный или истёк - отбриваем с 401
			if err != nil || !token.Valid {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid or expired token")
			}

			// Выцепляем клеймы из токена и конвертим в удобный формат
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token claims")
			}

			// Конвертим user_id из токена в UUID для бэкенда
			// Здесь важно обработать все возможные кейсы типов данных
			var userID uuid.UUID
			switch id := claims["user_id"].(type) {
			case string:
				// Парсим строковый UUID в бинарный
				// Это критично для корректной работы с GORM/Postgres
				userID, err = uuid.Parse(id)
				if err != nil {
					return echo.NewHTTPError(http.StatusUnauthorized, "Invalid user_id format in token")
				}
			default:
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid user_id type in token")
			}

			// Пробрасываем распарсенные данные юзера в контекст
			// Эти данные будут доступны в дальнейших хендлерах
			c.Set("user_id", userID)
			c.Set("role", claims["role"])
			c.Set("email", claims["email"])

			// Передаём управление следующему хендлеру в цепочке
			return next(c)
		}
	}
}
