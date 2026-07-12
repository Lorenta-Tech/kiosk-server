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
	return s.adminrepo.AdminFetchPrintHistory(ctx)
}

func (s *AdminRepo) AdminGetTotalRevenue(ctx context.Context) (float64, error) {
	s.logger.Info("AdminGetTotalRevenue called")
	return s.adminrepo.AdminGetTotalRevenue(ctx)
}

func (s *AdminRepo) AdminGetTotalSheetsPrinted(ctx context.Context) (int, error) {
	s.logger.Info("AdminGetTotalSheetsPrinted called")
	return s.adminrepo.AdminGetTotalSheetsPrinted(ctx)
}

func (s *AdminRepo) AdminGetTotalColorSheetsPrinted(ctx context.Context) (int, error) {
	s.logger.Info("AdminGetTotalColorSheetsPrinted called")
	return s.adminrepo.AdminGetTotalColorSheetsPrinted(ctx)
}

func (s *AdminRepo) AdminGetTotalBlackAndWhiteSheetsPrinted(ctx context.Context) (int, error) {
	s.logger.Info("AdminGetTotalBlackAndWhiteSheetsPrinted called")
	return s.adminrepo.AdminGetTotalBlackAndWhiteSheetsPrinted(ctx)
}

func (s *AdminRepo) AdminGetRevenueLast24Hours(ctx context.Context) (float64, error) {
	s.logger.Info("AdminGetRevenueLast24Hours called")
	return s.adminrepo.AdminGetRevenueLast24Hours(ctx)
}

func (s *AdminRepo) AdminGetSheetsPrintedLast24Hours(ctx context.Context) (int, error) {
	s.logger.Info("AdminGetSheetsPrintedLast24Hours called")
	return s.adminrepo.AdminGetSheetsPrintedLast24Hours(ctx)
}

func (s *AdminRepo) AdminGetColorSheetsPrintedLast24Hours(ctx context.Context) (int, error) {
	s.logger.Info("AdminGetColorSheetsPrintedLast24Hours called")
	return s.adminrepo.AdminGetColorSheetsPrintedLast24Hours(ctx)
}

func (s *AdminRepo) AdminGetBlackAndWhiteSheetsPrintedLast24Hours(ctx context.Context) (int, error) {
	s.logger.Info("AdminGetBlackAndWhiteSheetsPrintedLast24Hours called")
	return s.adminrepo.AdminGetBlackAndWhiteSheetsPrintedLast24Hours(ctx)
}