package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/printprince/vitalem/FileServerService/internal/http/handler"
	"github.com/printprince/vitalem/FileServerService/internal/http/middleware"
	"github.com/printprince/vitalem/FileServerService/internal/service"
)

func NewRouter(fileService service.FileService, jwtSecret string) http.Handler {
	r := chi.NewRouter()

	r.Get("/health", handler.HealthHandler) // статус сервиса

	// Public routes
	r.Route("/public", func(r chi.Router) {
		fh := handler.NewFileHandler(fileService)

		r.Get("/{id}", fh.PublicDownload) // Публичный доступ к файлам (без авторизации)
	})

	// Protected routes
	r.Route("/files", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(jwtSecret))

		fh := handler.NewFileHandler(fileService)

		r.Post("/", fh.Upload)                       // Загрузить файл
		r.Get("/", fh.List)                          // Получить все файлы пользователя
		r.Get("/{id}", fh.Get)                       // Получить мета-инфу по ID
		r.Delete("/{id}", fh.Delete)                 // Удалить файл
		r.Get("/{id}/download", fh.Download)         // Скачать файл (авторизованный)
		r.Get("/{id}/preview", fh.Preview)           // Предпросмотр файла (опционально)
		r.Patch("/{id}/visibility", fh.TogglePublic) // Изменить публичность файла
	})

	return r
}
