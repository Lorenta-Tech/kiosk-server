package service

import (
	"context"
	"log/slog"

	"github.com/Lorenta-Tech/kiosk-server/internal/models"
	"github.com/Lorenta-Tech/kiosk-server/internal/repository"
)

type AdminService struct {
	filerepo  repository.FileRepo
	adminrepo repository.AdminRepo
	logger    *slog.Logger
}

func NewAdminService(filerepo repository.FileRepo, adminrepo repository.AdminRepo, logger *slog.Logger) *AdminService {
	return &AdminService{
		filerepo:  filerepo,
		adminrepo: adminrepo,
		logger:    logger,
	}
}

func (s *AdminService) AdminFetchPrintHistory(ctx context.Context) ([]models.PrintJob, error) {
	// Implement the logic to fetch print history
	// This is a placeholder for the actual implementation

	s.logger.Info("FetchPrintHistory called")
	return s.adminrepo.AdminFetchPrintHistory(ctx)
}

func (s *AdminService) AdminGetTotalRevenue(ctx context.Context) (float64, error) {
	s.logger.Info("AdminGetTotalRevenue called")
	return s.adminrepo.AdminGetTotalRevenue(ctx)
}

func (s *AdminService) AdminGetTotalSheetsPrinted(ctx context.Context) (int, error) {
	s.logger.Info("AdminGetTotalSheetsPrinted called")
	return s.adminrepo.AdminGetTotalSheetsPrinted(ctx)
}

func (s *AdminService) AdminGetTotalColorSheetsPrinted(ctx context.Context) (int, error) {
	s.logger.Info("AdminGetTotalColorSheetsPrinted called")
	return s.adminrepo.AdminGetTotalColorSheetsPrinted(ctx)
}

func (s *AdminService) AdminGetTotalBlackAndWhiteSheetsPrinted(ctx context.Context) (int, error) {
	s.logger.Info("AdminGetTotalBlackAndWhiteSheetsPrinted called")
	return s.adminrepo.AdminGetTotalBlackAndWhiteSheetsPrinted(ctx)
}

func (s *AdminService) AdminGetRevenueLast24Hours(ctx context.Context) (float64, error) {
	s.logger.Info("AdminGetRevenueLast24Hours called")
	return s.adminrepo.AdminGetRevenueLast24Hours(ctx)
}

func (s *AdminService) AdminGetSheetsPrintedLast24Hours(ctx context.Context) (int, error) {
	s.logger.Info("AdminGetSheetsPrintedLast24Hours called")
	return s.adminrepo.AdminGetSheetsPrintedLast24Hours(ctx)
}

func (s *AdminService) AdminGetColorSheetsPrintedLast24Hours(ctx context.Context) (int, error) {
	s.logger.Info("AdminGetColorSheetsPrintedLast24Hours called")
	return s.adminrepo.AdminGetColorSheetsPrintedLast24Hours(ctx)
}

func (s *AdminService) AdminGetBlackAndWhiteSheetsPrintedLast24Hours(ctx context.Context) (int, error) {
	s.logger.Info("AdminGetBlackAndWhiteSheetsPrintedLast24Hours called")
	return s.adminrepo.AdminGetBlackAndWhiteSheetsPrintedLast24Hours(ctx)
}

func (s *AdminService) AdminGetTotalSheetsPrintedInLast24HoursByDouble_Sided_Prints(ctx context.Context) (int, error) {
	s.logger.Info("AdminGetTotalSheetsPrintedInLast24HoursByDouble_Sided_Prints called")
	return s.adminrepo.AdminGetTotalSheetsPrintedInLast24HoursByDouble_Sided_Prints(ctx)
}

func (s *AdminService) AdminGetTotalSheetsPrintedInLast24HoursBySingle_Sided_Prints(ctx context.Context) (int, error) {
	s.logger.Info("AdminGetTotalSheetsPrintedInLast24HoursBySingle_Sided_Prints called")
	return s.adminrepo.AdminGetTotalSheetsPrintedInLast24HoursBySingle_Sided_Prints(ctx)
}

func (s *AdminService) AdminGetLast24HoursRevenueFromDouble_Sided_Prints(ctx context.Context) (int, error) {
	s.logger.Info("AdminGetLast24HoursRevenueFromDouble_Sided_Prints called")
	return s.adminrepo.AdminGetLast24HoursRevenueFromDouble_Sided_Prints(ctx)
}
func (s *AdminService) AdminGetLast24HoursRevenueFromSingle_Sided_Prints(ctx context.Context) (int, error) {
	s.logger.Info("AdminGetLast24HoursRevenueFromSingle_Sided_Prints called")
	return s.adminrepo.AdminGetLast24HoursRevenueFromSingle_Sided_Prints(ctx)
}

func (s *AdminService) AdminGetRevenueFromDouble_Sided_Prints(ctx context.Context) (int, error) {
	s.logger.Info("AdminGetRevenueFromDouble_Sided_Prints called")
	return s.adminrepo.AdminGetRevenueFromDouble_Sided_Prints(ctx)
}

func (s *AdminService) AdminGetRevenueFromSingle_Sided_Prints(ctx context.Context) (int, error) {
	s.logger.Info("AdminGetRevenueFromSingle_Sided_Prints called")
	return s.adminrepo.AdminGetRevenueFromSingle_Sided_Prints(ctx)
}

func (s *AdminService) AdminGetTotalSheetsPrintedByDouble_Sided_Prints(ctx context.Context) (int, error) {
	s.logger.Info("AdminGetTotalSheetsPrintedByDouble_Sided_Prints called")
	return s.adminrepo.AdminGetTotalSheetsPrintedByDouble_Sided_Prints(ctx)
}

func (s *AdminService) AdminGetTotalSheetsPrintedBySingle_Sided_Prints(ctx context.Context) (int, error) {
	s.logger.Info("AdminGetTotalSheetsPrintedBySingle_Sided_Prints called")
	return s.adminrepo.AdminGetTotalSheetsPrintedBySingle_Sided_Prints(ctx)
}

func (s *AdminService) AdminFetchPrintHistoryFor24H(ctx context.Context)([]models.PrintJob,error){
	s.logger.Info("AdminFetchPrintHistoryfor-24H")
	return s.adminrepo.AdminFetchPrintHistoryfor24H(ctx)
}

func (s *AdminService) AdminFetchPrintJobsOnlyPaidInLast24H(ctx context.Context)([]models.PrintJob,error){
	s.logger.Info("AdminFetchPrintJobOnlyPaidInLast24H")
	return s.adminrepo.AdminFetchPrintJobsOnlyPaidInLast24H(ctx)
}