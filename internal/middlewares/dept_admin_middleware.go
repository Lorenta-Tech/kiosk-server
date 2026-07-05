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

type ContextKey string

const (
	ContextAdminID    ContextKey = "admin_id"
	ContextBranchID   ContextKey = "branch_id"
	ContextAdminEmail ContextKey = "admin_email"
	ContextAdminRole  ContextKey = "admin_role"
)

// Secrets holds every JWT secret RequireRole might need. Built once in
// app.go and passed to every RequireRole call, so no route ever hardcodes
// or mismatches which secret belongs to which role again.
type Secrets struct {
	DeptAdmin  string
	SuperAdmin string
}

// RequireRole is the single, generic role-gate for all admin routes.
// It is given the roles a route accepts (e.g. RoleSuperAdmin, or both
// RoleDeptAdmin and RoleSuperAdmin), and for each one attempts to verify
// the bearer token against that role's own secret with that role's own
// claims type. The first role whose secret+parser accepts the token wins.
//
// This keeps authentication (is the signature valid for this role's
// secret) and authorization (is this role allowed on this route) as two
// clearly separated checks inside one pass, without needing a
// role-specific middleware per role.
func RequireRole(secrets Secrets, logger *slog.Logger, allowedRoles ...appjwt.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				utils.HandleError(w, logger, apperror.Unauthorized("missing authorization header"))
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				utils.HandleError(w, logger, apperror.Unauthorized("invalid authorization header"))
				return
			}
			tokenString := parts[1]

			var (
				adminID, branchID, email, role string
				verified                        bool
			)

			for _, allowedRole := range allowedRoles {
				switch allowedRole {

				case appjwt.RoleDeptAdmin:
					claims, err := appjwt.ParseDeptAdminToken(secrets.DeptAdmin, tokenString)
					if err == nil && claims.Role == string(appjwt.RoleDeptAdmin) {
						adminID, branchID, email, role = claims.AdminID, claims.BranchID, claims.Email, claims.Role
						verified = true
					}

				case appjwt.RoleSuperAdmin:
					claims, err := appjwt.ParseSuperAdminToken(secrets.SuperAdmin, tokenString)
					if err == nil && claims.Role == string(appjwt.RoleSuperAdmin) {
						email, role = claims.Email, claims.Role
						verified = true
					}
				}

				if verified {
					break
				}
			}

			if !verified {
				utils.HandleError(w, logger, apperror.Forbidden("insufficient permissions"))
				return
			}

			ctx := context.WithValue(r.Context(), ContextAdminID, adminID)
			ctx = context.WithValue(ctx, ContextBranchID, branchID)
			ctx = context.WithValue(ctx, ContextAdminEmail, email)
			ctx = context.WithValue(ctx, ContextAdminRole, role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}