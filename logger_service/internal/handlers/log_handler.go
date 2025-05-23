package handlers

import (
	"log/slog"
	"net/http"

	"github.com/printprince/vitalem/logger_service/internal/models"
	"github.com/printprince/vitalem/logger_service/internal/service"

	"github.com/labstack/echo/v4"
)

type LogHandler struct {
	logService *service.LogService
	logger     *slog.Logger
}

type createLogRequest struct {
	Service  string                 `json:"service" validate:"required"`
	Level    string                 `json:"level" validate:"required, oneof=debug info warn error"`
	Message  string                 `json:"message" validate:"required"`
	Metadata map[string]interface{} `json:"metadata"`
}

func NewLogHandler(logService *service.LogService, logger *slog.Logger) *LogHandler {
	return &LogHandler{
		logService: logService,
		logger:     logger,
	}
}

func (h *LogHandler) CreateLog(c echo.Context) error {
	// Создаем переменную с запросом если формат верный
	var req createLogRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Error("failed to bind request", "error", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	// Создаем переменную с уровнем логирования если формат верный
	var level models.LogLevel
	switch req.Level {
	case "debug":
		level = models.LogLevelDebug
	case "info":
		level = models.LogLevelInfo
	case "warn":
		level = models.LogLevelWarn
	case "error":
		level = models.LogLevelError
	default:
		h.logger.Error("Invalid log level", "level", req.Level)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid log level")
	}

	err := h.logService.Log(req.Service, level, req.Message, req.Metadata)
	if err != nil {
		h.logger.Error("Failed to create log", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create log")
	}

	// Если все нормально то возвращаем статус создан и сообщение об успехе
	return c.JSON(http.StatusCreated, map[string]string{"status": "success"})
}

func RegisterRoutes(e *echo.Echo, logService *service.LogService, logger *slog.Logger) {
	handler := NewLogHandler(logService, logger)

	// Маршруты
	e.POST("/logs", handler.CreateLog)
}
