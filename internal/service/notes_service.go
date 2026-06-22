package service

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Lorenta-Tech/kiosk-server/internal/models"
	"github.com/Lorenta-Tech/kiosk-server/internal/repository"
	"github.com/Lorenta-Tech/kiosk-server/pkg/apperror"
	s3pkg "github.com/Lorenta-Tech/kiosk-server/pkg/s3"
	"github.com/Lorenta-Tech/kiosk-server/pkg/utils"
	"github.com/google/uuid"
)

// ================================================================
// Struct + constructor
// ================================================================

type NotesService struct {
	notesrepo repository.NotesRepo
	filerepo  repository.FileRepo // for NotesPrintInit session creation
	db        *sql.DB             // for transactions in NotesPrintInit
	s3        *s3pkg.Client
	logger    *slog.Logger
}

func NewNotesService(
	notesrepo repository.NotesRepo,
	filerepo repository.FileRepo,
	db *sql.DB,
	s3 *s3pkg.Client,
	logger *slog.Logger,
) *NotesService {
	return &NotesService{
		notesrepo: notesrepo,
		filerepo:  filerepo,
		db:        db,
		s3:        s3,
		logger:    logger,
	}
}

// ================================================================
// ADMIN — InitNoteUpload
// POST /admin/notes/upload/init
// ================================================================

func (ns *NotesService) InitNoteUpload(
	ctx context.Context,
	adminID string,
	req models.InitNoteUploadRequest,
) (models.InitNoteUploadResponse, error) {

	ns.logger.Info("init note upload started",
		"admin_id", adminID,
		"module_id", req.ModuleID,
		"file_type", req.FileType,
		"original_filename", req.OriginalFilename,
	)

	// 1. Resolve module → subject → semester chain
	module, err := ns.notesrepo.GetModuleByID(ctx, req.ModuleID)
	if err != nil {
		ns.logger.Error("failed to get module", "module_id", req.ModuleID, "error", err)
		return models.InitNoteUploadResponse{}, err
	}

	subject, err := ns.notesrepo.GetSubjectByID(ctx, module.SubjectID)
	if err != nil {
		ns.logger.Error("failed to get subject", "subject_id", module.SubjectID, "error", err)
		return models.InitNoteUploadResponse{}, err
	}

	semester, err := ns.notesrepo.GetBranchSemesterByID(ctx, subject.BranchSemesterID)
	if err != nil {
		ns.logger.Error("failed to get semester", "semester_id", subject.BranchSemesterID, "error", err)
		return models.InitNoteUploadResponse{}, err
	}

	// 2. Verify admin belongs to this branch
	admin, err := ns.notesrepo.GetDeptAdminByID(ctx, adminID)
	if err != nil {
		ns.logger.Error("failed to get department admin", "admin_id", adminID, "error", err)
		return models.InitNoteUploadResponse{}, err
	}

	if admin.BranchID != semester.BranchID {
		ns.logger.Warn("branch mismatch — admin tried to upload to wrong branch",
			"admin_id", adminID,
			"admin_branch_id", admin.BranchID,
			"target_branch_id", semester.BranchID,
		)
		return models.InitNoteUploadResponse{}, apperror.BadRequest(
			"branch_mismatch",
			"you are not authorised to upload notes for this branch",
		)
	}

	// 3. Build S3 key and generate presigned PUT URL
	noteID := uuid.NewString()
	ext := noteFileExt(req.OriginalFilename, req.FileType)
	fileKey := noteS3Key(semester.BranchID, semester.ID, subject.ID, module.ID, noteID, ext)

	ns.logger.Info("note s3 key built", "note_id", noteID, "file_key", fileKey)

	uploadURL, err := ns.s3.PresignPut(ctx, fileKey)
	log.Printf("Presigned URL: %s", uploadURL)
	if err != nil {
		ns.logger.Error("failed to generate presigned put url", "file_key", fileKey, "error", err)
		return models.InitNoteUploadResponse{}, apperror.Internal("failed to generate upload url", err)
	}

	// 4. Persist note record with status = pending
	note := models.Note{
		ID:               noteID,
		ModuleID:         req.ModuleID,
		UploadedBy:       adminID,
		Title:            req.Title,
		Description:      req.Description,
		FileKey:          fileKey,
		FileType:         req.FileType,
		OriginalFilename: req.OriginalFilename,
		FileSizeBytes:    req.FileSizeBytes,
		Status:           "pending",
	}

	if err := ns.notesrepo.CreateNote(ctx, note); err != nil {
		ns.logger.Error("failed to create note record", "note_id", noteID, "error", err)
		return models.InitNoteUploadResponse{}, err
	}

	ns.logger.Info("init note upload completed",
		"admin_id", adminID,
		"note_id", noteID,
		"file_key", fileKey,
	)

	return models.InitNoteUploadResponse{
		NoteID:    noteID,
		UploadURL: uploadURL,
	}, nil
}

