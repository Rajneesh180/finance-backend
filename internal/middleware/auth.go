package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/Rajneesh180/finance-backend/internal/api"
	"github.com/Rajneesh180/finance-backend/internal/service"
)

type contextKey string

const claimsKey contextKey = "claims"

func Auth(auth *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				api.Unauthorized(w, "missing authorization header")
				return
			}

			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				api.Unauthorized(w, "invalid authorization format")
				return
			}

			claims, err := auth.ValidateToken(parts[1])
			if err != nil {
				api.Unauthorized(w, "invalid or expired token")
				return
			}

			ctx := context.WithValue(r.Context(), claimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetClaims(ctx context.Context) *service.Claims {
	claims, _ := ctx.Value(claimsKey).(*service.Claims)
	return claims
}
