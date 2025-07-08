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

	// Теперь сервисы используют простые пути, поэтому минимальные преобразования
	switch serviceName {
	case "identity":
		// Gateway: /auth/* → Identity: /auth/*
		// Без изменений, Identity Service использует /auth/*

	case "patient":
		// Gateway: /patients/* → Patient: /patients/*
		// Gateway: /users/:userID/patient/* → Patient: /users/:userID/patient/*
		// Без изменений, Patient Service теперь использует простые пути

	case "specialist":
		// Gateway: /doctors/* → Specialist: /doctors/* (защищенные)
		// Gateway: /doctors/* → Specialist: /api/doctors/* (публичные)
		if strings.HasPrefix(path, "/doctors") {
			// Если у нас есть JWT токен в заголовках, то это защищенный роут
			auth := c.Request().Header.Get("Authorization")
			if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
				// Публичные роуты остаются с префиксом /api
				path = "/api" + path
			}
			// Защищенные роуты без изменений
		}
		// Gateway: /users/:userID/doctor/* → Specialist: /users/:userID/doctor/*
		// Без изменений

	case "appointment":
		// Gateway: /appointments/* → Appointment: /appointments/*
		// Без изменений, Appointment Service теперь использует простые пути

	case "notification":
		// Gateway: /notifications/* → Notification: /notifications/*
		// Без изменений, Notification Service теперь использует простые пути

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
