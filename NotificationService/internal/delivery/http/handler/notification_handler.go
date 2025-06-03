package handler

import (
	"net/http"
	"strconv"

	"NotificationService/internal/domain/models"
	"NotificationService/internal/service"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type NotificationHandler struct {
	service service.NotificationService
}

func NewNotificationHandler(service service.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		service: service,
	}
}

// Send обрабатывает запрос на отправку уведомления
func (h *NotificationHandler) Send(c echo.Context) error {
	var notification models.Notification
	if err := c.Bind(&notification); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	if err := h.service.Send(c.Request().Context(), &notification); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, notification)
}

// Get обрабатывает запрос на получение уведомления по ID
func (h *NotificationHandler) Get(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid notification ID",
		})
	}

	notification, err := h.service.Get(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Notification not found",
		})
	}

	return c.JSON(http.StatusOK, notification)
}

// List обрабатывает запрос на получение списка уведомлений для получателя
func (h *NotificationHandler) List(c echo.Context) error {
	recipientID, err := uuid.Parse(c.Param("recipient_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid recipient ID",
		})
	}

	notifications, err := h.service.List(c.Request().Context(), recipientID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, notifications)
}

// MarkAsSent обрабатывает запрос на отметку уведомления как отправленного
func (h *NotificationHandler) MarkAsSent(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid notification ID",
		})
	}

	if err := h.service.MarkAsSent(c.Request().Context(), id); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.NoContent(http.StatusNoContent)
}
