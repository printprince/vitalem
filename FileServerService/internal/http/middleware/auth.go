package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type contextKey string

const UserIDKey contextKey = "userID"

func AuthMiddleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
				return []byte(secret), nil
			})
			if err != nil || !token.Valid {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, "invalid token claims", http.StatusUnauthorized)
				return
			}

			userIDStr, ok := claims["user_id"].(string)
			if !ok {
				http.Error(w, "invalid user_id type", http.StatusUnauthorized)
				return
			}

			// Валидация формата UUID
			userID, err := uuid.Parse(userIDStr)
			if err != nil {
				http.Error(w, "invalid user_id format", http.StatusUnauthorized)
				return
			}

			// Передаём UUID в контекст как строку или как uuid.UUID (по твоему усмотрению)
			ctx := context.WithValue(r.Context(), UserIDKey, userID.String())
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
