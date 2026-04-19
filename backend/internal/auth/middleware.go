package auth

import (
	"context"
	"net/http"
	"strings"
)

type userIDKey struct{}

func UserIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(userIDKey{}).(string)
	return id, ok
}

func Middleware(jwtSecret string) func(http.Handler) http.Handler {
	secret := []byte(jwtSecret)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				writeError(w, http.StatusUnauthorized, "missing token")
				return
			}
			claims, err := ValidateAccessToken(secret, strings.TrimPrefix(header, "Bearer "))
			if err != nil {
				writeError(w, http.StatusUnauthorized, "invalid token")
				return
			}
			next.ServeHTTP(w, r.WithContext(
				context.WithValue(r.Context(), userIDKey{}, claims.UserID),
			))
		})
	}
}
