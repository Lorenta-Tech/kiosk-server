package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"log/slog"
	"math/big"
	"strconv"
	"time"

	"github.com/Lorenta-Tech/kiosk-server/internal/models"
	"github.com/Lorenta-Tech/kiosk-server/internal/repository"
	"github.com/Lorenta-Tech/kiosk-server/pkg/apperror"
	"github.com/Lorenta-Tech/kiosk-server/pkg/mail"
	s3pkg "github.com/Lorenta-Tech/kiosk-server/pkg/s3"
	"github.com/Lorenta-Tech/kiosk-server/pkg/utils"
	"github.com/google/uuid"
)

const defaultRecentJobsLimit = 10

type FileService struct {
	filerepo   repository.FileRepo
	userrepo   repository.UserRepo
	s3         *s3pkg.Client
	db         *sql.DB
	mailclient *mail.ResendClient
	logger     *slog.Logger
}

func NewFileService(
	filerepo repository.FileRepo,
	userrepo repository.UserRepo,
	s3 *s3pkg.Client,
	db *sql.DB,
	mailclient *mail.ResendClient,
	logger *slog.Logger,
) *FileService {
	return &FileService{filerepo: filerepo, userrepo: userrepo, s3: s3, db: db, mailclient: mailclient, logger: logger}
}

func (fs *FileService) InitUpload(
	ctx context.Context,
	userID string,
	userEmail string,
	req models.InitUploadRequest,
) (models.InitUploadResponse, error) {

	fs.logger.Info("init upload started",
		"user_id", userID,
		"file_count", len(req.Files),
		"request_payload", req,
	)

	tx, err := fs.db.BeginTx(ctx, nil)
	if err != nil {
		return models.InitUploadResponse{}, apperror.Internal("failed to begin transaction", err)
	}
	defer tx.Rollback()

	txRepo := fs.filerepo.WithTx(tx)

	sessionID := uuid.NewString()

	// Generate 6-digit token and store as string in DB (VARCHAR column)
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

	if err := txRepo.CreateSession(ctx, session); err != nil {
		return models.InitUploadResponse{}, err
	}

	fs.logger.Info("upload session created",
		"session_id", sessionID,
		"user_id", userID,
		"expires_at", session.ExpiresAt,
	)

	dbFiles := make([]models.UploadFile, 0, len(req.Files))
	responseFiles := make([]models.InitFileResponse, 0, len(req.Files))

	for _, f := range req.Files {
		fileID := uuid.NewString()
		stagingKey := s3pkg.StagingKey(userID, sessionID, fileID)

		uploadURL, err := fs.s3.PresignPut(ctx, stagingKey)
		if err != nil {
			fs.logger.Error("presign failed",
				"session_id", sessionID,
				"file_name", f.FileName,
				"error", err,
			)
			return models.InitUploadResponse{}, apperror.Internal(
				"failed to generate upload URL for "+f.FileName, err,
			)
		}

		dbFiles = append(dbFiles, models.UploadFile{
			ID:         fileID,
			SessionID:  sessionID,
			FileName:   f.FileName,
			StagingKey: stagingKey,
			FileStatus: "pending",
		})

		responseFiles = append(responseFiles, models.InitFileResponse{
			FileID:     fileID,
			FileName:   f.FileName,
			UploadURL:  uploadURL,
			StagingKey: stagingKey,
		})
	}

	if err := txRepo.CreateFiles(ctx, dbFiles); err != nil {
		return models.InitUploadResponse{}, err
	}

	if err := tx.Commit(); err != nil {
		return models.InitUploadResponse{}, apperror.Internal("failed to commit transaction", err)
	}

	fs.logger.Info("init upload completed",
		"session_id", sessionID,
		"user_id", userID,
		"file_count", len(dbFiles),
	)

	return models.InitUploadResponse{
		SessionID: sessionID,
		Token:     tokenInt, // return as int so frontend shows 6-digit number cleanly
		ExpiresAt: session.ExpiresAt,
		Files:     responseFiles,
	}, nil
}

