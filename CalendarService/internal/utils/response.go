package utils

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func JSONSuccess(c echo.Context, data interface{}) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    data,
	})
}

func JSONError(c echo.Context, status int, message string) error {
	return c.JSON(status, map[string]interface{}{
		"success": false,
		"error":   message,
	})
}
