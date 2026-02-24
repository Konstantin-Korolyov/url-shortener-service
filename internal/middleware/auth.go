package middleware

import (
	"context"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

type contextKey string

const UserIDKey contextKey = "user_id"

// AuthMiddleware проверяет API-ключ и добавляет user_id в контекст
func AuthMiddleware(db *pgxpool.Pool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey := r.Header.Get("X-API-Key")
			if apiKey == "" {
				http.Error(w, "API key required", http.StatusUnauthorized)
				return
			}

			var userID int64
			err := db.QueryRow(r.Context(), "SELECT id FROM users WHERE api_key = $1", apiKey).Scan(&userID)
			if err != nil {
				// Можно проверить, что ошибка именно "no rows"
				http.Error(w, "Invalid API key", http.StatusUnauthorized)
				return
			}

			// Добавляем user_id в контекст
			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
