package service

import (
	"context"
	"log/slog"

	"github.com/Lorenta-Tech/kiosk-server/internal/models"
	"github.com/Lorenta-Tech/kiosk-server/internal/repository"
	"github.com/Lorenta-Tech/kiosk-server/pkg/apperror"
	appjwt "github.com/Lorenta-Tech/kiosk-server/pkg/jwt"
	"github.com/Lorenta-Tech/kiosk-server/pkg/password"
	"github.com/google/uuid"
)

type DeptAdminService struct {
	adminRepo          repository.AdminRepo
	logger             *slog.Logger
	jwtSecret          string
	superAdminEmail    string
	superAdminPassword string
}

func NewDeptAdminService(
	adminRepo repository.AdminRepo,
	logger *slog.Logger,
	jwtSecret string,
	superAdminEmail string,
	superAdminPassword string,
) *DeptAdminService {
	return &DeptAdminService{
		adminRepo:          adminRepo,
		logger:             logger,
		jwtSecret:          jwtSecret,
		superAdminEmail:    superAdminEmail,
		superAdminPassword: superAdminPassword,
	}
}

// ================================================================
// Super Admin Login
// ================================================================

func (s *DeptAdminService) SuperAdminLogin(
	ctx context.Context,
	req models.SuperAdminLoginRequest,
) (models.SuperAdminLoginResponse, error) {

	s.logger.Info("super admin login started")

	if req.Email != s.superAdminEmail {
		return models.SuperAdminLoginResponse{},
			apperror.Unauthorized("invalid email or password")
	}

	if req.Password != s.superAdminPassword {
		return models.SuperAdminLoginResponse{},
			apperror.Unauthorized("invalid email or password")
	}

	token, err := appjwt.GenerateSuperAdminToken(
		s.jwtSecret,
		req.Email,
	)
	if err != nil {
		return models.SuperAdminLoginResponse{},
			apperror.Internal("failed to generate auth token", err)
	}

	s.logger.Info("super admin authenticated", "email", req.Email)

	return models.SuperAdminLoginResponse{
		Token: token,
	}, nil
}

// ================================================================
// Register Dept Admin
// ================================================================

func (s *DeptAdminService) RegisterDeptAdmin(
	ctx context.Context,
	req models.RegisterDeptAdminRequest,
) (models.RegisterDeptAdminResponse, error) {

	s.logger.Info(
		"department admin registration started",
		"email", req.Email,
		"branch_id", req.BranchID,
	)

	_, err := s.adminRepo.GetDeptAdminByEmail(
		ctx,
		req.Email,
	)

	if err == nil {
		return models.RegisterDeptAdminResponse{},
			apperror.Conflict(
				"dept_admin_exists",
				"department admin already exists",
			)
	}

	hashedPassword, err := password.Hash(req.Password)
	if err != nil {
		return models.RegisterDeptAdminResponse{},
			apperror.Internal(
				"failed to hash password",
				err,
			)
	}

	admin := models.DeptAdmin{
		ID:           uuid.NewString(),
		Name:         req.Name,
		Email:        req.Email,
		BranchID:     req.BranchID,
		PasswordHash: hashedPassword,
		Role:         "dept_admin",
	}

	if err := s.adminRepo.CreateDeptAdmin(
		ctx,
		admin,
	); err != nil {
		return models.RegisterDeptAdminResponse{}, err
	}

	s.logger.Info(
		"department admin created",
		"admin_id", admin.ID,
		"email", admin.Email,
	)

	return models.RegisterDeptAdminResponse{
		ID:       admin.ID,
		Name:     admin.Name,
		Email:    admin.Email,
		BranchID: admin.BranchID,
	}, nil
}

// ================================================================
// Dept Admin Login
// ================================================================

func (s *DeptAdminService) DeptAdminLogin(
	ctx context.Context,
	req models.DeptAdminLoginRequest,
) (models.DeptAdminLoginResponse, error) {

	s.logger.Info(
		"department admin login started",
		"email", req.Email,
	)

	admin, err := s.adminRepo.GetDeptAdminByEmail(
		ctx,
		req.Email,
	)
	if err != nil {
		return models.DeptAdminLoginResponse{},
			apperror.Unauthorized(
				"invalid email or password",
			)
	}

	if err := password.Verify(
		admin.PasswordHash,
		req.Password,
	); err != nil {

		return models.DeptAdminLoginResponse{},
			apperror.Unauthorized(
				"invalid email or password",
			)
	}

	token, err := appjwt.GenerateDeptAdminToken(
		s.jwtSecret,
		admin.ID,
		admin.Email,
		admin.BranchID,
	)
	if err != nil {
		return models.DeptAdminLoginResponse{},
			apperror.Internal(
				"failed to generate auth token",
				err,
			)
	}

	s.logger.Info(
		"department admin authenticated",
		"admin_id", admin.ID,
		"email", admin.Email,
	)

	return models.DeptAdminLoginResponse{
		Token: token,
		Admin: models.DeptAdminProfile{
			ID:       admin.ID,
			Name:     admin.Name,
			Email:    admin.Email,
			BranchID: admin.BranchID,
		},
	}, nil
}

// ================================================================
// List Dept Admins
// ================================================================

func (s *DeptAdminService) ListDeptAdmins(
	ctx context.Context,
) ([]models.DeptAdminListItem, error) {

	admins, err := s.adminRepo.ListDeptAdmins(ctx)
	if err != nil {
		return nil, err
	}

	response := make(
		[]models.DeptAdminListItem,
		0,
		len(admins),
	)

	for _, admin := range admins {
		response = append(
			response,
			models.DeptAdminListItem{
				ID:        admin.ID,
				Name:      admin.Name,
				Email:     admin.Email,
				BranchID:  admin.BranchID,
				CreatedAt: admin.CreatedAt,
			},
		)
	}

	return response, nil
}
