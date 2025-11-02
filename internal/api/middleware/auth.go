package middleware

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gabrielmelo/tg-forward/internal/api/common"
)

func Auth(token string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(common.ApiErrorResponse{
					Code:    "UNAUTHORIZED",
					Message: "Missing authorization header",
				})
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(common.ApiErrorResponse{
					Code:    "UNAUTHORIZED",
					Message: "Invalid authorization header format",
				})
				return
			}

			if parts[1] != token {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(common.ApiErrorResponse{
					Code:    "UNAUTHORIZED",
					Message: "Invalid token",
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
