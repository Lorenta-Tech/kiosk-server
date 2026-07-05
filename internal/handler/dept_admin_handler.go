package handler

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/Lorenta-Tech/kiosk-server/internal/models"
	"github.com/Lorenta-Tech/kiosk-server/internal/service"
	"github.com/Lorenta-Tech/kiosk-server/internal/validator"
	"github.com/Lorenta-Tech/kiosk-server/pkg/apperror"
	"github.com/Lorenta-Tech/kiosk-server/pkg/utils"
)

type DeptAdminHandler struct {
	deptAdminService *service.DeptAdminService
	logger           *slog.Logger
}

func NewDeptAdminHandler(
	deptAdminService *service.DeptAdminService,
	logger *slog.Logger,
) *DeptAdminHandler {
	return &DeptAdminHandler{
		deptAdminService: deptAdminService,
		logger:           logger,
	}
}

// ================================================================
// Super Admin Login
// ================================================================

func (h *DeptAdminHandler) HandleSuperAdminLogin(
	w http.ResponseWriter,
	r *http.Request,
) {

	ctx, cancel := context.WithTimeout(
		r.Context(),
		10*time.Second,
	)
	defer cancel()

	req, err := utils.DecodeJSON[models.SuperAdminLoginRequest](r)
	if err != nil {
		utils.HandleError(w, h.logger, err)
		return
	}

	if err := validator.Validate(req); err != nil {
		utils.HandleError(
			w,
			h.logger,
			apperror.BadRequest(
				"validation_error",
				err.Error(),
			),
		)
		return
	}

	resp, err := h.deptAdminService.SuperAdminLogin(
		ctx,
		req,
	)
	if err != nil {
		utils.HandleError(w, h.logger, err)
		return
	}

	utils.WriteJSON(
		w,
		http.StatusOK,
		utils.Envelope{
			"data": resp,
		},
	)
}

// ================================================================
// Dept Admin Login
// ================================================================

func (h *DeptAdminHandler) HandleDeptAdminLogin(
	w http.ResponseWriter,
	r *http.Request,
) {

	ctx, cancel := context.WithTimeout(
		r.Context(),
		10*time.Second,
	)
	defer cancel()

	req, err := utils.DecodeJSON[models.DeptAdminLoginRequest](r)
	if err != nil {
		utils.HandleError(w, h.logger, err)
		return
	}

	if err := validator.Validate(req); err != nil {
		utils.HandleError(
			w,
			h.logger,
			apperror.BadRequest(
				"validation_error",
				err.Error(),
			),
		)
		return
	}

	resp, err := h.deptAdminService.DeptAdminLogin(
		ctx,
		req,
	)
	if err != nil {
		utils.HandleError(w, h.logger, err)
		return
	}

	utils.WriteJSON(
		w,
		http.StatusOK,
		utils.Envelope{
			"data": resp,
		},
	)
}

// ================================================================
// Register Dept Admin
// ================================================================

func (h *DeptAdminHandler) HandleRegisterDeptAdmin(
	w http.ResponseWriter,
	r *http.Request,
) {

	ctx, cancel := context.WithTimeout(
		r.Context(),
		10*time.Second,
	)
	defer cancel()

	req, err := utils.DecodeJSON[models.RegisterDeptAdminRequest](r)
	if err != nil {
		utils.HandleError(w, h.logger, err)
		return
	}

	if err := validator.Validate(req); err != nil {
		utils.HandleError(
			w,
			h.logger,
			apperror.BadRequest(
				"validation_error",
				err.Error(),
			),
		)
		return
	}

	resp, err := h.deptAdminService.RegisterDeptAdmin(
		ctx,
		req,
	)
	if err != nil {
		utils.HandleError(w, h.logger, err)
		return
	}

	utils.WriteJSON(
		w,
		http.StatusCreated,
		utils.Envelope{
			"data": resp,
		},
	)
}

// ================================================================
// List Dept Admins
// ================================================================

func (h *DeptAdminHandler) HandleListDeptAdmins(
	w http.ResponseWriter,
	r *http.Request,
) {

	ctx, cancel := context.WithTimeout(
		r.Context(),
		10*time.Second,
	)
	defer cancel()

	resp, err := h.deptAdminService.ListDeptAdmins(ctx)
	if err != nil {
		utils.HandleError(w, h.logger, err)
		return
	}

	utils.WriteJSON(
		w,
		http.StatusOK,
		utils.Envelope{
			"data": resp,
		},
	)
}
