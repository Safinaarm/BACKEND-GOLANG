// postgre/service/auth_service.go
package service

import (
	"context"
	"errors"

	"BACKEND-UAS/postgre/jwt"
	"BACKEND-UAS/postgre/model"
	"BACKEND-UAS/postgre/repository"
)

type AuthService interface {
	Login(ctx context.Context, identifier, password string) (string, string, *model.User, string, []string, error)
}

type authService struct {
	userRepo repository.UserRepository
	jwtSvc   jwt.JWTService
}

func NewAuthService(r repository.UserRepository, j jwt.JWTService) AuthService {
	return &authService{userRepo: r, jwtSvc: j}
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
	if !jwt.CheckPasswordHash(password, user.PasswordHash) {
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
	// simple refresh token: same as access for demo, in prod use separate
	refreshToken, err := s.jwtSvc.GenerateToken(user.ID, user.RoleID, roleName)
	if err != nil {
		return "", "", nil, "", nil, errors.New("failed to generate refresh token")
	}
	return accessToken, refreshToken, user, roleName, perms, nil
}