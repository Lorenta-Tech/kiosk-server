package handler

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/Lorenta-Tech/kiosk-server/internal/service"
	"github.com/Lorenta-Tech/kiosk-server/pkg/utils"
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

func (h *AdminHandler) HandleFetchPrintHistory(w http.ResponseWriter, r *http.Request) {
	// Implement the logic to handle fetching print history
	// This is a placeholder for the actual implementation
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	h.logger.Info("HandleFetchPrintHistory called")

	history, err := h.AdminService.AdminFetchPrintHistory(ctx)

	if err != nil {
		utils.HandleError(w,h.logger,err)
		return
	}

	utils.WriteJSON(w, http.StatusOK,utils.Envelope{"history": history})
}