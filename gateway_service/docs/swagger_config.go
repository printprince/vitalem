package docs

import (
	"net/http"

	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
)

// SwaggerConfig содержит настройки для Swagger UI
type SwaggerConfig struct {
	Title       string `json:"title"`
	Version     string `json:"version"`
	Description string `json:"description"`
	BasePath    string `json:"basePath"`
	Host        string `json:"host"`
}

// GetSwaggerConfig возвращает конфигурацию Swagger
func GetSwaggerConfig() SwaggerConfig {
	return SwaggerConfig{
		Title:       "Vitalem API Gateway",
		Version:     "1.0.0",
		Description: "Единая точка входа для всех микросервисов медицинской платформы Vitalem",
		BasePath:    "/",
		Host:        "localhost:8800",
	}
}

// SetupSwaggerRoutes настраивает маршруты для Swagger
func SetupSwaggerRoutes(e *echo.Echo) {
	// Основной Swagger UI
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	// Дополнительные удобные маршруты
	e.GET("/docs", func(c echo.Context) error {
		return c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})

	e.GET("/api-docs", func(c echo.Context) error {
		return c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})

	// JSON схема API
	e.GET("/swagger.json", func(c echo.Context) error {
		return c.File("docs/swagger.json")
	})

	// YAML схема API
	e.GET("/swagger.yaml", func(c echo.Context) error {
		return c.File("docs/swagger.yaml")
	})

	// Информация об API
	e.GET("/api/info", func(c echo.Context) error {
		config := GetSwaggerConfig()
		return c.JSON(http.StatusOK, map[string]interface{}{
			"title":         config.Title,
			"version":       config.Version,
			"description":   config.Description,
			"documentation": "/swagger/",
			"endpoints": map[string]string{
				"swagger_ui":   "/swagger/",
				"swagger_json": "/swagger.json",
				"swagger_yaml": "/swagger.yaml",
				"health":       "/health",
			},
			"features": []string{
				"JWT Authentication",
				"7 Microservices Integration",
				"RESTful API",
				"CORS Support",
				"Request Validation",
				"Error Handling",
				"File Upload/Download",
				"Real-time Notifications",
			},
			"services": []string{
				"Identity Service (Auth)",
				"Patient Service",
				"Specialist Service",
				"Appointment Service",
				"Notification Service",
				"FileServer Service",
				"Logger Service",
			},
		})
	})
}