func (fs *FileService) ConfirmUpload(
	ctx context.Context,
	req models.ConfirmUploadRequest,
) (models.ConfirmUploadResponse, error) {

	fs.logger.Info("confirm upload started",
		"session_id", req.SessionID,
		"file_count", len(req.Files),
	)

	session, err := fs.filerepo.GetSessionByID(ctx, req.SessionID)
	if err != nil {
		return models.ConfirmUploadResponse{}, err
	}

	if session.Status != "created" && session.Status != "uploaded" {
		return models.ConfirmUploadResponse{}, apperror.BadRequest(
			"session_not_confirmable",
			fmt.Sprintf("session is in status '%s' and cannot be confirmed", session.Status),
		)
	}

	if time.Now().After(session.ExpiresAt) {
		return models.ConfirmUploadResponse{}, apperror.BadRequest(
			"session_expired",
			"this upload session has expired, please start again",
		)
	}

	fs.logger.Info("session validated for confirm",
		"session_id", req.SessionID,
		"status", session.Status,
	)

	type enrichedFile struct {
		dbRow    models.UploadFile
		response models.ConfirmFileResponse
	}

	enriched := make([]enrichedFile, 0, len(req.Files))
	var totalAmount float64
	var totalSheets int

	for _, f := range req.Files {
		dbFile, err := fs.filerepo.GetFileByID(ctx, f.FileID, req.SessionID)
		if err != nil {
			return models.ConfirmUploadResponse{}, err
		}

		//TODO : Debugging the file exist in s3 if not comment out this and check the flow
		exists, err := fs.s3.FileExists(ctx, dbFile.StagingKey)
		if err != nil {
			fs.logger.Error("s3 existence check failed",
				"session_id", req.SessionID,
				"file_id", f.FileID,
				"staging_key", dbFile.StagingKey,
				"error", err,
			)
			return models.ConfirmUploadResponse{}, apperror.Internal(
				"failed to verify file upload status", err,
			)
		}
		if !exists {
			fs.logger.Warn("file not found in s3 staging",
				"session_id", req.SessionID,
				"file_id", f.FileID,
				"staging_key", dbFile.StagingKey,
			)
			return models.ConfirmUploadResponse{}, apperror.BadRequest(
				"file_not_uploaded",
				fmt.Sprintf("file %s was not found in storage, please re-upload", dbFile.FileName),
			)
		}

		price, sheets := utils.CalculateFilePrice(
			f.NumOfPages, f.PageRange, f.Copies, f.PageLayout,
			f.PrintingMode, f.PrintingSide,
		)

		fs.logger.Info("file price calculated",
			"price", price,
			"sheets", sheets,
			"num_of_pages", f.NumOfPages,
			"copies", f.Copies,
			"page_layout", f.PageLayout,
			"printing_mode", f.PrintingMode,
			"printing_side", f.PrintingSide,
		)

		totalAmount += price
		totalSheets += sheets

		fs.logger.Info("Total done bebug", "TotalAmount:", totalAmount, "TotalSheetes:", totalSheets)

		dbFile.PrintingMode = &f.PrintingMode
		dbFile.PrintingSide = &f.PrintingSide
		dbFile.PageRange = f.PageRange
		dbFile.PageLayout = &f.PageLayout
		dbFile.Copies = &f.Copies
		dbFile.NumberOfPages = &f.NumOfPages
		dbFile.Price = &price

		enriched = append(enriched, enrichedFile{
			dbRow: dbFile,
			response: models.ConfirmFileResponse{
				FileID:     dbFile.ID,
				FileName:   dbFile.FileName,
				NumOfPages: f.NumOfPages,
				Copies:     f.Copies,
				Price:      price,
			},
		})

		fs.logger.Info("Page Range bebugging", "page_range:", dbFile.PageRange)

		fs.logger.Info("file confirmed",
			"session_id", req.SessionID,
			"file_id", f.FileID,
			"file_name", dbFile.FileName,
			"sheets", sheets,
			"price", price,
		)
	}

	tx, err := fs.db.BeginTx(ctx, nil)
	if err != nil {
		return models.ConfirmUploadResponse{}, apperror.Internal("failed to begin transaction", err)
	}
	defer tx.Rollback()

	txRepo := fs.filerepo.WithTx(tx)

	for _, ef := range enriched {
		if err := txRepo.UpdateFileWithPrintOptions(ctx, ef.dbRow); err != nil {
			return models.ConfirmUploadResponse{}, err
		}
	}

	if err := txRepo.UpdateSessionPriced(ctx, req.SessionID, totalAmount, totalSheets); err != nil {
		return models.ConfirmUploadResponse{}, err
	}

	if err := tx.Commit(); err != nil {
		return models.ConfirmUploadResponse{}, apperror.Internal("failed to commit transaction", err)
	}

	fs.logger.Info("confirm upload completed",
		"session_id", req.SessionID,
		"total_sheets", totalSheets,
		"total_amount", totalAmount,
	)

	responseFiles := make([]models.ConfirmFileResponse, 0, len(enriched))
	for _, ef := range enriched {
		responseFiles = append(responseFiles, ef.response)
	}

	return models.ConfirmUploadResponse{
		SessionID:   req.SessionID,
		Status:      "paid",
		Files:       responseFiles,
		TotalSheets: totalSheets,
		TotalAmount: totalAmount,
	}, nil
}

