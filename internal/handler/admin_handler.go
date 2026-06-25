package handler

import (
	"log/slog"

	"github.com/Lorenta-Tech/kiosk-server/internal/service"
)


type AdminHandler struct {
	AdminService *service.AdminRepo
	logger	   *slog.Logger
}

func NewAdminHandler(AdminService *service.AdminRepo, logger *slog.Logger) *AdminHandler {
	return &AdminHandler{
		AdminService: AdminService,
		logger:         logger,
	}
}

func (h *AdminHandler) HandleFetchPrintHistory() {
	// Implement the logic to handle fetching print history
	// This is a placeholder for the actual implementation
	h.logger.Info("HandleFetchPrintHistory called")
}