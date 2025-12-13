// postgre/service/auth_service.go
package service

import (
	"context"
	"errors"

	"BACKEND-UAS/pgmongo/jwt"
	"BACKEND-UAS/pgmongo/model"
	"BACKEND-UAS/pgmongo/repository"
)

// AuthService defines the interface for authentication operations
type AuthService interface {
	Login(ctx context.Context, identifier, password string) (string, string, *model.User, string, []string, error)
	Refresh(ctx context.Context, refreshToken string) (string, string, *model.User, string, []string, error)
	Logout(ctx context.Context, refreshToken string) error
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
// @Router /login [post]
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
// @Router /refresh [post]
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

// @Summary Logout user
// @Description Melakukan logout dengan invalidasi refresh token (client-side untuk demo)
// @Tags Auth
// @Accept json
// @Produce json
// @Param refresh_token body RefreshRequest true "Refresh token to invalidate"
// @Success 200 {object} model.SuccessResponse
// @Failure 400 {object} model.ErrorResponse "Invalid refresh token"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /logout [post]
func (s *authService) Logout(ctx context.Context, refreshToken string) error {
	// For demo with stateless JWT, logout is client-side.
	// In production, you could blacklist the refresh token in a DB or Redis.
	// Here, we just return success without invalidation.
	return nil
}