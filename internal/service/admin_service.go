package service

import (
	"log/slog"

	"github.com/Lorenta-Tech/kiosk-server/internal/repository"
)

type AdminRepo struct {
	filerepo   repository.FileRepo
	adminrepo repository.AdminRepository
	logger     *slog.Logger
}

func NewAdminRepo(filerepo repository.FileRepo, adminrepo repository.AdminRepository, logger *slog.Logger) *AdminRepo {
	return &AdminRepo{
		filerepo:   filerepo,
		adminrepo: adminrepo,
		logger:     logger,
	}
}

func (s *AdminRepo) FetchPrintHistory() {
	// Implement the logic to fetch print history
	// This is a placeholder for the actual implementation
	s.logger.Info("FetchPrintHistory called")
}