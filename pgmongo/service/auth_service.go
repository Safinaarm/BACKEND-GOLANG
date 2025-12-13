// postgre/service/auth_service.go
package service

import (
	"errors"
	"net/http"
	"context"

	"BACKEND-UAS/pgmongo/jwt"
	"BACKEND-UAS/pgmongo/model"
	"BACKEND-UAS/pgmongo/repository"

	"github.com/gofiber/fiber/v2"
)

// AuthService defines the interface for authentication operations
type AuthService interface {
	Login(ctx context.Context, identifier, password string) (string, string, *model.User, string, []string, error)
	Refresh(ctx context.Context, refreshToken string) (string, string, *model.User, string, []string, error)
	Logout(ctx context.Context, refreshToken string) error

	// Handlers
	LoginHandler(c *fiber.Ctx) error
	RefreshHandler(c *fiber.Ctx) error
	LogoutHandler(c *fiber.Ctx) error
	ProfileHandler(c *fiber.Ctx) error
}

type authService struct {
	userRepo repository.UserRepository
	jwtSvc   jwt.JWTService
}

func NewAuthService(r repository.UserRepository, j jwt.JWTService) AuthService {
	return &authService{userRepo: r, jwtSvc: j}
}

// @Summary Login user
// @Description Melakukan login dengan username/email dan password, menghasilkan access dan refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Param login body LoginRequest true "Credentials (username or email and password)"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} model.ErrorResponse "Invalid credentials or missing fields"
// @Failure 401 {object} model.ErrorResponse "Account inactive"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /api/v1/auth/login [post]
func (s *authService) LoginHandler(c *fiber.Ctx) error {
	var req struct {
		Identifier string `json:"username"` // allow username or email
		Password   string `json:"password"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "invalid body"})
	}
	if req.Identifier == "" || req.Password == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "identifier and password required"})
	}
	access, refresh, user, role, perms, err := s.Login(c.Context(), req.Identifier, req.Password)
	if err != nil {
		if err.Error() == "user not found" || err.Error() == "invalid credentials" {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"status": "error", "message": "invalid credentials"})
		}
		if err.Error() == "account inactive" {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"status": "error", "message": "account inactive"})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": err.Error()})
	}
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status": "success",
		"data": fiber.Map{
			"token":       access,
			"refreshToken": refresh,
			"user": fiber.Map{
				"id":        user.ID,
				"username":  user.Username,
				"fullName":  user.FullName,
				"role_id":   user.RoleID,
				"role":      role,
				"is_active": user.IsActive,
			},
			"permissions": perms,
		},
	})
}

// @Summary Refresh token
// @Description Memperbarui access dan refresh token menggunakan refresh token yang valid
// @Tags Auth
// @Accept json
// @Produce json
// @Param refresh_token body RefreshRequest true "Refresh token"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} model.ErrorResponse "Invalid refresh token"
// @Failure 401 {object} model.ErrorResponse "Account inactive"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /api/v1/auth/refresh [post]
func (s *authService) RefreshHandler(c *fiber.Ctx) error {
	var req struct {
		RefreshToken string `json:"refreshToken"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "invalid body"})
	}
	if req.RefreshToken == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "refreshToken required"})
	}
	access, refresh, user, role, perms, err := s.Refresh(c.Context(), req.RefreshToken)
	if err != nil {
		if err.Error() == "invalid refresh token" {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"status": "error", "message": "invalid refresh token"})
		}
		if err.Error() == "account inactive" {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"status": "error", "message": "account inactive"})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": err.Error()})
	}
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status": "success",
		"data": fiber.Map{
			"token":       access,
			"refreshToken": refresh,
			"user": fiber.Map{
				"id":        user.ID,
				"username":  user.Username,
				"fullName":  user.FullName,
				"role_id":   user.RoleID,
				"role":      role,
				"is_active": user.IsActive,
			},
			"permissions": perms,
		},
	})
}

