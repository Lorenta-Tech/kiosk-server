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
	logger       *slog.Logger
}

func NewAdminHandler(AdminService *service.AdminRepo, logger *slog.Logger) *AdminHandler {
	return &AdminHandler{
		AdminService: AdminService,
		logger:       logger,
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
		utils.HandleError(w, h.logger, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"history": history})
}

func (h *AdminHandler) HandleGetTotalRevenue(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	h.logger.Info("HandleGetTotalRevenue called")

	revenue, err := h.AdminService.AdminGetTotalRevenue(ctx)

	if err != nil {
		utils.HandleError(w, h.logger, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"total_revenue": revenue})
}

func (h *AdminHandler) HandleGetTotalSheetsPrinted(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	h.logger.Info("HandleGetTotalSheetsPrinted called")

	totalSheets, err := h.AdminService.AdminGetTotalSheetsPrinted(ctx)

	if err != nil {
		utils.HandleError(w, h.logger, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"total_sheets_printed": totalSheets})
}

func (h *AdminHandler) HandleGetTotalColorSheetsPrinted(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	h.logger.Info("HandleGetTotalColorSheetsPrinted called")

	totalColorSheets, err := h.AdminService.AdminGetTotalColorSheetsPrinted(ctx)

	if err != nil {
		utils.HandleError(w, h.logger, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"total_color_sheets_printed": totalColorSheets})
}

func (h *AdminHandler) HandleGetTotalBlackAndWhiteSheetsPrinted(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	h.logger.Info("HandleGetTotalBlackAndWhiteSheetsPrinted called")

	totalBWSheets, err := h.AdminService.AdminGetTotalBlackAndWhiteSheetsPrinted(ctx)

	if err != nil {
		utils.HandleError(w, h.logger, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"total_black_and_white_sheets_printed": totalBWSheets})
}

func (h *AdminHandler) HandleGetRevenueLast24Hours(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	h.logger.Info("HandleGetRevenueLast24Hours called")

	revenue, err := h.AdminService.AdminGetRevenueLast24Hours(ctx)

	if err != nil {
		utils.HandleError(w, h.logger, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"revenue_last_24_hours": revenue})
}

func (h *AdminHandler) HandleGetSheetsPrintedLast24Hours(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	h.logger.Info("HandleGetSheetsPrintedLast24Hours called")

	totalSheets, err := h.AdminService.AdminGetSheetsPrintedLast24Hours(ctx)

	if err != nil {
		utils.HandleError(w, h.logger, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"sheets_printed_last_24_hours": totalSheets})
}

func (h *AdminHandler) HandleGetColorSheetsPrintedLast24Hours(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	h.logger.Info("HandleGetColorSheetsPrintedLast24Hours called")

	totalColorSheets, err := h.AdminService.AdminGetColorSheetsPrintedLast24Hours(ctx)

	if err != nil {
		utils.HandleError(w, h.logger, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"color_sheets_printed_last_24_hours": totalColorSheets})
}

func (h *AdminHandler) HandleGetBlackAndWhiteSheetsPrintedLast24Hours(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	h.logger.Info("HandleGetBlackAndWhiteSheetsPrintedLast24Hours called")

	totalBWSheets, err := h.AdminService.AdminGetBlackAndWhiteSheetsPrintedLast24Hours(ctx)

	if err != nil {
		utils.HandleError(w, h.logger, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"black_and_white_sheets_printed_last_24_hours": totalBWSheets})
}	

