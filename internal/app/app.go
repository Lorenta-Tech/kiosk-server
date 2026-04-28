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
	"github.com/Lorenta-Tech/kiosk-server/pkg/s3"
	"github.com/Lorenta-Tech/kiosk-server/pkg/utils"
	_ "github.com/joho/godotenv/autoload"
)

type Application struct {
	DB          *sql.DB
	Logger      *slog.Logger
	S3          *s3.Client
	FileHandler *handler.FileHandler
	UserHandler *handler.UserHandler
	JWTSecret   string
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

	logger.Info("s3 client initialized", "bucket", env.GetString("BUCKET", "aiet-printflow-upload-prod"))

	jwtSecret := env.GetString("JWT_SECRET", "")
	googleClientID := env.GetString("GOOGLE_CLIENT_ID", "")

	// File feature
	filerepo := repository.NewFileRepository(pgdb)
	fileservice := service.NewFileService(filerepo, s3Client, pgdb, logger)
	fileHandler := handler.NewFileHandler(fileservice, logger)

	// User / Auth feature
	userrepo := repository.NewUserRepository(pgdb)
	userservice := service.NewUserService(userrepo, logger, jwtSecret, googleClientID)
	userHandler := handler.NewUserHandler(userservice, logger)

	app := &Application{
		DB:          pgdb,
		Logger:      logger,
		S3:          s3Client,
		FileHandler: fileHandler,
		UserHandler: userHandler,
		JWTSecret:   jwtSecret,
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
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{
		"status": "healthy",
		"db":     "ok",
	})
}
