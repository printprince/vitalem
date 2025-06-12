package service

import "errors"

// ErrForbidden — используется, когда пользователь пытается получить доступ к чужим данным.
var ErrForbidden = errors.New("access forbidden")