// ================================================================
// ADMIN — ConfirmNoteUpload
// POST /admin/notes/upload/confirm
// ================================================================

func (ns *NotesService) ConfirmNoteUpload(
	ctx context.Context,
	adminID string,
	req models.ConfirmNoteUploadRequest,
) error {

	ns.logger.Info("confirm note upload started", "admin_id", adminID, "note_id", req.NoteID)

	note, err := ns.notesrepo.GetNoteByID(ctx, req.NoteID)
	if err != nil {
		ns.logger.Error("failed to get note", "note_id", req.NoteID, "error", err)
		return err
	}

	if note.UploadedBy != adminID {
		ns.logger.Warn("confirm attempted by wrong admin",
			"note_id", req.NoteID,
			"note_uploaded_by", note.UploadedBy,
			"requesting_admin", adminID,
		)
		return apperror.BadRequest("unauthorised", "you did not initiate this upload")
	}

	if note.Status != "pending" {
		ns.logger.Warn("confirm attempted on non-pending note",
			"note_id", req.NoteID,
			"current_status", note.Status,
		)
		return apperror.BadRequest(
			"invalid_note_status",
			fmt.Sprintf("note is in status '%s' and cannot be confirmed", note.Status),
		)
	}

	exists, err := ns.s3.FileExists(ctx, note.FileKey)
	if err != nil {
		ns.logger.Error("s3 existence check failed", "note_id", req.NoteID, "error", err)
		return apperror.Internal("failed to verify file in storage", err)
	}
	if !exists {
		ns.logger.Warn("file not found in s3 during confirm", "note_id", req.NoteID, "file_key", note.FileKey)
		return apperror.BadRequest("file_not_uploaded", "file was not found in storage, please re-upload")
	}

	if err := ns.notesrepo.UpdateNoteStatus(ctx, note.ID, "active"); err != nil {
		ns.logger.Error("failed to activate note", "note_id", req.NoteID, "error", err)
		return err
	}

	ns.logger.Info("confirm note upload completed", "admin_id", adminID, "note_id", req.NoteID)

	return nil
}

// ================================================================
// ADMIN — UpdateNote
// PUT /admin/notes/:note_id
// ================================================================

func (ns *NotesService) UpdateNote(
	ctx context.Context,
	adminID string,
	noteID string,
	req models.UpdateNoteRequest,
) error {

	ns.logger.Info("update note started", "admin_id", adminID, "note_id", noteID)

	note, err := ns.notesrepo.GetNoteByID(ctx, noteID)
	if err != nil {
		ns.logger.Error("failed to get note for update", "note_id", noteID, "error", err)
		return err
	}

	if note.UploadedBy != adminID {
		ns.logger.Warn("update attempted by wrong admin",
			"note_id", noteID,
			"note_uploaded_by", note.UploadedBy,
			"requesting_admin", adminID,
		)
		return apperror.BadRequest("unauthorised", "you are not authorised to edit this note")
	}

	if note.Status == "deleted" {
		ns.logger.Warn("update attempted on deleted note", "note_id", noteID)
		return apperror.BadRequest("note_deleted", "this note has been deleted and cannot be edited")
	}

	if err := ns.notesrepo.UpdateNote(ctx, noteID, req.Title, req.Description); err != nil {
		ns.logger.Error("failed to update note", "note_id", noteID, "error", err)
		return err
	}

	ns.logger.Info("update note completed", "admin_id", adminID, "note_id", noteID)

	return nil
}

