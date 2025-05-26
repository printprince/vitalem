package models

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
)

type TokenClaims struct {
	jwt.StandardClaims
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	Role   string    `json:"role"`
}