func (fs *FileService) GetRecentPrintJobs(
	ctx context.Context,
	userID string,
) (models.RecentPrintJobsResponse, error) {

	fs.logger.Info("get recent print jobs started", "user_id", userID)

	exist,err := fs.userrepo.CheckUserExists(ctx,userID)

	if !exist {
		return models.RecentPrintJobsResponse{}, apperror.NotFound(
			"user_not_found in the database, please logout and login again",
			"user not found",
		)
	}

	sessions, err := fs.filerepo.GetRecentPrintJobs(ctx, userID, defaultRecentJobsLimit)
	if err != nil {
		return models.RecentPrintJobsResponse{}, err
	}

	if len(sessions) == 0 {
		fs.logger.Info("no print jobs found", "user_id", userID)
		return models.RecentPrintJobsResponse{Jobs: []models.PrintJob{}, Total: 0}, nil
	}

	jobs, err := fs.buildPrintJobs(ctx, sessions)
	if err != nil {
		return models.RecentPrintJobsResponse{}, err
	}

	fs.logger.Info("get recent print jobs completed",
		"user_id", userID,
		"job_count", len(jobs),
	)

	return models.RecentPrintJobsResponse{Jobs: jobs, Total: len(jobs)}, nil
}

func (fs *FileService) GetActivePrintJobs(
	ctx context.Context,
	userID string,
) (models.RecentPrintJobsResponse, error) {

	fs.logger.Info("get recent print jobs started", "user_id", userID)

	exist,err := fs.userrepo.CheckUserExists(ctx,userID)
	if !exist {
		return models.RecentPrintJobsResponse{}, apperror.NotFound(
			"user_not_found in the database, please logout and login with old google account to get active orders",
			"user not found",
		)
	}

	sessions, err := fs.filerepo.GetActivePrintJobs(ctx, userID)
	if err != nil {
		return models.RecentPrintJobsResponse{}, err
	}

	if len(sessions) == 0 {
		fs.logger.Info("no print jobs found", "user_id", userID)
		return models.RecentPrintJobsResponse{Jobs: []models.PrintJob{}, Total: 0}, nil
	}

	jobs, err := fs.buildPrintJobs(ctx, sessions)
	if err != nil {
		return models.RecentPrintJobsResponse{}, err
	}

	fs.logger.Info("get recent print jobs completed",
		"user_id", userID,
		"job_count", len(jobs),
	)

	return models.RecentPrintJobsResponse{Jobs: jobs, Total: len(jobs)}, nil
}