// ================================================================
// ADMIN — DeleteNote
// DELETE /admin/notes/:note_id
// ================================================================

func (ns *NotesService) DeleteNote(
	ctx context.Context,
	adminID string,
	noteID string,
) error {

	ns.logger.Info("delete note started", "admin_id", adminID, "note_id", noteID)

	note, err := ns.notesrepo.GetNoteByID(ctx, noteID)
	if err != nil {
		ns.logger.Error("failed to get note for deletion", "note_id", noteID, "error", err)
		return err
	}

	if note.UploadedBy != adminID {
		ns.logger.Warn("delete attempted by wrong admin",
			"note_id", noteID,
			"note_uploaded_by", note.UploadedBy,
			"requesting_admin", adminID,
		)
		return apperror.BadRequest("unauthorised", "you are not authorised to delete this note")
	}

	if note.Status == "deleted" {
		ns.logger.Warn("delete attempted on already deleted note", "note_id", noteID)
		return apperror.BadRequest("note_already_deleted", "this note has already been deleted")
	}

	if err := ns.notesrepo.DeleteNote(ctx, noteID); err != nil {
		ns.logger.Error("failed to delete note", "note_id", noteID, "error", err)
		return err
	}

	ns.logger.Info("delete note completed", "admin_id", adminID, "note_id", noteID)

	return nil
}

// ================================================================
// ADMIN — CreateSubject
// POST /admin/subjects
// ================================================================

func (ns *NotesService) CreateSubject(
	ctx context.Context,
	adminID string,
	req models.CreateSubjectRequest,
) (models.SubjectResponse, error) {

	ns.logger.Info("create subject started",
		"admin_id", adminID,
		"branch_semester_id", req.BranchSemesterID,
		"subject_code", req.SubjectCode,
	)

	semester, err := ns.notesrepo.GetBranchSemesterByID(ctx, req.BranchSemesterID)
	if err != nil {
		ns.logger.Error("failed to get semester", "branch_semester_id", req.BranchSemesterID, "error", err)
		return models.SubjectResponse{}, err
	}

	admin, err := ns.notesrepo.GetDeptAdminByID(ctx, adminID)
	if err != nil {
		ns.logger.Error("failed to get department admin", "admin_id", adminID, "error", err)
		return models.SubjectResponse{}, err
	}

	if admin.BranchID != semester.BranchID {
		ns.logger.Warn("branch mismatch — admin tried to create subject in wrong branch",
			"admin_id", adminID,
			"admin_branch_id", admin.BranchID,
			"target_branch_id", semester.BranchID,
		)
		return models.SubjectResponse{}, apperror.BadRequest(
			"branch_mismatch",
			"you are not authorised to create subjects for this branch",
		)
	}

	subjectID := uuid.NewString()
	subject := models.Subject{
		ID:               subjectID,
		BranchSemesterID: req.BranchSemesterID,
		Name:             req.Name,
		SubjectCode:      req.SubjectCode,
	}

	if err := ns.notesrepo.CreateSubject(ctx, subject); err != nil {
		ns.logger.Error("failed to create subject", "subject_id", subjectID, "error", err)
		return models.SubjectResponse{}, err
	}

	ns.logger.Info("create subject completed",
		"admin_id", adminID,
		"subject_id", subjectID,
		"subject_code", req.SubjectCode,
	)

	return models.SubjectResponse{
		ID:          subjectID,
		Name:        req.Name,
		SubjectCode: req.SubjectCode,
	}, nil
}

// ================================================================
// STUDENT — ListBranches
// GET /notes/branches
// ================================================================

func (ns *NotesService) ListBranches(ctx context.Context) ([]models.BranchResponse, error) {

	ns.logger.Info("list branches started")

	branches, err := ns.notesrepo.GetActiveBranches(ctx)
	if err != nil {
		ns.logger.Error("failed to get branches", "error", err)
		return nil, err
	}

	resp := make([]models.BranchResponse, 0, len(branches))
	for _, b := range branches {
		resp = append(resp, models.BranchResponse{
			ID:   b.ID,
			Name: b.Name,
			Code: b.Code,
		})
	}

	ns.logger.Info("list branches completed", "count", len(resp))

	return resp, nil
}

