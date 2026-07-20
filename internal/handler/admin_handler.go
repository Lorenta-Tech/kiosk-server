package handler

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/Lorenta-Tech/kiosk-server/internal/service"
	"github.com/Lorenta-Tech/kiosk-server/pkg/apperror"
	"github.com/Lorenta-Tech/kiosk-server/pkg/utils"
)

func reportDate(r *http.Request) (string, error) {
	date, err := utils.ReadParam(r, "date")
	if err != nil {
		return "", apperror.BadRequest("invalid_date", "date must be in YYYY-MM-DD format")
	}
	if _, err := time.Parse(time.DateOnly, date); err != nil {
		return "", apperror.BadRequest("invalid_date", "date must be in YYYY-MM-DD format")
	}
	return date, nil
}

type AdminHandler struct {
	AdminService *service.AdminService
	logger       *slog.Logger
}

func NewAdminHandler(AdminService *service.AdminService, logger *slog.Logger) *AdminHandler {
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

func (h *AdminHandler) HandleGetTotalSheetsPrintedInLast24HoursByDouble_Sided_Prints(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	h.logger.Info("HandleGetTotalSheetsPrintedInLast24HoursByDouble_Sided_Prints called")
	totalSheets, err := h.AdminService.AdminGetTotalSheetsPrintedInLast24HoursByDouble_Sided_Prints(ctx)

	if err != nil {
		utils.HandleError(w, h.logger, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"total_sheets_printed_last_24_hours_by_double_sided_prints": totalSheets})
}

func (h *AdminHandler) HandleGetTotalSheetsPrintedInLast24HoursBySingle_Sided_Prints(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	h.logger.Info("HandleGetTotalSheetsPrintedInLast24HoursBySingle_Sided_Prints called")
	totalSheets, err := h.AdminService.AdminGetTotalSheetsPrintedInLast24HoursBySingle_Sided_Prints(ctx)

	if err != nil {
		utils.HandleError(w, h.logger, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"total_sheets_printed_last_24_hours_by_single_sided_prints": totalSheets})
}

func (h *AdminHandler) HandleGetLast24HoursRevenueFromDouble_Sided_Prints(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	h.logger.Info("HandleGetLast24HoursRevenueFromDouble_Sided_Prints called")
	revenue, err := h.AdminService.AdminGetLast24HoursRevenueFromDouble_Sided_Prints(ctx)

	if err != nil {
		utils.HandleError(w, h.logger, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"last_24_hours_revenue_from_double_sided_prints": revenue})
}

func (h *AdminHandler) HandleGetLast24HoursRevenueFromSingle_Sided_Prints(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	h.logger.Info("HandleGetLast24HoursRevenueFromSingle_Sided_Prints called")
	revenue, err := h.AdminService.AdminGetLast24HoursRevenueFromSingle_Sided_Prints(ctx)

	if err != nil {
		utils.HandleError(w, h.logger, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"last_24_hours_revenue_from_single_sided_prints": revenue})
}

func (h *AdminHandler) HandleGetRevenueFromDouble_Sided_Prints(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	h.logger.Info("HandleGetRevenueFromDouble_Sided_Prints called")
	revenue, err := h.AdminService.AdminGetRevenueFromDouble_Sided_Prints(ctx)

	if err != nil {
		utils.HandleError(w, h.logger, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"revenue_from_double_sided_prints": revenue})
}

func (h *AdminHandler) HandleGetRevenueFromSingle_Sided_Prints(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	h.logger.Info("HandleGetRevenueFromSingle_Sided_Prints called")
	revenue, err := h.AdminService.AdminGetRevenueFromSingle_Sided_Prints(ctx)

	if err != nil {
		utils.HandleError(w, h.logger, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"revenue_from_single_sided_prints": revenue})
}

func (h *AdminHandler) HandleGetDoubleSidePrintsCount(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	h.logger.Info("HandleGetDoubleSidePrintsCount called")
	totalSheets, err := h.AdminService.AdminGetTotalSheetsPrintedByDouble_Sided_Prints(ctx)

	if err != nil {
		utils.HandleError(w, h.logger, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"total_sheets_printed_by_double_sided_prints": totalSheets})
}

func (h *AdminHandler) HandleGetSingleSidePrintsCount(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	h.logger.Info("HandleGetSingleSidePrintsCount called")
	totalSheets, err := h.AdminService.AdminGetTotalSheetsPrintedBySingle_Sided_Prints(ctx)

	if err != nil {
		utils.HandleError(w, h.logger, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"total_sheets_printed_by_single_sided_prints": totalSheets})
}

func (h *AdminHandler) HandleFetchPrintHistoryFor24H(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	h.logger.Info("HandleFetchPrintHistoryFor-24H")

	history, err := h.AdminService.AdminFetchPrintHistoryFor24H(ctx)

	if err != nil {
		utils.HandleError(w, h.logger, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"print_history": history})
}

func (h *AdminHandler) HandleFetchPrintJobOnlyPaidInLast24H(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	Job, err := h.AdminService.AdminFetchPrintJobsOnlyPaidInLast24H(ctx)

	if err != nil {
		utils.HandleError(w, h.logger, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"jobs": Job})
}

func (h *AdminHandler) HandleGetTotalSessionCountOfToday(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	total, err := h.AdminService.AdminGetTotalSessionCountOfToday(ctx)
	if err != nil {
		utils.HandleError(w, h.logger, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"total_sessions_today": total})
}

func (h *AdminHandler) HandleGetTodaysPaymentHistory(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	payments, err := h.AdminService.AdminGetTodaysPaymentHistory(ctx)
	if err != nil {
		utils.HandleError(w, h.logger, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"payments": payments})
}

func (h *AdminHandler) HandleGetPrintJobsForPricedStatus(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	jobs, err := h.AdminService.AdminGetPrintJobsForPricedStatus(ctx)
	if err != nil {
		utils.HandleError(w, h.logger, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"jobs": jobs})
}

func (h *AdminHandler) HandleGetRevenueForDate(w http.ResponseWriter, r *http.Request) {
	date, err := reportDate(r)
	if err != nil {
		utils.HandleError(w, h.logger, err)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	revenue, err := h.AdminService.AdminGetRevenueForDate(ctx, date)
	if err != nil {
		utils.HandleError(w, h.logger, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"date": date, "total_revenue": revenue})
}

func (h *AdminHandler) HandleGetTotalSheetsPrintedForDate(w http.ResponseWriter, r *http.Request) {
	date, err := reportDate(r)
	if err != nil {
		utils.HandleError(w, h.logger, err)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	totalSheets, err := h.AdminService.AdminGetTotalSheetsPrintedForDate(ctx, date)
	if err != nil {
		utils.HandleError(w, h.logger, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"date": date, "total_sheets_printed": totalSheets})
}

func (h *AdminHandler) HandleFetchPrintHistoryForDate(w http.ResponseWriter, r *http.Request) {
	date, err := reportDate(r)
	if err != nil {
		utils.HandleError(w, h.logger, err)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	history, err := h.AdminService.AdminFetchPrintHistoryForDate(ctx, date)
	if err != nil {
		utils.HandleError(w, h.logger, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"date": date, "history": history})
}
