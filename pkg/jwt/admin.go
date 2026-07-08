package jwt

import (
	"fmt"
	"time"

	"github.com/Lorenta-Tech/kiosk-server/pkg/apperror"
	"github.com/golang-jwt/jwt/v5"
)

type DeptAdminClaims struct {
	AdminID  string `json:"admin_id"`
	Email    string `json:"email"`
	BranchID string `json:"branch_id"`
	Role     string `json:"role"`

	jwt.RegisteredClaims
}

func GenerateDeptAdminToken(secret, adminID, email, branchID string) (string, error) {
	claims := DeptAdminClaims{
		AdminID:  adminID,
		Email:    email,
		BranchID: branchID,
		Role:     string(RoleDeptAdmin),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("jwt.GenerateDeptAdminToken: %w", err)
	}
	return signed, nil
}

func ParseDeptAdminToken(secret, tokenStr string) (*DeptAdminClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &DeptAdminClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, apperror.Unauthorized("invalid or expired token")
	}

	claims, ok := token.Claims.(*DeptAdminClaims)
	if !ok || !token.Valid {
		return nil, apperror.Unauthorized("invalid token claims")
	}
	return claims, nil
}