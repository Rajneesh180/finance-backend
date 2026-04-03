package middleware

import (
	"net/http"

	"github.com/Rajneesh180/finance-backend/internal/api"
	"github.com/Rajneesh180/finance-backend/internal/domain"
)

func RequireRole(allowed ...domain.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetClaims(r.Context())
			if claims == nil {
				api.Unauthorized(w, "authentication required")
				return
			}

			for _, role := range allowed {
				if claims.Role == role {
					next.ServeHTTP(w, r)
					return
				}
			}

			api.Forbidden(w, "insufficient permissions")
		})
	}
}
