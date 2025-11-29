// postgre/service/user_service.go (updated to handle password field better)
package service

import (
	"context"
	"errors"
	"time"

	"BACKEND-UAS/pgmongo/jwt"
	"BACKEND-UAS/pgmongo/model"
	"BACKEND-UAS/pgmongo/repository"

	"github.com/google/uuid"
)

type CreateUserReq struct {
	Username     string `json:"username"`
	Email        string `json:"email"`
	Password     string `json:"password"` // Plain password for input
	FullName     string `json:"full_name"`
	RoleID       string `json:"role_id"`
}

type UserService interface {
	GetAll(ctx context.Context) ([]*model.User, error)
	GetByID(ctx context.Context, id string) (*model.User, error)
	Create(ctx context.Context, userReq *CreateUserReq) (*model.User, error)
	Update(ctx context.Context, id string, userReq *model.User) (*model.User, error)
	Delete(ctx context.Context, id string) error
	UpdateUserRole(ctx context.Context, id, roleID string) (*model.User, error)
}

type userService struct {
	userRepo repository.UserRepository
	jwtSvc   jwt.JWTService
}

func NewUserService(r repository.UserRepository, j jwt.JWTService) UserService {
	return &userService{userRepo: r, jwtSvc: j}
}

func (s *userService) GetAll(ctx context.Context) ([]*model.User, error) {
	return s.userRepo.GetAll(ctx)
}

func (s *userService) GetByID(ctx context.Context, id string) (*model.User, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (s *userService) Create(ctx context.Context, userReq *CreateUserReq) (*model.User, error) {
	if userReq.Username == "" || userReq.Email == "" || userReq.Password == "" || userReq.FullName == "" || userReq.RoleID == "" {
		return nil, errors.New("missing required fields")
	}

	// Hash plain password
	hash, err := s.jwtSvc.HashPassword(userReq.Password)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	user := &model.User{
		ID:           uuid.New().String(),
		Username:     userReq.Username,
		Email:        userReq.Email,
		PasswordHash: hash,
		FullName:     userReq.FullName,
		RoleID:       userReq.RoleID,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err = s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, err
	}
	fullUser, err := s.userRepo.FindByID(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	return fullUser, nil
}

func (s *userService) Update(ctx context.Context, id string, userReq *model.User) (*model.User, error) {
	existing, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Optional: re-hash if password provided
	if userReq.PasswordHash != "" {
		hash, err := s.jwtSvc.HashPassword(userReq.PasswordHash)
		if err != nil {
			return nil, errors.New("failed to hash password")
		}
		userReq.PasswordHash = hash
	} else {
		userReq.PasswordHash = existing.PasswordHash
	}

	userReq.UpdatedAt = time.Now()
	err = s.userRepo.Update(ctx, id, userReq)
	if err != nil {
		return nil, err
	}
	fullUser, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return fullUser, nil
}

func (s *userService) Delete(ctx context.Context, id string) error {
	_, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return errors.New("user not found")
	}
	err = s.userRepo.Delete(ctx, id)
	if err != nil {
		return err
	}
	return nil
}

func (s *userService) UpdateUserRole(ctx context.Context, id, roleID string) (*model.User, error) {
	_, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("user not found")
	}
	err = s.userRepo.UpdateRole(ctx, id, roleID)
	if err != nil {
		return nil, err
	}
	fullUser, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return fullUser, nil
}