// ================================================================
// STUDENT — ListSemesters
// GET /notes/branches/:branch_id/semesters
// ================================================================

func (ns *NotesService) ListSemesters(
	ctx context.Context,
	branchID string,
) ([]models.SemesterResponse, error) {

	ns.logger.Info("list semesters started", "branch_id", branchID)

	// Verify branch exists first
	if _, err := ns.notesrepo.GetBranchByID(ctx, branchID); err != nil {
		ns.logger.Error("failed to get branch", "branch_id", branchID, "error", err)
		return nil, err
	}

	semesters, err := ns.notesrepo.GetSemestersByBranchID(ctx, branchID)
	if err != nil {
		ns.logger.Error("failed to get semesters", "branch_id", branchID, "error", err)
		return nil, err
	}

	resp := make([]models.SemesterResponse, 0, len(semesters))
	for _, s := range semesters {
		resp = append(resp, models.SemesterResponse{
			ID:             s.ID,
			SemesterNumber: s.SemesterNumber,
		})
	}

	ns.logger.Info("list semesters completed", "branch_id", branchID, "count", len(resp))

	return resp, nil
}

// ================================================================
// STUDENT — ListSubjects
// GET /notes/semesters/:semester_id/subjects
// ================================================================

func (ns *NotesService) ListSubjects(
	ctx context.Context,
	semesterID string,
) ([]models.SubjectListResponse, error) {

	ns.logger.Info("list subjects started", "semester_id", semesterID)

	// Verify semester exists first
	if _, err := ns.notesrepo.GetBranchSemesterByID(ctx, semesterID); err != nil {
		ns.logger.Error("failed to get semester", "semester_id", semesterID, "error", err)
		return nil, err
	}

	subjects, err := ns.notesrepo.GetSubjectsBySemesterID(ctx, semesterID)
	if err != nil {
		ns.logger.Error("failed to get subjects", "semester_id", semesterID, "error", err)
		return nil, err
	}

	resp := make([]models.SubjectListResponse, 0, len(subjects))
	for _, s := range subjects {
		resp = append(resp, models.SubjectListResponse{
			ID:          s.ID,
			Name:        s.Name,
			SubjectCode: s.SubjectCode,
		})
	}

	ns.logger.Info("list subjects completed", "semester_id", semesterID, "count", len(resp))

	return resp, nil
}

// ================================================================
// STUDENT — ListModules
// GET /notes/subjects/:subject_id/modules
// ================================================================

func (ns *NotesService) ListModules(
	ctx context.Context,
	subjectID string,
) ([]models.ModuleResponse, error) {

	ns.logger.Info("list modules started", "subject_id", subjectID)

	// Verify subject exists first
	if _, err := ns.notesrepo.GetSubjectByID(ctx, subjectID); err != nil {
		ns.logger.Error("failed to get subject", "subject_id", subjectID, "error", err)
		return nil, err
	}

	modules, err := ns.notesrepo.GetModulesBySubjectID(ctx, subjectID)
	if err != nil {
		ns.logger.Error("failed to get modules", "subject_id", subjectID, "error", err)
		return nil, err
	}

	resp := make([]models.ModuleResponse, 0, len(modules))
	for _, m := range modules {
		resp = append(resp, models.ModuleResponse{
			ID:           m.ID,
			ModuleNumber: m.ModuleNumber,
			Title:        m.Title,
		})
	}

	ns.logger.Info("list modules completed", "subject_id", subjectID, "count", len(resp))

	return resp, nil
}

// ================================================================
// STUDENT — ListNotes
// GET /notes/modules/:module_id/notes
// ================================================================

