package middleware

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// CustomValidator содержит валидатор запросов
type CustomValidator struct {
	validator *validator.Validate
}

// Validate проверяет данные на соответствие правилам валидации
func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}

// NewValidator создает новый экземпляр валидатора
func NewValidator() *CustomValidator {
	return &CustomValidator{
		validator: validator.New(),
	}
}