func (fs *FileService) GetJobByToken(
	ctx context.Context,
	req models.GetJobByTokenRequest,
) (models.TokenJobResponse, error) {

	tokenStr := strconv.Itoa(req.Token)

	fs.logger.Info("get job by token started", "token", req.Token)

	session, err := fs.filerepo.GetSessionByToken(ctx, tokenStr)
	if err != nil {
		return models.TokenJobResponse{}, err
	}

	if session.Status != "paid" {
		fs.logger.Warn("token lookup rejected — wrong session status",
			"token", req.Token,
			"session_id", session.ID,
			"status", session.Status,
		)
		return models.TokenJobResponse{}, apperror.BadRequest(
			"token_not_ready",
			tokenStatusMessage(session.Status),
		)
	}

	// Check expiry
	if time.Now().After(session.ExpiresAt) {
		fs.logger.Warn("token lookup rejected — session expired",
			"token", req.Token,
			"session_id", session.ID,
			"expired_at", session.ExpiresAt,
		)
		return models.TokenJobResponse{}, apperror.BadRequest(
			"token_expired",
			"this token has expired, please start a new upload",
		)
	}

	files, err := fs.filerepo.GetFilesBySessionID(ctx, session.ID)
	if err != nil {
		return models.TokenJobResponse{}, err
	}

	printJobFiles := make([]models.PrintJobFile, 0, len(files))
	for _, f := range files {
		//url expires in 15 minutes
		url, err := fs.s3.PresignGet(ctx, *f.FinalKey) //need to put final key
		fs.logger.Info("final_Key_Bebugging", "final_key:", *f.FinalKey)
		if err != nil {
			fs.logger.Error("failed to presign get url for file",
				"session_id", session.ID,
				"file_id", f.ID,
				"staging_key", f.StagingKey,
				"error", err,
			)
			return models.TokenJobResponse{}, apperror.Internal(
				fmt.Sprintf("failed to generate download URL for file %s", f.FileName), err,
			)
		}

		pf := models.PrintJobFile{
			FileID:        f.ID,
			FileName:      f.FileName,
			PrintingMode:  f.PrintingMode,
			PrintingSide:  f.PrintingSide,
			PageRange:     f.PageRange,
			PageLayout:    f.PageLayout,
			Copies:        f.Copies,
			NumberOfPages: f.NumberOfPages,
			Price:         f.Price,
			FileStatus:    f.FileStatus,
			DownloadURL:   &url,
		}
		printJobFiles = append(printJobFiles, pf)
	}

	fs.logger.Info("get job by token completed",
		"token", req.Token,
		"session_id", session.ID,
		"file_count", len(files),
	)

	return models.TokenJobResponse{
		Job: models.PrintJob{
			SessionID:   session.ID,
			Status:      session.Status,
			TotalAmount: session.TotalAmount,
			TotalSheets: session.TotalSheets,
			CreatedAt:   session.CreatedAt,
			Files:       printJobFiles,
		},
	}, nil
}

func (fs *FileService) ExpireSessionAfterPrinting(
	ctx context.Context,
	req models.ExpireSessionRequest,
) error {

	fs.logger.Info("expire session after printing started",
		"session_id", req.SessionID,
	)

	session, err := fs.filerepo.GetSessionByID(ctx, req.SessionID)
	if err != nil {
		return err
	}

	if session.Status == "completed" {
		return apperror.BadRequest(
			"session_already_completed",
			"this session has already been completed",
		)
	}

	if err := fs.filerepo.ExpireSessionAfterPrinting(ctx, req.SessionID); err != nil {
		return err
	}

	fs.logger.Info("session marked completed",
		"session_id", req.SessionID,
	)

	return nil
}
func (fs *FileService) ErrorRequestFromPrinter(
	ctx context.Context,
	req models.ErrorRequestFromPrinter,
) error {

	fs.logger.Info(
		"printer error received",
		"session_id", req.SessionID,
		"printer_id", req.PrinterID,
		"error", req.Error,
	)

	template, err := mail.BuildPrinterErrorTemplate(
		req.Error,
		req.PrinterID,
		req.SessionID,
	)
	if err != nil {
		return apperror.BadRequest(
			"invalid_printer_error",
			err.Error(),
		)
	}

	recipients := []string{
		"suhasdeveloper07@gmail.com",
		"manojseetaram.dev@gmail.com",
		"prasannacharya428@gmail.com",
	}

	if err := fs.mailclient.Send(
		recipients,
		template.Subject,
		template.Body,
	); err != nil {

		fs.logger.Error(
			"failed to send printer alert email",
			"error", err,
		)

		return apperror.Internal(
			"failed to send printer alert email",
			err,
		)
	}

	fs.logger.Info(
		"printer alert email sent successfully",
		"error", req.Error,
	)

	return nil
}

