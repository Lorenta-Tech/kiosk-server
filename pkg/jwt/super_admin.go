package jwt

import (
	"fmt"
	"time"

	"github.com/Lorenta-Tech/kiosk-server/pkg/apperror"
	"github.com/golang-jwt/jwt/v5"
)

type SuperAdminClaims struct {
	Email string `json:"email"`
	Role  string `json:"role"`

	jwt.RegisteredClaims
}

func GenerateSuperAdminToken(
	secret string,
	email string,
) (string, error) {

	claims := SuperAdminClaims{
		Email: email,
		Role:  "super_admin",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf(
			"jwt.GenerateSuperAdminToken: %w",
			err,
		)
	}

	return signed, nil
}

func ParseSuperAdminToken(
	secret string,
	tokenStr string,
) (*SuperAdminClaims, error) {

	token, err := jwt.ParseWithClaims(
		tokenStr,
		&SuperAdminClaims{},
		func(t *jwt.Token) (any, error) {

			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf(
					"unexpected signing method: %v",
					t.Header["alg"],
				)
			}

			return []byte(secret), nil
		},
	)

	if err != nil {
		return nil, apperror.Unauthorized("invalid or expired token")
	}

	claims, ok := token.Claims.(*SuperAdminClaims)
	if !ok || !token.Valid {
		return nil, apperror.Unauthorized("invalid token claims")
	}

	return claims, nil
}
