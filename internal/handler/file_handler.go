package handler

import (
	"log/slog"
	"net/http"

	"github.com/Lorenta-Tech/kiosk-server/internal/models"
	"github.com/Lorenta-Tech/kiosk-server/internal/service"
	"github.com/Lorenta-Tech/kiosk-server/internal/validator"
	"github.com/Lorenta-Tech/kiosk-server/pkg/apperror"
	"github.com/Lorenta-Tech/kiosk-server/pkg/utils"
)

type FileHandler struct {
	fileservice *service.FileService
	logger      *slog.Logger
}

func NewFileHandler(fileservice *service.FileService, logger *slog.Logger) *FileHandler {
	return &FileHandler{fileservice: fileservice, logger: logger}
}

// HandleInitFileUpload godoc
//
//	@Summary      Initialise a file upload session
//	@Description  Creates an upload session and returns presigned S3 PUT URLs
//	              for each file. The client uploads files directly to S3 using
//	              these URLs, then calls POST /files/upload/confirm.
//	@Tags         uploads
//	@Accept       json
//	@Produce      json
//	@Param        body  body      models.InitUploadRequest   true  "Files to upload"
//	@Success      201   {object}  models.InitUploadResponse
//	@Failure      400   {object}  utils.Envelope
//	@Failure      500   {object}  utils.Envelope
//	@Router       /files/upload/init [post]
func (fh *FileHandler) HandleInitFileUpload(w http.ResponseWriter, r *http.Request) {

	req, err := utils.DecodeJSON[models.InitUploadRequest](r)
	if err != nil {
		utils.HandleError(w, fh.logger, err)
		return
	}

	if err := validator.Validate(req); err != nil {
		utils.HandleError(w, fh.logger, apperror.BadRequest("validation_error", err.Error()))
		return
	}

	// TODO: replace with real values once auth middleware is wired.
	userID := "hardcoded-user-id"
	userEmail := "hardcoded@email.com"

	resp, err := fh.fileservice.InitUpload(r.Context(), userID, userEmail, req)
	if err != nil {
		utils.HandleError(w, fh.logger, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, utils.Envelope{
		"data": resp,
	})
}