package service

import (
	"context"
	"log/slog"

	"github.com/Lorenta-Tech/kiosk-server/internal/models"
	"github.com/Lorenta-Tech/kiosk-server/internal/repository"
)

type AdminRepo struct {
	filerepo   repository.FileRepo
	adminrepo repository.AdminRepo
	logger     *slog.Logger
}

func NewAdminRepo(filerepo repository.FileRepo, adminrepo repository.AdminRepo, logger *slog.Logger) *AdminRepo {
	return &AdminRepo{
		filerepo:   filerepo,
		adminrepo: adminrepo,
		logger:     logger,
	}
}

func (s *AdminRepo) AdminFetchPrintHistory(ctx context.Context) ([]models.PrintJob, error) {
	// Implement the logic to fetch print history
	// This is a placeholder for the actual implementation
	
	s.logger.Info("FetchPrintHistory called")
	return s.filerepo.AdminFetchPrintHistory(ctx)
}