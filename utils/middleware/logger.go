package middleware

import (
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
)

// LoggerMiddleware возвращает middleware для логирования с исключением health check
func LoggerMiddleware() echo.MiddlewareFunc {
	return echomiddleware.LoggerWithConfig(echomiddleware.LoggerConfig{
		// Пропускаем логирование для health check эндпоинтов
		Skipper: func(c echo.Context) bool {
			path := c.Request().URL.Path
			return path == "/health" // Пропускаем логирование для /health
		},
		Format: `{"time":"${time_rfc3339_nano}","id":"${id}","remote_ip":"${remote_ip}",` +
			`"host":"${host}","method":"${method}","uri":"${uri}",` +
			`"user_agent":"${user_agent}","status":${status},"error":"${error}",` +
			`"latency":${latency},"latency_human":"${latency_human}",` +
			`"bytes_in":${bytes_in},"bytes_out":${bytes_out}}` + "\n",
	})
}