func (ns *NotesService) ListNotes(
	ctx context.Context,
	moduleID string,
) ([]models.NoteListItem, error) {

	ns.logger.Info("list notes started", "module_id", moduleID)

	// Verify module exists first
	if _, err := ns.notesrepo.GetModuleByID(ctx, moduleID); err != nil {
		ns.logger.Error("failed to get module", "module_id", moduleID, "error", err)
		return nil, err
	}

	notes, err := ns.notesrepo.GetNotesByModuleID(ctx, moduleID)
	if err != nil {
		ns.logger.Error("failed to get notes", "module_id", moduleID, "error", err)
		return nil, err
	}

	// Map to safe response — file_key is NEVER included
	resp := make([]models.NoteListItem, 0, len(notes))
	for _, n := range notes {
		resp = append(resp, models.NoteListItem{
			ID:               n.ID,
			Title:            n.Title,
			Description:      n.Description,
			FileType:         n.FileType,
			OriginalFilename: n.OriginalFilename,
			FileSizeBytes:    n.FileSizeBytes,
			CreatedAt:        n.CreatedAt,
		})
	}

	ns.logger.Info("list notes completed", "module_id", moduleID, "count", len(resp))

	return resp, nil
}

// ================================================================
// STUDENT — GetNoteViewURL
// GET /notes/:note_id/view
// Opens presigned S3 URL directly in browser as a PDF viewer
// ================================================================

func (ns *NotesService) GetNoteViewURL(
	ctx context.Context,
	noteID string,
) (models.NoteViewResponse, error) {

	ns.logger.Info("get note view url started", "note_id", noteID)

	note, err := ns.notesrepo.GetNoteByID(ctx, noteID)
	if err != nil {
		ns.logger.Error("failed to get note for view", "note_id", noteID, "error", err)
		return models.NoteViewResponse{}, err
	}

	if note.Status != "active" {
		ns.logger.Warn("view attempted on non-active note",
			"note_id", noteID,
			"status", note.Status,
		)
		return models.NoteViewResponse{}, apperror.BadRequest(
			"note_not_available",
			"this note is not available for viewing",
		)
	}

	viewURL, err := ns.s3.PresignGet(ctx, note.FileKey)
	if err != nil {
		ns.logger.Error("failed to generate presigned get url",
			"note_id", noteID,
			"file_key", note.FileKey,
			"error", err,
		)
		return models.NoteViewResponse{}, apperror.Internal("failed to generate view url", err)
	}

	expiresAt := time.Now().Add(15 * time.Minute)

	ns.logger.Info("get note view url completed", "note_id", noteID, "expires_at", expiresAt)

	return models.NoteViewResponse{
		ViewURL:   viewURL,
		ExpiresAt: expiresAt,
	}, nil
}

// ================================================================
// STUDENT — NotesPrintInit
// POST /notes/print/init
// Creates a print session reusing the note's existing S3 file
// ================================================================

