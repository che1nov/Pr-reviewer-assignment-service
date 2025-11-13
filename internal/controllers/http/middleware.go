package httpcontroller

import (
	"log/slog"
	"net/http"
)

func adminAuth(logger *slog.Logger, token string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !checkBearer(r, token) {
				logger.WarnContext(r.Context(), "неверный админ токен")
				respondError(logger, w, http.StatusUnauthorized, "UNAUTHORIZED", "admin token required")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func userAuth(logger *slog.Logger, adminToken, userToken string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if checkBearer(r, adminToken) || checkBearer(r, userToken) {
				next.ServeHTTP(w, r)
				return
			}
			logger.WarnContext(r.Context(), "неверный пользовательский токен")
			respondError(logger, w, http.StatusUnauthorized, "UNAUTHORIZED", "token required")
		})
	}
}

func checkBearer(r *http.Request, token string) bool {
	if token == "" {
		return false
	}
	value := r.Header.Get("Authorization")
	const prefix = "Bearer "
	return len(value) == len(prefix)+len(token) && value[:len(prefix)] == prefix && value[len(prefix):] == token
}
