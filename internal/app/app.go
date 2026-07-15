package app

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/Lorenta-Tech/kiosk-server/internal/env"
	"github.com/Lorenta-Tech/kiosk-server/internal/handler"
	"github.com/Lorenta-Tech/kiosk-server/internal/repository"
	"github.com/Lorenta-Tech/kiosk-server/internal/service"
	"github.com/Lorenta-Tech/kiosk-server/pkg/db"
	"github.com/Lorenta-Tech/kiosk-server/pkg/mail"
	"github.com/Lorenta-Tech/kiosk-server/pkg/s3"
	"github.com/Lorenta-Tech/kiosk-server/pkg/utils"
	_ "github.com/joho/godotenv/autoload"
)

type Application struct {
	DB               *sql.DB
	Logger           *slog.Logger
	S3               *s3.Client
	FileHandler      *handler.FileHandler
	UserHandler      *handler.UserHandler
	PaymentHandler   *handler.PaymentHandler
	AdminHandler     *handler.AdminHandler
	NotesHandler     *handler.NotesHandler
	DeptAdminHandler *handler.DeptAdminHandler
	JWTSecret        string
}

func NewApplication() (*Application, error) {
	if err := env.Require(
		"DATABASE_HOST",
		"DATABASE_PORT",
		"DATABASE_USER",
		"DATABASE_NAME",
		"DATABASE_PASSWORD",
		"REGION",
		"ACCESS_KEY",
		"SECRETE_KEY",
		"BUCKET",
		"JWT_SECRET",
		"GOOGLE_CLIENT_ID",
		"RZP_KEY",
		"RZP_SECRET",
		"RZP_WEBHOOK_SECRET",
	); err != nil {
		return nil, err
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	//database
	pgdb, err := db.Connect()
	if err != nil {
		return nil, err
	}

	pgdb.SetMaxOpenConns(25)
	pgdb.SetMaxIdleConns(10)
	pgdb.SetConnMaxLifetime(5 * time.Minute)
	pgdb.SetConnMaxIdleTime(2 * time.Minute)

	{
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := pgdb.PingContext(ctx); err != nil {
			return nil, fmt.Errorf("database unreachable:%w", err)
		}
	}

	//s3 connection
	s3Client, err := s3.Connect()
	if err != nil {
		return nil, fmt.Errorf("s3 unreachable: %w", err)
	}

	//mail client
	mailClient, err := mail.NewResendClient()
	if err != nil {
		return nil, fmt.Errorf("failed to innitialize mail client:%w", err)
	}

	logger.Info("s3 client initialized", "bucket", env.GetString("BUCKET", "aiet-printflow-upload-prod"))

	jwtSecret := env.GetString("JWT_SECRET", "")
	googleClientID := env.GetString("GOOGLE_CLIENT_ID", "")
	razorpayKey := env.GetString("RZP_KEY", "")
	razorpaySecret := env.GetString("RZP_SECRET", "")
	webhookSecret := env.GetString("RZP_WEBHOOK_SECRET", "")

	//repositories
	userrepo := repository.NewUserRepository(pgdb)
	filerepo := repository.NewFileRepository(pgdb)
	paymentRepo := repository.NewPaymentRepository(pgdb)
	notesrepo := repository.NewNotesRepository(pgdb)
	deptadminRepo := repository.NewDeptAdminRepository(pgdb)
	adminrepo := repository.NewAdminRepository(pgdb)

	//services
	userservice := service.NewUserService(userrepo, logger, jwtSecret, googleClientID)
	fileservice := service.NewFileService(filerepo, notesrepo, userrepo, s3Client, pgdb, mailClient, logger)
	paymentService := service.NewPaymentService(paymentRepo, filerepo, pgdb, s3Client, logger, razorpayKey, razorpaySecret, webhookSecret)
	notesservice := service.NewNotesService(notesrepo, filerepo, pgdb, s3Client, logger)
	deptAdminService := service.NewDeptAdminService(
		deptadminRepo,
		logger,
		jwtSecret,
		env.GetString("SUPER_ADMIN_EMAIL", ""), //this two needs to add to required env variables
		env.GetString("SUPER_ADMIN_PASSWORD", ""),
	)
	adminservice := service.NewAdminRepo(filerepo, adminrepo, logger)

	//Handlers
	userHandler := handler.NewUserHandler(userservice, logger)
	fileHandler := handler.NewFileHandler(fileservice, logger)
	paymentHandler := handler.NewPaymentHandler(paymentService, logger)
	notesHandler := handler.NewNotesHandler(notesservice, logger)
	deptAdminHandler := handler.NewDeptAdminHandler(deptAdminService, logger)
	adminHandler := handler.NewAdminHandler(adminservice, logger)

	app := &Application{
		DB:               pgdb,
		Logger:           logger,
		S3:               s3Client,
		FileHandler:      fileHandler,
		UserHandler:      userHandler,
		JWTSecret:        jwtSecret,
		PaymentHandler:   paymentHandler,
		AdminHandler:     adminHandler,
		NotesHandler:     notesHandler,
		DeptAdminHandler: deptAdminHandler,
	}

	return app, nil
}

func (a *Application) HealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := a.DB.PingContext(ctx); err != nil {
		utils.WriteJSON(w, http.StatusServiceUnavailable, utils.Envelope{
			"status": "unhealthy",
			"db":     "unreachable",
		})
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{
		"status": "healthy",
		"db":     "ok",
	})
}