func (ns *NotesService) NotesPrintInit(
	ctx context.Context,
	userID string,
	userEmail string,
	req models.NotesPrintInitRequest,
) (models.InitUploadResponse, error) {

	ns.logger.Info("notes print init started",
		"user_id", userID,
		"note_count", len(req.Notes),
	)

	// 1. Validate every note and calculate price up front
	type enrichedItem struct {
		note   models.Note
		item   models.NotePrintItem
		price  float64
		sheets int
	}

	enriched := make([]enrichedItem, 0, len(req.Notes))

	for _, item := range req.Notes {
		note, err := ns.notesrepo.GetNoteByID(ctx, item.NoteID)
		if err != nil {
			ns.logger.Error("failed to get note for print init",
				"note_id", item.NoteID,
				"error", err,
			)
			return models.InitUploadResponse{}, err
		}

		if note.Status != "active" {
			ns.logger.Warn("print attempted on non-active note",
				"note_id", item.NoteID,
				"status", note.Status,
			)
			return models.InitUploadResponse{}, apperror.BadRequest(
				"note_not_available",
				fmt.Sprintf("note '%s' is not available for printing", note.Title),
			)
		}

		price, sheets := utils.CalculateFilePrice(
			item.NumOfPages,
			item.PageRange,
			item.Copies,
			item.PageLayout,
			item.PrintingMode,
			item.PrintingSide,
		)

		ns.logger.Info("note price calculated",
			"note_id", item.NoteID,
			"price", price,
			"sheets", sheets,
		)

		enriched = append(enriched, enrichedItem{
			note:   note,
			item:   item,
			price:  price,
			sheets: sheets,
		})
	}

	// 2. Build session
	sessionID := uuid.NewString()
	tokenInt, err := generateToken()
	if err != nil {
		return models.InitUploadResponse{}, apperror.Internal("failed to generate session token", err)
	}
	tokenStr := strconv.Itoa(tokenInt)

	session := models.UploadSession{
		ID:        sessionID,
		UserID:    userID,
		UserEmail: userEmail,
		Status:    "created",
		Token:     tokenStr,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	// 3. Build upload_files reusing note S3 keys directly
	var totalAmount float64
	var totalSheets int
	dbFiles := make([]models.UploadFile, 0, len(enriched))
	responseFiles := make([]models.InitFileResponse, 0, len(enriched))

	for _, e := range enriched {
		fileID := uuid.NewString()
		totalAmount += e.price
		totalSheets += e.sheets

		// Key design: staging_key = final_key = note's existing S3 path
		// file_status = "promoted" — file already in S3, no upload step needed
		// The printer reads final_key, which points directly to the note file
		dbFiles = append(dbFiles, models.UploadFile{
			ID:            fileID,
			SessionID:     sessionID,
			FileName:      e.note.OriginalFilename,
			StagingKey:    e.note.FileKey,
			FinalKey:      &e.note.FileKey,
			PrintingMode:  &e.item.PrintingMode,
			PrintingSide:  &e.item.PrintingSide,
			PageRange:     e.item.PageRange,
			PageLayout:    &e.item.PageLayout,
			Copies:        &e.item.Copies,
			NumberOfPages: &e.item.NumOfPages,
			Price:         &e.price,
			FileStatus:    "promoted",
		})

		responseFiles = append(responseFiles, models.InitFileResponse{
			FileID:   fileID,
			FileName: e.note.OriginalFilename,
			// UploadURL is empty — no S3 upload needed, file already exists
			UploadURL: "",
		})
	}

	// 4. Persist session + files + price in one transaction
	tx, err := ns.db.BeginTx(ctx, nil)
	if err != nil {
		return models.InitUploadResponse{}, apperror.Internal("failed to begin transaction", err)
	}
	defer tx.Rollback()

	txRepo := ns.filerepo.WithTx(tx)

	if err := txRepo.CreateSession(ctx, session); err != nil {
		ns.logger.Error("failed to create print session", "session_id", sessionID, "error", err)
		return models.InitUploadResponse{}, err
	}

	if err := txRepo.CreateFiles(ctx, dbFiles); err != nil {
		ns.logger.Error("failed to create print session files", "session_id", sessionID, "error", err)
		return models.InitUploadResponse{}, err
	}

	// Price is fully known at init time — skip the confirm step entirely
	if err := txRepo.UpdateSessionPriced(ctx, sessionID, totalAmount, totalSheets); err != nil {
		ns.logger.Error("failed to price session", "session_id", sessionID, "error", err)
		return models.InitUploadResponse{}, err
	}

	if err := tx.Commit(); err != nil {
		return models.InitUploadResponse{}, apperror.Internal("failed to commit transaction", err)
	}

	ns.logger.Info("notes print init completed",
		"user_id", userID,
		"session_id", sessionID,
		"note_count", len(enriched),
		"total_amount", totalAmount,
		"total_sheets", totalSheets,
	)

	return models.InitUploadResponse{
		SessionID: sessionID,
		Token:     tokenInt,
		ExpiresAt: session.ExpiresAt,
		Files:     responseFiles,
	}, nil
}

// ================================================================
// Private helpers
// ================================================================

func noteFileExt(originalFilename, fileType string) string {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(originalFilename), "."))
	if ext == "" {
		return fileType
	}
	return ext
}

func noteS3Key(branchID, semesterID, subjectID, moduleID, noteID, ext string) string {
	return fmt.Sprintf(
		"notes/%s/%s/%s/%s/%s.%s",
		branchID, semesterID, subjectID, moduleID, noteID, ext,
	)
}