// @Summary Logout user
// @Description Melakukan logout dengan invalidasi refresh token (client-side untuk demo)
// @Tags Auth
// @Accept json
// @Produce json
// @Param refresh_token body RefreshRequest true "Refresh token to invalidate"
// @Success 200 {object} model.SuccessResponse
// @Failure 400 {object} model.ErrorResponse "Invalid refresh token"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /api/v1/auth/logout [post]
func (s *authService) LogoutHandler(c *fiber.Ctx) error {
	var req struct {
		RefreshToken string `json:"refreshToken"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "invalid body"})
	}
	// RefreshToken is optional for demo; client should discard tokens client-side
	err := s.Logout(c.Context(), req.RefreshToken)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": err.Error()})
	}
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "logged out successfully",
	})
}

// @Summary Get profile
// @Description Mengambil data user berdasarkan JWT token
// @Tags Auth
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} ProfileResponse
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Router /api/v1/auth/profile [get]
func (s *authService) ProfileHandler(c *fiber.Ctx) error {
	userID := c.Locals("userId").(string)
	role := c.Locals("role").(string)
	perms := c.Locals("permissions").([]string)
	return c.JSON(fiber.Map{
		"status": "success",
		"data": fiber.Map{
			"userId": userID,
			"role":   role,
			"permissions": perms,
		},
	})
}

func (s *authService) Login(ctx context.Context, identifier, password string) (string, string, *model.User, string, []string, error) {
	user, err := s.userRepo.FindByUsernameOrEmail(ctx, identifier)
	if err != nil {
		return "", "", nil, "", nil, errors.New("user not found")
	}
	if !user.IsActive {
		return "", "", nil, "", nil, errors.New("account inactive")
	}
	// check password
	if !s.jwtSvc.CheckPasswordHash(password, user.PasswordHash) {
		return "", "", nil, "", nil, errors.New("invalid credentials")
	}
	// get role name and permissions
	roleName, err := s.userRepo.GetRoleNameByID(ctx, user.RoleID)
	if err != nil {
		return "", "", nil, "", nil, errors.New("failed to fetch role")
	}
	perms, err := s.userRepo.GetPermissionsByRoleID(ctx, user.RoleID)
	if err != nil {
		return "", "", nil, "", nil, errors.New("failed to fetch permissions")
	}
	// generate access token
	accessToken, err := s.jwtSvc.GenerateToken(user.ID, user.RoleID, roleName)
	if err != nil {
		return "", "", nil, "", nil, errors.New("failed to generate token")
	}
	// simple refresh token: same as access for demo, in prod use separate longer exp
	refreshToken, err := s.jwtSvc.GenerateToken(user.ID, user.RoleID, roleName)
	if err != nil {
		return "", "", nil, "", nil, errors.New("failed to generate refresh token")
	}
	return accessToken, refreshToken, user, roleName, perms, nil
}

func (s *authService) Refresh(ctx context.Context, refreshToken string) (string, string, *model.User, string, []string, error) {
	// Validate the refresh token and extract claims
	token, err := s.jwtSvc.ValidateToken(refreshToken)
	if err != nil {
		return "", "", nil, "", nil, errors.New("invalid refresh token")
	}

	claims, ok := token.Claims.(*jwt.Claims)
	if !ok {
		return "", "", nil, "", nil, errors.New("invalid claims format")
	}

	userID := claims.UserID

	// Fetch user by ID from claims
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return "", "", nil, "", nil, errors.New("user not found")
	}
	if !user.IsActive {
		return "", "", nil, "", nil, errors.New("account inactive")
	}

	// Get role name and permissions
	roleName, err := s.userRepo.GetRoleNameByID(ctx, user.RoleID)
	if err != nil {
		return "", "", nil, "", nil, errors.New("failed to fetch role")
	}
	perms, err := s.userRepo.GetPermissionsByRoleID(ctx, user.RoleID)
	if err != nil {
		return "", "", nil, "", nil, errors.New("failed to fetch permissions")
	}

	// Generate new access token
	accessToken, err := s.jwtSvc.GenerateToken(user.ID, user.RoleID, roleName)
	if err != nil {
		return "", "", nil, "", nil, errors.New("failed to generate access token")
	}

	// Generate new refresh token (rotation for better security, even in demo)
	refreshTokenNew, err := s.jwtSvc.GenerateToken(user.ID, user.RoleID, roleName)
	if err != nil {
		return "", "", nil, "", nil, errors.New("failed to generate refresh token")
	}

	return accessToken, refreshTokenNew, user, roleName, perms, nil
}

func (s *authService) Logout(ctx context.Context, refreshToken string) error {
	// For demo with stateless JWT, logout is client-side.
	// In production, you could blacklist the refresh token in a DB or Redis.
	// Here, we just return success without invalidation.
	return nil
}