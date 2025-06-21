package service

import "errors"

var (
	ErrForbidden    = errors.New("access forbidden")
	ErrNotFound     = errors.New("file not found")
	ErrInvalidInput = errors.New("invalid input")
	ErrUploadFailed = errors.New("file upload failed")
	ErrDeleteFailed = errors.New("file delete failed")
)
