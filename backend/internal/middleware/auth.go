package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/sergiohpreis/bolaocopa/backend/internal/service"
	"github.com/sergiohpreis/bolaocopa/backend/pkg/apierror"
)

type contextKey string

const userIDKey contextKey = "userID"

func Auth(authSvc *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				apierror.Unauthorized(w, "missing or invalid authorization header")
				return
			}
			token := strings.TrimPrefix(header, "Bearer ")
			userID, err := authSvc.ValidateAccessToken(token)
			if err != nil {
				apierror.Unauthorized(w, "invalid or expired token")
				return
			}
			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(userIDKey).(string)
	return id
}
