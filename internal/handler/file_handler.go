package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

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
//	              for each file. Also returns a 6-digit token the user can
//	              enter at a kiosk to retrieve their job.
//	@Tags         uploads
//	@Accept       json
//	@Produce      json
//	@Param        body  body      models.InitUploadRequest  true  "Files to upload"
//	@Success      201   {object}  models.InitUploadResponse
//	@Failure      400   {object}  utils.Envelope
//	@Failure      500   {object}  utils.Envelope
//	@Router       /files/upload/init [post]
func (fh *FileHandler) HandleInitFileUpload(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	req, err := utils.DecodeJSON[models.InitUploadRequest](r)
	if err != nil {
		utils.HandleError(w, fh.logger, err)
		return
	}

	if err := validator.Validate(req); err != nil {
		utils.HandleError(w, fh.logger, apperror.BadRequest("validation_error", err.Error()))
		return
	}

	// TODO: replace with auth middleware values
	// userID    := r.Context().Value(middlewares.UserIDKey).(string)
	// userEmail := r.Context().Value(middlewares.UserEmailKey).(string)
	userID := "8473e7f9-2c72-4baf-b861-cd8238b15af6"
	userEmail := "hardcoded@email.com"

	resp, err := fh.fileservice.InitUpload(ctx, userID, userEmail, req)
	if err != nil {
		utils.HandleError(w, fh.logger, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, utils.Envelope{"data": resp})
}

// HandleConfirmFileUpload godoc
//
//	@Summary      Confirm uploaded files and set print options
//	@Description  Called after files are uploaded directly to S3.
//	              Verifies each file exists in S3 staging, saves print options,
//	              calculates the total price, and moves the session to "priced".
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
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	req, err := utils.DecodeJSON[models.ConfirmUploadRequest](r)
	if err != nil {
		utils.HandleError(w, fh.logger, err)
		return
	}

	if err := validator.Validate(req); err != nil {
		utils.HandleError(w, fh.logger, apperror.BadRequest("validation_error", err.Error()))
		return
	}

	resp, err := fh.fileservice.ConfirmUpload(ctx, req)
	if err != nil {
		utils.HandleError(w, fh.logger, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"data": resp})
}

// HandleGetRecentPrintJobs godoc
//
//	@Summary      Get recent print jobs for the current user
//	@Description  Returns the last 10 print jobs for the authenticated user.
//	              Jobs with status "created" are excluded.
//	@Tags         jobs
//	@Produce      json
//	@Success      200  {object}  models.RecentPrintJobsResponse
//	@Failure      500  {object}  utils.Envelope
//	@Router       /files/jobs/recent [get]
func (fh *FileHandler) HandleGetRecentPrintJobs(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// TODO: replace with auth middleware value
	// userID := r.Context().Value(middlewares.UserIDKey).(string)
	userID := "8473e7f9-2c72-4baf-b861-cd8238b15af6"

	resp, err := fh.fileservice.GetRecentPrintJobs(ctx, userID)
	if err != nil {
		utils.HandleError(w, fh.logger, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"data": resp})
}

func (fh *FileHandler) HandleActivePrintJobs(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// TODO: replace with auth middleware value
	// userID := r.Context().Value(middlewares.UserIDKey).(string)
	userID := "8473e7f9-2c72-4baf-b861-cd8238b15af6"

	resp, err := fh.fileservice.GetActivePrintJobs(ctx, userID)
	if err != nil {
		utils.HandleError(w, fh.logger, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"data": resp})
}

func (fh *FileHandler) HandleErrorRequestFromPrinter(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	req, err := utils.DecodeJSON[models.ErrorRequestFromPrinter](r)
	if err != nil {
		utils.HandleError(w, fh.logger, err)
		return
	}

	if err := validator.Validate(req); err != nil {
		utils.HandleError(w, fh.logger, apperror.BadRequest("validation_error", err.Error()))
		return
	}

	if err := fh.fileservice.ErrorRequestFromPrinter(ctx, req); err != nil {
		utils.HandleError(w, fh.logger, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"mail":"sent"})
}

// HandleGetJobByToken godoc
//
//	@Summary      Retrieve a print job by 6-digit token
//	@Description  The kiosk calls this with the token the user entered.
//	              Returns the full print job only if the session is in
//	              "priced" status and has not expired. Any other status
//	              returns a clear error message telling the user what to do.
//	@Tags         jobs
//	@Accept       json
//	@Produce      json
//	@Param        body  body      models.GetJobByTokenRequest  true  "6-digit token"
//	@Success      200   {object}  models.TokenJobResponse
//	@Failure      400   {object}  utils.Envelope
//	@Failure      404   {object}  utils.Envelope
//	@Failure      500   {object}  utils.Envelope
//	@Router       /files/jobs/token [post]
func (fh *FileHandler) HandleGetJobByToken(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	req, err := utils.DecodeJSON[models.GetJobByTokenRequest](r)
	if err != nil {
		utils.HandleError(w, fh.logger, err)
		return
	}

	if err := validator.Validate(req); err != nil {
		utils.HandleError(w, fh.logger, apperror.BadRequest("validation_error", err.Error()))
		return
	}

	resp, err := fh.fileservice.GetJobByToken(ctx, req)
	if err != nil {
		utils.HandleError(w, fh.logger, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)

	if err := encoder.Encode(utils.Envelope{"data": resp}); err != nil {
		fh.logger.Error("failed to encode response", "error", err)
	}
}

// HandleExpireSessionAfterPrinting godoc
//
//	@Summary      Expire session after printing
//	@Description  Marks the upload session as completed and expires it immediately.
//	@Tags         jobs
//	@Accept       json
//	@Produce      json
//	@Param        body  body      models.ExpireSessionRequest  true  "Session ID"
//	@Success      200   {object}  utils.Envelope
//	@Failure      400   {object}  utils.Envelope
//	@Failure      404   {object}  utils.Envelope
//	@Failure      500   {object}  utils.Envelope
//	@Router       /files/jobs/complete [post]
func (fh *FileHandler) HandleExpireSessionAfterPrinting(
	w http.ResponseWriter,
	r *http.Request,
) {

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	req, err := utils.DecodeJSON[models.ExpireSessionRequest](r)
	if err != nil {
		utils.HandleError(w, fh.logger, err)
		return
	}

	if err := validator.Validate(req); err != nil {
		utils.HandleError(
			w,
			fh.logger,
			apperror.BadRequest("validation_error", err.Error()),
		)
		return
	}

	if err := fh.fileservice.ExpireSessionAfterPrinting(ctx, req); err != nil {
		utils.HandleError(w, fh.logger, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{
		"message": "session marked as completed",
	})
}
