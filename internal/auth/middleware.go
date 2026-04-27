package auth

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
)

type userIDKey struct{}

func UserIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(userIDKey{}).(string)
	return id, ok
}

func Middleware(jwtSecret string, log *slog.Logger) func(http.Handler) http.Handler {
	secret := []byte(jwtSecret)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				log.WarnContext(r.Context(), "missing token", slog.String("error", "auth: bearer token absent"))
				http.Error(w, "missing token", http.StatusUnauthorized)
				return
			}
			claims, err := ValidateAccessToken(secret, strings.TrimPrefix(header, "Bearer "))
			if err != nil {
				log.WarnContext(r.Context(), "invalid token", slog.String("error", err.Error()))
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r.WithContext(
				context.WithValue(r.Context(), userIDKey{}, claims.UserID),
			))
		})
	}
}
