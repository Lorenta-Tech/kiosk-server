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
	userID := "8473e7f9-2c72-4baf-b861-cd8238b15af6"
	userEmail := "suhas@test.com"

	resp, err := fh.fileservice.InitUpload(r.Context(), userID, userEmail, req)
	if err != nil {
		utils.HandleError(w, fh.logger, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, utils.Envelope{
		"data": resp,
	})
}
// HandleConfirmFileUpload godoc
//
//	@Summary      Confirm uploaded files and set print options
//	@Description  Called after files are uploaded directly to S3.
//	              Verifies each file exists in S3 staging, saves print options,
//	              calculates the total price, and moves the session to "priced".
//	              The response contains the per-file price breakdown and the
//	              total amount to charge — use this to create the Razorpay order.
//	@Tags         uploads
//	@Accept       json
//	@Produce      json
//	@Param        body  body      models.ConfirmUploadRequest  true  "Print options per file"
//	@Success      200   {object}  models.ConfirmUploadResponse
//	@Failure      400   {object}  utils.Envelope
//	@Failure      404   {object}  utils.Envelope
//	@Failure      500   {object}  utils.Envelope
//	@Router       /files/upload/confirm [post]
func (fh *FileHandler) HandleConfirmFileUpload(w http.ResponseWriter, r *http.Request) {
	req, err := utils.DecodeJSON[models.ConfirmUploadRequest](r)
	if err != nil {
		utils.HandleError(w, fh.logger, err)
		return
	}
 
	if err := validator.Validate(req); err != nil {
		utils.HandleError(w, fh.logger, apperror.BadRequest("validation_error", err.Error()))
		return
	}
 
	resp, err := fh.fileservice.ConfirmUpload(r.Context(), req)
	if err != nil {
		utils.HandleError(w, fh.logger, err)
		return
	}
 
	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"data": resp})
}
// GetPrintJobByToken godoc
//
//	@Summary      Get print job files by token
//	@Description  Fetches all files associated with a completed upload session
//	              identified by a 6-digit token. The session must be in "completed"
//	              status and must not have expired.
//	@Tags         uploads
//	@Accept       json
//	@Produce      json
//	@Param        body  body      models.Token  true  "6-digit session token"
//	@Success      200   {object}  models.GetPrintJobByTokenResponse
//	@Failure      400   {object}  utils.Envelope
//	@Failure      404   {object}  utils.Envelope
//	@Failure      500   {object}  utils.Envelope
//	@Router       /files/print-job [post]
func (fh *FileHandler) GetPrintJobByToken(w http.ResponseWriter, r *http.Request) {
	req, err := utils.DecodeJSON[models.Token](r)
	if err != nil {
		utils.HandleError(w, fh.logger, err)
		return
	}

	if err := validator.Validate(req); err != nil {
		utils.HandleError(w, fh.logger, apperror.BadRequest("validation_error", err.Error()))
		return
	}

	resp, err := fh.fileservice.GetPrintJobByToken(r.Context(), req.Token)
	if err != nil {
		utils.HandleError(w, fh.logger, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"data": resp})
}