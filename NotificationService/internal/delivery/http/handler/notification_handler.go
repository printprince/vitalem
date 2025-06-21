package handler

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"NotificationService/internal/domain/models"
	"NotificationService/internal/service"
	"NotificationService/internal/utils"
)

type NotificationHandler struct {
	service service.NotificationService
}

func NewNotificationHandler(s service.NotificationService) *NotificationHandler {
	return &NotificationHandler{service: s}
}

// RegisterRoutes регистрирует маршруты в Echo группе
func (h *NotificationHandler) RegisterRoutes(g *echo.Group) {
	g.POST("/notifications", h.Create)
	g.GET("/notifications/:id", h.GetByID)
	g.GET("/notifications/recipient/:recipientId", h.ListByRecipient)
	g.PUT("/notifications/:id/sent", h.MarkAsSent)
	g.GET("/notifications/my", h.GetMyNotifications)
}

// CreateNotificationRequest — структура запроса для создания уведомления
type CreateNotificationRequest struct {
	Type        models.NotificationType    `json:"type" binding:"required"`
	Channel     models.NotificationChannel `json:"channel" binding:"required"`
	RecipientID uuid.UUID                  `json:"recipientId" binding:"required"`
	Recipient   string                     `json:"recipient" binding:"required"`
	Message     string                     `json:"message,omitempty"`
	Metadata    interface{}                `json:"metadata,omitempty"`
}

// Create создает новое уведомление
func (h *NotificationHandler) Create(c echo.Context) error {
	var req CreateNotificationRequest

	if err := c.Bind(&req); err != nil {
		return utils.JSONError(c, http.StatusBadRequest, "invalid request body")
	}

	// Проверка обязательных полей (но message может быть пустым!)
	if req.RecipientID == uuid.Nil || req.Recipient == "" || req.Type == "" || req.Channel == "" {
		return utils.JSONError(c, http.StatusBadRequest, "missing required fields")
	}

	// Создаем уведомление
	notification := &models.Notification{
		Type:        req.Type,
		Channel:     req.Channel,
		RecipientID: req.RecipientID,
		Recipient:   req.Recipient,
		Message:     req.Message,
	}

	// Устанавливаем метаданные если переданы
	if req.Metadata != nil {
		if err := notification.SetMetadata(req.Metadata); err != nil {
			return utils.JSONError(c, http.StatusBadRequest, "invalid metadata format")
		}
	}

	if err := h.service.Send(c.Request().Context(), notification); err != nil {
		return utils.JSONError(c, http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, notification)
}

// GetByID возвращает уведомление по ID
func (h *NotificationHandler) GetByID(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return utils.JSONError(c, http.StatusBadRequest, "invalid id")
	}

	n, err := h.service.Get(c.Request().Context(), id)
	if err != nil {
		return utils.JSONError(c, http.StatusNotFound, err.Error())
	}

	return c.JSON(http.StatusOK, n)
}

// ListByRecipient возвращает уведомления по recipientID
func (h *NotificationHandler) ListByRecipient(c echo.Context) error {
	recipientStr := c.Param("recipientId")
	recipientID, err := uuid.Parse(recipientStr)
	if err != nil {
		return utils.JSONError(c, http.StatusBadRequest, "invalid recipient id")
	}

	list, err := h.service.List(c.Request().Context(), recipientID)
	if err != nil {
		return utils.JSONError(c, http.StatusNotFound, err.Error())
	}

	return c.JSON(http.StatusOK, list)
}

// MarkAsSent помечает уведомление как отправленное
func (h *NotificationHandler) MarkAsSent(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return utils.JSONError(c, http.StatusBadRequest, "invalid id")
	}

	if err := h.service.MarkAsSent(c.Request().Context(), id); err != nil {
		return utils.JSONError(c, http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}

// GetMyNotifications возвращает уведомления текущего пользователя
func (h *NotificationHandler) GetMyNotifications(c echo.Context) error {
	// Получаем ID пользователя из JWT токена
	userIDVal := c.Get("user_id")
	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		return utils.JSONError(c, http.StatusUnauthorized, "unauthorized: invalid user ID")
	}

	list, err := h.service.List(c.Request().Context(), userID)
	if err != nil {
		return utils.JSONError(c, http.StatusNotFound, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"user_id":       userID,
		"total_count":   len(list),
		"notifications": list,
	})
}
