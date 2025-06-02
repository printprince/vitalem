package utils

import (
	"github.com/labstack/echo/v4"
)

// JSONError отправляет json с ошибкой и кодом
func JSONError(c echo.Context, code int, message string) error {
	return c.JSON(code, map[string]interface{}{
		"error":   true,
		"message": message,
	})
}
