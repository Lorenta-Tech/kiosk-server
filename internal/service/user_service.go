package service

import (
	"context"
	"log/slog"

	"github.com/Lorenta-Tech/kiosk-server/internal/models"
	"github.com/Lorenta-Tech/kiosk-server/internal/repository"
	"github.com/Lorenta-Tech/kiosk-server/pkg/apperror"
	appjwt "github.com/Lorenta-Tech/kiosk-server/pkg/jwt"
	"github.com/google/uuid"
	"google.golang.org/api/idtoken"
)

type UserService struct {
	userrepo        repository.UserRepo
	logger          *slog.Logger
	jwtSecret       string
	googleClientID  string
}

func NewUserService(
	userrepo repository.UserRepo,
	logger *slog.Logger,
	jwtSecret string,
	googleClientID string,
) *UserService {
	return &UserService{
		userrepo:       userrepo,
		logger:         logger,
		jwtSecret:      jwtSecret,
		googleClientID: googleClientID,
	}
}

func (s *UserService) GoogleAuth(
	ctx context.Context,
	req models.GoogleAuthRequest,
) (models.AuthResponse, error) {

	s.logger.Info("google auth started")

	payload, err := idtoken.Validate(ctx, req.IDToken, s.googleClientID)
	if err != nil {
		s.logger.Warn("google id token validation failed", "error", err)
		return models.AuthResponse{}, apperror.Unauthorized(
			"google token is invalid or expired",
		)
	}

	claims, err := extractGoogleClaims(payload)
	if err != nil {
		s.logger.Warn("failed to extract google claims", "error", err)
		return models.AuthResponse{}, err
	}

	s.logger.Info("google token verified",
		"google_id", claims.Sub,
		"email", claims.Email,
	)

	user, err := s.userrepo.GetByGoogleID(ctx, claims.Sub)
	if err != nil {
		var appErr *apperror.AppError
		if apperror.As(err, &appErr) && appErr.Code == "user_not_found" {
			s.logger.Info("new user — creating account",
				"google_id", claims.Sub,
				"email", claims.Email,
			)

			user, err = s.userrepo.Create(ctx, models.User{
				ID:       uuid.NewString(),
				GoogleID: claims.Sub,
				Email:    claims.Email,
				Name:     claims.Name,
			})
			if err != nil {
				return models.AuthResponse{}, err
			}

			s.logger.Info("new user created",
				"user_id", user.ID,
				"email", user.Email,
			)
		} else {
			// Real DB error — not a missing row
			return models.AuthResponse{}, err
		}
	} else {
		s.logger.Info("returning user authenticated",
			"user_id", user.ID,
			"email", user.Email,
		)
	}

	token, err := appjwt.Generate(s.jwtSecret, user.ID, user.Email, user.Name)
	if err != nil {
		return models.AuthResponse{}, apperror.Internal("failed to generate auth token", err)
	}

	s.logger.Info("google auth completed",
		"user_id", user.ID,
		"email", user.Email,
	)

	return models.AuthResponse{
		Token: token,
		User: models.UserInfo{
			ID:    user.ID,
			Email: user.Email,
			Name:  user.Name,
		},
	}, nil
}

//helpers
func extractGoogleClaims(payload *idtoken.Payload) (models.GoogleClaims, error) {
	sub := payload.Subject
	if sub == "" {
		return models.GoogleClaims{}, apperror.Unauthorized("google token missing subject claim")
	}

	email, _ := payload.Claims["email"].(string)
	if email == "" {
		return models.GoogleClaims{}, apperror.Unauthorized("google token missing email claim")
	}

	name, _ := payload.Claims["name"].(string)
	if name == "" {
		name = email 
	}

	return models.GoogleClaims{
		Sub:   sub,
		Email: email,
		Name:  name,
	}, nil
}