func (fs *FileService) GetJobBySessionID(
	ctx context.Context,
	sessionId string,
) (models.TokenJobResponse, error) {
	fs.logger.Info("get job by session ID started", "session_id", sessionId)

	session, err := fs.filerepo.GetSessionByID(ctx, sessionId)
	if err != nil {
		return models.TokenJobResponse{}, err
	}

	files, err := fs.filerepo.GetFilesBySessionID(ctx, session.ID)
	if err != nil {
		return models.TokenJobResponse{}, err
	}

	printJobFiles := make([]models.PrintJobFile, 0, len(files))
	for _, f := range files {
		pf := models.PrintJobFile{
			FileID:        f.ID,
			FileName:      f.FileName,
			PrintingMode:  f.PrintingMode,
			PrintingSide:  f.PrintingSide,
			PageRange:     f.PageRange,
			PageLayout:    f.PageLayout,
			Copies:        f.Copies,
			NumberOfPages: f.NumberOfPages,
			Price:         f.Price,
			FileStatus:    f.FileStatus,
		}
		printJobFiles = append(printJobFiles, pf)
	}

	fs.logger.Info("get job by session ID completed",
		"session_id", sessionId,
		"file_count", len(files),
	)

	return models.TokenJobResponse{
		Job: models.PrintJob{
			SessionID:   session.ID,
			Status:      session.Status,
			TotalAmount: session.TotalAmount,
			TotalSheets: session.TotalSheets,
			CreatedAt:   session.CreatedAt,
			Files:       printJobFiles,
		},
	}, nil
}



func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func intSliceToStringSlice(nums []int) []string {

	result := make([]string, 0, len(nums))

	for _, n := range nums {
		result = append(result, strconv.Itoa(n))
	}

	return result
}

func stringPtrToIntPtr(s *string) *int {

	if s == nil {
		return nil
	}

	v, err := strconv.Atoi(*s)
	if err != nil {
		return nil
	}

	return &v
}
//helpers

func (fs *FileService) buildPrintJobs(ctx context.Context, sessions []models.UploadSession) ([]models.PrintJob, error) {
	jobs := make([]models.PrintJob, 0, len(sessions))
	for _, s := range sessions {
		files, err := fs.filerepo.GetFilesBySessionID(ctx, s.ID)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, models.PrintJob{
			SessionID:   s.ID,
			Status:      s.Status,
			Token:       s.Token,
			TotalAmount: s.TotalAmount,
			TotalSheets: s.TotalSheets,
			CreatedAt:   s.CreatedAt,
			Files:       buildPrintJobFiles(files),
		})
	}
	return jobs, nil
}

func buildPrintJobFiles(files []models.UploadFile) []models.PrintJobFile {
	result := make([]models.PrintJobFile, 0, len(files))
	for _, f := range files {
		result = append(result, models.PrintJobFile{
			FileID:        f.ID,
			FileName:      f.FileName,
			PrintingMode:  f.PrintingMode,
			PrintingSide:  f.PrintingSide,
			PageRange:     f.PageRange,
			PageLayout:    f.PageLayout,
			Copies:        f.Copies,
			NumberOfPages: f.NumberOfPages,
			Price:         f.Price,
			FileStatus:    f.FileStatus,
		})
	}
	return result
}

func tokenStatusMessage(status string) string {
	switch status {
	case "created", "uploaded":
		return "this job has not been confirmed yet, please complete the upload first"
	case "paid":
		return "this job has already been paid and processed"
	case "expired":
		return "this token has expired, please start a new upload"
	default:
		return "this token is not ready for printing"
	}
}

// generateToken returns a cryptographically random 6-digit integer (100000–999999).
func generateToken() (int, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(900000))
	if err != nil {
		return 0, err
	}
	return int(n.Int64()) + 100000, nil
}

// func safeString(v *string) string {
// 	if v == nil {
// 		return ""
// 	}
// 	return *v
// }

// func safeInt(v *int) int {
// 	if v == nil {
// 		return 0
// 	}
// 	return *v
// }

// func safeFloat(v *float64) float64 {
// 	if v == nil {
// 		return 0
// 	}
// 	return *v
// }
