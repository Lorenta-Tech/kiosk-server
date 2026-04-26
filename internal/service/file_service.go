package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log/slog"
	"time"

	"github.com/Lorenta-Tech/kiosk-server/internal/models"
	"github.com/Lorenta-Tech/kiosk-server/internal/repository"
	"github.com/Lorenta-Tech/kiosk-server/pkg/apperror"
	s3pkg "github.com/Lorenta-Tech/kiosk-server/pkg/s3"
	"github.com/Lorenta-Tech/kiosk-server/pkg/utils"
	"github.com/google/uuid"
)

type FileService struct {
	filerepo repository.FileRepo
	s3       *s3pkg.Client
	db       *sql.DB   
	logger   *slog.Logger
}

func NewFileService(
	filerepo repository.FileRepo,
	s3 *s3pkg.Client,
	db *sql.DB,
	logger *slog.Logger,
) *FileService {
	return &FileService{
		filerepo: filerepo,
		s3:       s3,
		db:       db,
		logger:   logger,
	}
}

// InitUpload is the service entry point for POST /files/upload/init.
//
// It does the following in order:
//  1. Opens a DB transaction
//  2. Creates the upload_session row
//  3. For each file: generates a staging key + presigned PUT URL
//  4. Inserts all upload_files rows inside the same transaction
//  5. Commits — if anything above fails, the transaction is rolled back
//     automatically and no partial data is left in the DB
func (fs *FileService) InitUpload(
	ctx context.Context,
	userID string,
	userEmail string,
	req models.InitUploadRequest,
) (models.InitUploadResponse, error) {

	fs.logger.Info("init upload started",
		"user_id", userID,
		"file_count", len(req.Files),
	)
	
	//Opens DB Transaction
	tx, err := fs.db.BeginTx(ctx, nil)
	if err != nil {
		fs.logger.Error("init upload failed to open transaction",
			"user_id", userID,
			"error", err,
		)
		return models.InitUploadResponse{}, apperror.Internal("failed to begin transaction", err)
	}
	defer tx.Rollback() // no-op if tx.Commit()

	txRepo := fs.filerepo.WithTx(tx)

	sessionID := uuid.NewString()
	token, err := generateToken()
	if err != nil {
		return models.InitUploadResponse{}, apperror.Internal("failed to generate session token", err)
	}

	session := models.UploadSession{
		ID:        sessionID,
		UserID:    userID,
		UserEmail: userEmail,
		Status:    "created",
		Token:     token,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	if err := txRepo.CreateSession(ctx, session); err != nil {
		// err is already an *apperror.AppError from the repository layer
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
			fs.logger.Error("failed to generate presigned url",
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

	// Commit 
	if err := tx.Commit(); err != nil {
		fs.logger.Error("init upload failed to commit transaction",
			"session_id", sessionID,
			"user_id", userID,
			"error", err,
		)
		return models.InitUploadResponse{}, apperror.Internal("failed to commit transaction", err)
	}

	fs.logger.Info("init upload completed",
		"session_id", sessionID,
		"user_id", userID,
		"file_count", len(dbFiles),
	)

	return models.InitUploadResponse{
		SessionID: sessionID,
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
			f.NumOfPages,
			f.Copies,
			f.PageLayout,
			f.PrintingMode,
			f.PrintingSide,
		)
 
		totalAmount += price
		totalSheets += sheets

		dbFile.PrintingMode  = &f.PrintingMode
		dbFile.PrintingSide  = &f.PrintingSide
		dbFile.PageRange     = &f.PageRange
		dbFile.PageLayout    = &f.PageLayout
		dbFile.Copies        = &f.Copies
		dbFile.NumberOfPages = &f.NumOfPages
		dbFile.Price         = &price
 
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
 
		fs.logger.Info("file confirmed",
			"session_id", req.SessionID,
			"file_id", f.FileID,
			"file_name", dbFile.FileName,
			"sheets", sheets,
			"price", price,
		)
	}
 
	// Begin Transaction
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
		Status:      "priced",
		Files:       responseFiles,
		TotalSheets: totalSheets,
		TotalAmount: totalAmount,
	}, nil
}

// generateToken returns a 32-byte cryptographically random hex string.
func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

