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
}

// Create создает новое уведомление
func (h *NotificationHandler) Create(c echo.Context) error {
	var n models.Notification

	if err := c.Bind(&n); err != nil {
		return utils.JSONError(c, http.StatusBadRequest, "invalid request body")
	}

	// Проверка обязательных полей (но message может быть пустым!)
	if n.RecipientID == uuid.Nil || n.Recipient == "" || n.Type == "" || n.Channel == "" {
		return utils.JSONError(c, http.StatusBadRequest, "missing required fields")
	}

	if err := h.service.Send(c.Request().Context(), &n); err != nil {
		return utils.JSONError(c, http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, n)
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
