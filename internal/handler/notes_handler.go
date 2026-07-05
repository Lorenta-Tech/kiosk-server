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

// ================================================================
// Struct + constructor
// ================================================================

type NotesHandler struct {
	notesservice *service.NotesService
	logger       *slog.Logger
}

func NewNotesHandler(notesservice *service.NotesService, logger *slog.Logger) *NotesHandler {
	return &NotesHandler{notesservice: notesservice, logger: logger}
}

// ================================================================
// helper — pulls the authenticated admin's ID out of context.
// Set by middlewares.RequireRole after successful dept-admin /
// super-admin token verification. Returns a clean 401 instead of
// panicking if it's ever missing (defense in depth — RequireRole
// should always set this, but handlers shouldn't trust that blindly).
// ================================================================

func adminIDFromContext(w http.ResponseWriter, r *http.Request, logger *slog.Logger) (string, bool) {
	adminID, ok := r.Context().Value(middlewares.ContextAdminID).(string)
	if !ok || adminID == "" {
		utils.HandleError(w, logger, apperror.Unauthorized("missing admin identity"))
		return "", false
	}
	return adminID, true
}

// ================================================================
// ADMIN — HandleInitNoteUpload
// POST /admin/notes/upload/init
// ================================================================

func (nh *NotesHandler) HandleInitNoteUpload(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	req, err := utils.DecodeJSON[models.InitNoteUploadRequest](r)
	if err != nil {
		utils.HandleError(w, nh.logger, err)
		return
	}

	if err := validator.Validate(req); err != nil {
		utils.HandleError(w, nh.logger, apperror.BadRequest("validation_error", err.Error()))
		return
	}

	adminID, ok := adminIDFromContext(w, r, nh.logger)
	if !ok {
		return
	}

	resp, err := nh.notesservice.InitNoteUpload(ctx, adminID, req)
	if err != nil {
		utils.HandleError(w, nh.logger, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, utils.Envelope{"data": resp})
}

// ================================================================
// ADMIN — HandleConfirmNoteUpload
// POST /admin/notes/upload/confirm
// ================================================================

func (nh *NotesHandler) HandleConfirmNoteUpload(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	req, err := utils.DecodeJSON[models.ConfirmNoteUploadRequest](r)
	if err != nil {
		utils.HandleError(w, nh.logger, err)
		return
	}

	if err := validator.Validate(req); err != nil {
		utils.HandleError(w, nh.logger, apperror.BadRequest("validation_error", err.Error()))
		return
	}

	adminID, ok := adminIDFromContext(w, r, nh.logger)
	if !ok {
		return
	}

	if err := nh.notesservice.ConfirmNoteUpload(ctx, adminID, req); err != nil {
		utils.HandleError(w, nh.logger, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"message": "note upload confirmed"})
}

// ================================================================
// ADMIN — HandleUpdateNote
// PUT /admin/notes/:note_id
// ================================================================

func (nh *NotesHandler) HandleUpdateNote(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	noteID, err := utils.ReadParamID(r, "note_id")
	if err != nil {
		utils.HandleError(w, nh.logger, err)
		return
	}

	req, err := utils.DecodeJSON[models.UpdateNoteRequest](r)
	if err != nil {
		utils.HandleError(w, nh.logger, err)
		return
	}

	if err := validator.Validate(req); err != nil {
		utils.HandleError(w, nh.logger, apperror.BadRequest("validation_error", err.Error()))
		return
	}

	adminID, ok := adminIDFromContext(w, r, nh.logger)
	if !ok {
		return
	}

	if err := nh.notesservice.UpdateNote(ctx, adminID, noteID, req); err != nil {
		utils.HandleError(w, nh.logger, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"message": "note updated"})
}

// ================================================================
// ADMIN — HandleDeleteNote
// DELETE /admin/notes/:note_id
// ================================================================

func (nh *NotesHandler) HandleDeleteNote(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	noteID, err := utils.ReadParamID(r, "note_id")
	if err != nil {
		utils.HandleError(w, nh.logger, err)
		return
	}

	adminID, ok := adminIDFromContext(w, r, nh.logger)
	if !ok {
		return
	}

	if err := nh.notesservice.DeleteNote(ctx, adminID, noteID); err != nil {
		utils.HandleError(w, nh.logger, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"message": "note deleted"})
}

// ================================================================
// ADMIN — HandleCreateSubject
// POST /admin/subjects
// ================================================================

func (nh *NotesHandler) HandleCreateSubject(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	req, err := utils.DecodeJSON[models.CreateSubjectRequest](r)
	if err != nil {
		utils.HandleError(w, nh.logger, err)
		return
	}

	if err := validator.Validate(req); err != nil {
		utils.HandleError(w, nh.logger, apperror.BadRequest("validation_error", err.Error()))
		return
	}

	adminID, ok := adminIDFromContext(w, r, nh.logger)
	if !ok {
		return
	}

	resp, err := nh.notesservice.CreateSubject(ctx, adminID, req)
	if err != nil {
		utils.HandleError(w, nh.logger, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, utils.Envelope{"data": resp})
}

// ================================================================
// STUDENT — HandleListBranches
// GET /notes/branches
// ================================================================

func (nh *NotesHandler) HandleListBranches(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	resp, err := nh.notesservice.ListBranches(ctx)
	if err != nil {
		utils.HandleError(w, nh.logger, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"data": resp})
}

// ================================================================
// STUDENT — HandleListSemesters
// GET /notes/branches/:branch_id/semesters
// ================================================================

func (nh *NotesHandler) HandleListSemesters(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	branchID, err := utils.ReadParamID(r, "branch_id")
	if err != nil {
		utils.HandleError(w, nh.logger, err)
		return
	}

	resp, err := nh.notesservice.ListSemesters(ctx, branchID)
	if err != nil {
		utils.HandleError(w, nh.logger, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"data": resp})
}

// ================================================================
// STUDENT — HandleListSubjects
// GET /notes/semesters/:semester_id/subjects
// ================================================================

func (nh *NotesHandler) HandleListSubjects(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	semesterID, err := utils.ReadParamID(r, "semester_id")
	if err != nil {
		utils.HandleError(w, nh.logger, err)
		return
	}

	resp, err := nh.notesservice.ListSubjects(ctx, semesterID)
	if err != nil {
		utils.HandleError(w, nh.logger, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"data": resp})
}

// ================================================================
// STUDENT — HandleListModules
// GET /notes/subjects/:subject_id/modules
// ================================================================

func (nh *NotesHandler) HandleListModules(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	subjectID, err := utils.ReadParamID(r, "subject_id")
	if err != nil {
		utils.HandleError(w, nh.logger, err)
		return
	}

	resp, err := nh.notesservice.ListModules(ctx, subjectID)
	if err != nil {
		utils.HandleError(w, nh.logger, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"data": resp})
}

// ================================================================
// STUDENT — HandleListNotes
// GET /notes/modules/:module_id/notes
// ================================================================

func (nh *NotesHandler) HandleListNotes(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	moduleID, err := utils.ReadParamID(r, "module_id")
	if err != nil {
		utils.HandleError(w, nh.logger, err)
		return
	}

	resp, err := nh.notesservice.ListNotes(ctx, moduleID)
	if err != nil {
		utils.HandleError(w, nh.logger, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"data": resp})
}

// ================================================================
// STUDENT — HandleGetNoteViewURL
// GET /notes/:note_id/view
// Returns presigned S3 URL — frontend opens it directly in browser
// ================================================================

func (nh *NotesHandler) HandleGetNoteViewURL(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	noteID, err := utils.ReadParamID(r, "note_id")
	if err != nil {
		utils.HandleError(w, nh.logger, err)
		return
	}

	resp, err := nh.notesservice.GetNoteViewURL(ctx, noteID)
	if err != nil {
		utils.HandleError(w, nh.logger, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"data": resp})
}

// ================================================================
// STUDENT — HandleNotesPrintInit
// POST /notes/print/init
// Creates print session reusing note S3 files — feeds into existing payment flow
// ================================================================

func (nh *NotesHandler) HandleNotesPrintInit(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	req, err := utils.DecodeJSON[models.NotesPrintInitRequest](r)
	if err != nil {
		utils.HandleError(w, nh.logger, err)
		return
	}

	if err := validator.Validate(req); err != nil {
		utils.HandleError(w, nh.logger, apperror.BadRequest("validation_error", err.Error()))
		return
	}

	userID := r.Context().Value(middlewares.ContextUserID).(string)
	userEmail := r.Context().Value(middlewares.ContextUserEmail).(string)

	resp, err := nh.notesservice.NotesPrintInit(ctx, userID, userEmail, req)
	if err != nil {
		utils.HandleError(w, nh.logger, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"data": resp})
}