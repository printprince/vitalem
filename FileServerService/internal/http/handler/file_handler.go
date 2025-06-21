package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"fileserver/internal/http/middleware"
	"fileserver/internal/model"
	"fileserver/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type FileHandler struct {
	Service service.FileService
}

func NewFileHandler(s service.FileService) *FileHandler {
	return &FileHandler{Service: s}
}

func (h *FileHandler) Upload(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userIDStr == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	// Ограничение размера загружаемых данных (например, 50 МБ)
	err = r.ParseMultipartForm(50 << 20)
	if err != nil {
		http.Error(w, "cannot parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	files := r.MultipartForm.File["file"]
	if len(files) == 0 {
		http.Error(w, "no files uploaded", http.StatusBadRequest)
		return
	}

	var uploadedFiles []*model.File

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			http.Error(w, "cannot open file: "+err.Error(), http.StatusBadRequest)
			return
		}

		fileBytes, err := io.ReadAll(file)
		file.Close()
		if err != nil {
			http.Error(w, "cannot read file: "+err.Error(), http.StatusInternalServerError)
			return
		}

		contentType := fileHeader.Header.Get("Content-Type")

		f := &model.File{
			Name:         fileHeader.Filename,
			OriginalName: fileHeader.Filename,
			Size:         fileHeader.Size,
			UserID:       userID,
			Bucket:       "files", // default bucket
		}
		f.SetContentType(contentType)

		err = h.Service.Upload(r.Context(), f, bytes.NewReader(fileBytes), fileHeader.Size)
		if err != nil {
			http.Error(w, "saving file failed: "+err.Error(), http.StatusInternalServerError)
			return
		}

		uploadedFiles = append(uploadedFiles, f)
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(uploadedFiles) // Возвращаем массив
}

// Get возвращает мета информацию о файле
func (h *FileHandler) Get(w http.ResponseWriter, r *http.Request) {
	fileID := chi.URLParam(r, "id")
	if fileID == "" {
		http.Error(w, "file id required", http.StatusBadRequest)
		return
	}

	id, err := uuid.Parse(fileID)
	if err != nil {
		http.Error(w, "invalid file id", http.StatusBadRequest)
		return
	}

	file, err := h.Service.Get(r.Context(), id.String())
	if err != nil {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(file)
}

// Delete удаляет файл по ID
func (h *FileHandler) Delete(w http.ResponseWriter, r *http.Request) {
	fileID := chi.URLParam(r, "id")
	if fileID == "" {
		http.Error(w, "file id required", http.StatusBadRequest)
		return
	}

	id, err := uuid.Parse(fileID)
	if err != nil {
		http.Error(w, "invalid file id", http.StatusBadRequest)
		return
	}

	err = h.Service.Delete(r.Context(), id.String())
	if err != nil {
		http.Error(w, "failed to delete file", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// List возвращает список файлов пользователя
func (h *FileHandler) List(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userIDStr == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	files, err := h.Service.ListByUser(r.Context(), userID.String())
	if err != nil {
		http.Error(w, "cannot get files", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}

// PublicDownload позволяет скачать файл без авторизации, если он публичный.
func (h *FileHandler) PublicDownload(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	file, err := h.Service.Get(r.Context(), id)
	if err != nil || file == nil || !file.IsPublic {
		http.Error(w, "file not found or not public", http.StatusNotFound)
		return
	}

	reader, filename, contentType, err := h.Service.DownloadFile(r.Context(), id)
	if err != nil {
		http.Error(w, "failed to download file", http.StatusInternalServerError)
		return
	}
	defer reader.Close()

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	_, err = io.Copy(w, reader)
	if err != nil {
		http.Error(w, "failed to stream file", http.StatusInternalServerError)
		return
	}
}

func (h *FileHandler) Download(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Получаем файл по UUID
	file, err := h.Service.Get(r.Context(), id)
	if err != nil || file == nil {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}

	// Используем UUID для DownloadFile (чтобы найти файл в БД и скачать из MinIO)
	reader, filename, contentType, err := h.Service.DownloadFile(r.Context(), id)
	if err != nil {
		http.Error(w, "failed to download file", http.StatusInternalServerError)
		return
	}
	defer reader.Close()

	w.Header().Set("Content-Type", contentType)
	encodedFilename := url.PathEscape(filename)
	w.Header().Set("Content-Disposition", "attachment; filename*=UTF-8''"+encodedFilename)
	_, err = io.Copy(w, reader)
	if err != nil {
		http.Error(w, "failed to stream file", http.StatusInternalServerError)
		return
	}
}

// Preview позволяет показать предварительный просмотр, например, изображений.
func (h *FileHandler) Preview(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	file, err := h.Service.Get(r.Context(), id)
	if err != nil || file == nil {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}

	// вместо передачи file.Path в DownloadFile (который ожидает UUID), сделайте отдельный вызов DownloadByPath
	reader, err := h.Service.DownloadByPath(r.Context(), file.Path)
	if err != nil {
		http.Error(w, "failed to preview file", http.StatusInternalServerError)
		return
	}
	defer reader.Close()

	w.Header().Set("Content-Type", file.ContentType())
	w.Header().Set("Content-Disposition", "inline; filename=\""+file.Name+"\"")
	_, err = io.Copy(w, reader)
	if err != nil {
		http.Error(w, "failed to stream file", http.StatusInternalServerError)
		return
	}
}

// TogglePublic переключает флаг публичности файла.
func (h *FileHandler) TogglePublic(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	file, err := h.Service.Get(r.Context(), id)
	if err != nil || file == nil {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}

	file.IsPublic = !file.IsPublic
	err = h.Service.Update(r.Context(), file)
	if err != nil {
		http.Error(w, "failed to update visibility", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"id":        file.ID,
		"is_public": file.IsPublic,
	})
}
