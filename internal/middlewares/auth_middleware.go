package middlewares

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/Lorenta-Tech/kiosk-server/pkg/apperror"
	appjwt "github.com/Lorenta-Tech/kiosk-server/pkg/jwt"
	"github.com/Lorenta-Tech/kiosk-server/pkg/utils"
)

type contextKey string

const (
	ContextUserID    contextKey = "user_id"
	ContextUserEmail contextKey = "user_email"
	ContextUserName  contextKey = "user_name"
)

func AuthMiddleware(jwtSecret string,logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				utils.HandleError(w, logger, apperror.Unauthorized("authorization header is missing"))
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				utils.HandleError(w, logger, apperror.Unauthorized("authorization header format must be: Bearer <token>"))
				return
			}

			tokenStr := parts[1]

			claims, err := appjwt.Parse(jwtSecret, tokenStr)
			if err != nil {
				utils.HandleError(w, logger, err)
				return
			}

			ctx := context.WithValue(r.Context(), ContextUserID, claims.UserID)
			ctx = context.WithValue(ctx, ContextUserEmail, claims.Email)
			ctx = context.WithValue(ctx, ContextUserName, claims.Name)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserID(ctx context.Context) string {
	v, _ := ctx.Value(ContextUserID).(string)
	return v
}

func GetUserEmail(ctx context.Context) string {
	v, _ := ctx.Value(ContextUserEmail).(string)
	return v
}

func GetUserName(ctx context.Context) string {
	v, _ := ctx.Value(ContextUserName).(string)
	return v
}