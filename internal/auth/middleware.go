package auth

import (
	"context"
	"log"
	"net/http"
)

type AuthMiddleware struct {
	authService func(token string) (User, error)
	logger      func(v ...interface{})
}

func NewAuthMiddleware(authService func(token string) (User, error), logger func(v ...interface{})) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
		logger:      logger,
	}
}

func (am *AuthMiddleware) MiddlewareFunc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := extractTokenFromHeader(r)
		if tokenString == "" {
			log.Println("Unauthorized - Token not provided")
			http.Error(w, "Unauthorized - Token not provided", http.StatusUnauthorized)
			return
		}

		user, err := am.authService(tokenString)
		if err != nil {
			log.Println("Unauthorized - Invalid token:", err)
			http.Error(w, "Unauthorized - Invalid token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "user", user)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func extractTokenFromHeader(r *http.Request) string {
	authHeader := r.Header.Get("token")
	if authHeader == "" {
		return ""
	}

	return authHeader
}
