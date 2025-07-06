package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/printprince/vitalem/gateway_service/internal/config"
)

// ProxyHandler структура для проксирования запросов
type ProxyHandler struct {
	config     *config.Config
	httpClient *http.Client
}

// NewProxyHandler создает новый proxy handler
func NewProxyHandler(cfg *config.Config) *ProxyHandler {
	return &ProxyHandler{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ProxyToIdentity проксирует запросы к Identity Service
func (p *ProxyHandler) ProxyToIdentity(c echo.Context) error {
	return p.proxyRequest(c, "identity")
}

// ProxyToPatient проксирует запросы к Patient Service
func (p *ProxyHandler) ProxyToPatient(c echo.Context) error {
	return p.proxyRequest(c, "patient")
}

// ProxyToSpecialist проксирует запросы к Specialist Service
func (p *ProxyHandler) ProxyToSpecialist(c echo.Context) error {
	return p.proxyRequest(c, "specialist")
}

// ProxyToAppointment проксирует запросы к Appointment Service
func (p *ProxyHandler) ProxyToAppointment(c echo.Context) error {
	return p.proxyRequest(c, "appointment")
}

// ProxyToNotification проксирует запросы к Notification Service
func (p *ProxyHandler) ProxyToNotification(c echo.Context) error {
	return p.proxyRequest(c, "notification")
}

// ProxyToFileServer проксирует запросы к FileServer Service
func (p *ProxyHandler) ProxyToFileServer(c echo.Context) error {
	return p.proxyRequest(c, "fileserver")
}

// ProxyToLogger проксирует запросы к Logger Service
func (p *ProxyHandler) ProxyToLogger(c echo.Context) error {
	return p.proxyRequest(c, "logger")
}

// proxyRequest основная логика проксирования
func (p *ProxyHandler) proxyRequest(c echo.Context, serviceName string) error {
	// Получаем URL сервиса
	serviceURL := p.config.GetServiceURL(serviceName)
	if serviceURL == "" {
		return echo.NewHTTPError(http.StatusServiceUnavailable,
			fmt.Sprintf("Service %s not configured", serviceName))
	}

	// Строим целевой URL
	targetURL := p.buildTargetURL(c, serviceURL, serviceName)

	// Создаем новый запрос
	req, err := http.NewRequest(c.Request().Method, targetURL, c.Request().Body)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			fmt.Sprintf("Failed to create request: %v", err))
	}

	// Копируем заголовки
	p.copyHeaders(c.Request().Header, req.Header)

	// Выполняем запрос
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadGateway,
			fmt.Sprintf("Failed to proxy request: %v", err))
	}
	defer resp.Body.Close()

	// Копируем заголовки ответа
	p.copyHeaders(resp.Header, c.Response().Header())

	// Устанавливаем статус код
	c.Response().WriteHeader(resp.StatusCode)

	// Копируем тело ответа
	_, err = io.Copy(c.Response().Writer, resp.Body)
	return err
}

