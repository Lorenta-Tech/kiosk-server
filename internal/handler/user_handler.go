package handler

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/Lorenta-Tech/kiosk-server/internal/middlewares"
	"github.com/Lorenta-Tech/kiosk-server/internal/models"
	"github.com/Lorenta-Tech/kiosk-server/internal/service"
	"github.com/Lorenta-Tech/kiosk-server/internal/validator"
	"github.com/Lorenta-Tech/kiosk-server/pkg/apperror"
	"github.com/Lorenta-Tech/kiosk-server/pkg/utils"
)

type UserHandler struct {
	userservice *service.UserService
	logger      *slog.Logger
}

func NewUserHandler(userservice *service.UserService, logger *slog.Logger) *UserHandler {
	return &UserHandler{userservice: userservice, logger: logger}
}

// HandleGoogleAuth godoc
//
//	@Summary      Authenticate with Google OAuth
//	@Description  Accepts a Google ID token from the frontend (obtained after
//	              the user completes Google Sign-In). The server verifies the
//	              token with Google, upserts the user, and returns a signed JWT
//	              for all subsequent API calls.
//	@Tags         auth
//	@Accept       json
//	@Produce      json
//	@Param        body  body      models.GoogleAuthRequest  true  "Google ID token"
//	@Success      200   {object}  models.AuthResponse
//	@Failure      400   {object}  utils.Envelope
//	@Failure      401   {object}  utils.Envelope
//	@Failure      500   {object}  utils.Envelope
//	@Router       /auth/google [post]
func (uh *UserHandler) HandleGoogleAuth(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	req, err := utils.DecodeJSON[models.GoogleAuthRequest](r)
	if err != nil {
		utils.HandleError(w, uh.logger, err)
		return
	}

	if err := validator.Validate(req); err != nil {
		utils.HandleError(w, uh.logger, apperror.BadRequest("validation_error", err.Error()))
		return
	}

	resp, err := uh.userservice.GoogleAuth(ctx, req)
	if err != nil {
		utils.HandleError(w, uh.logger, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"data": resp})
}

// HandleMe godoc
//
//	@Summary      Get the authenticated user
//	@Description  Returns the user information encoded in a valid application JWT.
//	@Tags         auth
//	@Produce      json
//	@Security     BearerAuth
//	@Success      200  {object}  utils.Envelope
//	@Failure      401  {object}  utils.Envelope
//	@Router       /auth/me [get]
func (uh *UserHandler) HandleMe(w http.ResponseWriter, r *http.Request) {
	userID := middlewares.GetUserID(r.Context())
	userEmail := middlewares.GetUserEmail(r.Context())
	userName := middlewares.GetUserName(r.Context())

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"data": models.UserInfo{
		ID:    userID,
		Email: userEmail,
		Name:  userName,
	}})
}