// buildTargetURL строит целевой URL для проксирования
func (p *ProxyHandler) buildTargetURL(c echo.Context, serviceURL, serviceName string) string {
	path := c.Request().URL.Path

	// Преобразуем простые Gateway пути в пути ожидаемые сервисами
	switch serviceName {
	case "identity":
		// Gateway: /auth/* → Identity: /auth/*
		// Без изменений, Identity Service уже использует /auth/*

	case "patient":
		// Gateway: /patients/* → Patient: /api/v1/patients/*
		if strings.HasPrefix(path, "/patients") {
			path = "/api/v1" + path
		} else if strings.HasPrefix(path, "/users/") && strings.Contains(path, "/patient") {
			// Gateway: /users/:userID/patient/* → Patient: /api/v1/users/:userID/patient/*
			path = "/api/v1" + path
		}

	case "specialist":
		// Gateway: /doctors/* → Specialist: /api/doctors/* (публичные)
		// Gateway: /doctors/* → Specialist: /api/v1/doctors/* (защищенные)
		if strings.HasPrefix(path, "/doctors") {
			// Если у нас есть JWT токен в заголовках, то это защищенный роут
			auth := c.Request().Header.Get("Authorization")
			if auth != "" && strings.HasPrefix(auth, "Bearer ") {
				path = "/api/v1" + path // Защищенные роуты
			} else {
				path = "/api" + path // Публичные роуты
			}
		} else if strings.HasPrefix(path, "/users/") && strings.Contains(path, "/doctor") {
			// Gateway: /users/:userID/doctor/* → Specialist: /api/v1/users/:userID/doctor/*
			path = "/api/v1" + path
		}

	case "appointment":
		// Gateway: /appointments/* → Appointment: /api/*
		if strings.HasPrefix(path, "/appointments") {
			// Специальная обработка для schedules (врачи)
			if strings.Contains(path, "/schedules") {
				path = strings.Replace(path, "/appointments/schedules", "/api/doctor/schedules", 1)
			} else if strings.Contains(path, "/exceptions") {
				// /appointments/exceptions → /api/doctor/exceptions
				path = strings.Replace(path, "/appointments/exceptions", "/api/doctor/exceptions", 1)
			} else if strings.Contains(path, "/doctors/") && strings.Contains(path, "/available-slots") {
				// /appointments/doctors/123/available-slots → /api/doctors/123/available-slots
				path = strings.Replace(path, "/appointments/doctors", "/api/doctors", 1)
			} else if strings.Contains(path, "/book") {
				// /appointments/123/book → /api/appointments/123/book
				path = strings.Replace(path, "/appointments", "/api/appointments", 1)
			} else if strings.Contains(path, "/cancel") {
				// /appointments/123/cancel → /api/patient/appointments/123/cancel
				appointmentID := strings.TrimPrefix(path, "/appointments/")
				appointmentID = strings.TrimSuffix(appointmentID, "/cancel")
				path = "/api/patient/appointments/" + appointmentID + "/cancel"
			} else {
				// Остальные /appointments/* → /api/*
				path = strings.Replace(path, "/appointments", "/api", 1)
			}
		}

	case "notification":
		// Gateway: /notifications/* → Notification: /api/v1/notifications/* или /notifications/*
		if strings.HasPrefix(path, "/notifications") {
			if strings.HasPrefix(path, "/notifications/my") {
				// /notifications/my → /notifications/my (прямой маршрут)
				// Оставляем как есть
			} else if strings.Contains(path, "/sent") {
				// /notifications/123/sent → /notifications/123/sent (прямой маршрут)
				// Оставляем как есть
			} else if strings.Contains(path, "/recipient/") {
				// /notifications/recipient/123 → /api/v1/notifications/recipient/123
				path = "/api/v1" + path
			} else {
				// Остальные /notifications/* → /api/v1/notifications/*
				path = "/api/v1" + path
			}
		}

	case "fileserver":
		// Gateway: /files/* → FileServer: /files/*
		// Gateway: /public/* → FileServer: /public/*
		// Без изменений, FileServer уже использует эти пути

	case "logger":
		// Gateway: /logs/* → Logger: /logs/*
		// Без изменений, Logger Service уже использует /logs
	}

	// Добавляем query parameters если есть
	query := c.Request().URL.RawQuery
	if query != "" {
		return fmt.Sprintf("%s%s?%s", serviceURL, path, query)
	}

	return fmt.Sprintf("%s%s", serviceURL, path)
}

// copyHeaders копирует HTTP заголовки
func (p *ProxyHandler) copyHeaders(src, dst http.Header) {
	for key, values := range src {
		// Пропускаем некоторые системные заголовки
		if p.shouldSkipHeader(key) {
			continue
		}

		for _, value := range values {
			dst.Add(key, value)
		}
	}
}

// shouldSkipHeader проверяет нужно ли пропустить заголовок
func (p *ProxyHandler) shouldSkipHeader(header string) bool {
	// Список заголовков которые не должны проксироваться
	skipHeaders := []string{
		"Content-Length",
		"Transfer-Encoding",
		"Connection",
		"Upgrade",
		"Proxy-Authenticate",
		"Proxy-Authorization",
	}

	headerLower := strings.ToLower(header)
	for _, skip := range skipHeaders {
		if headerLower == strings.ToLower(skip) {
			return true
		}
	}

	return false
}
