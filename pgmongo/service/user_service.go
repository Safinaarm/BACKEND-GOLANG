// postgre/service/user_service.go
package service

import (
	"context"
	"errors"
	"net/http"
	"time"

	"BACKEND-UAS/pgmongo/jwt"
	"BACKEND-UAS/pgmongo/model"
	"BACKEND-UAS/pgmongo/repository"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type CreateUserReq struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"` // Plain password for input
	FullName string `json:"full_name"`
	RoleID   string `json:"role_id"`
}

type UserService interface {
	// Core methods
	GetAll(ctx context.Context) ([]*model.User, error)
	GetByID(ctx context.Context, id string) (*model.User, error)
	Create(ctx context.Context, req *CreateUserReq) (*model.User, error)
	Update(ctx context.Context, id string, req *model.User) (*model.User, error)
	Delete(ctx context.Context, id string) error
	UpdateUserRole(ctx context.Context, id, roleID string) (*model.User, error)

	// Handler methods (untuk route bersih)
	ListUsersHandler(c *fiber.Ctx) error
	GetUserHandler(c *fiber.Ctx) error
	CreateUserHandler(c *fiber.Ctx) error
	UpdateUserHandler(c *fiber.Ctx) error
	DeleteUserHandler(c *fiber.Ctx) error
	UpdateUserRoleHandler(c *fiber.Ctx) error
}

type userService struct {
	userRepo repository.UserRepository
	jwtSvc   jwt.JWTService
}

func NewUserService(r repository.UserRepository, j jwt.JWTService) UserService {
	return &userService{
		userRepo: r,
		jwtSvc:   j,
	}
}

// ==================== CORE LOGIC ====================

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

func (s *userService) Create(ctx context.Context, req *CreateUserReq) (*model.User, error) {
	if req.Username == "" || req.Email == "" || req.Password == "" || req.FullName == "" || req.RoleID == "" {
		return nil, errors.New("missing required fields")
	}

	hash, err := s.jwtSvc.HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	user := &model.User{
		ID:           uuid.New().String(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hash,
		FullName:     req.FullName,
		RoleID:       req.RoleID,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}
	return s.userRepo.FindByID(ctx, user.ID)
}

func (s *userService) Update(ctx context.Context, id string, req *model.User) (*model.User, error) {
	existing, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Jika password diisi (plain text), hash ulang
	if req.PasswordHash != "" && req.PasswordHash != existing.PasswordHash {
		hash, err := s.jwtSvc.HashPassword(req.PasswordHash)
		if err != nil {
			return nil, errors.New("failed to hash password")
		}
		req.PasswordHash = hash
	} else {
		req.PasswordHash = existing.PasswordHash
	}

	req.ID = id
	req.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, id, req); err != nil {
		return nil, err
	}
	return s.userRepo.FindByID(ctx, id)
}

func (s *userService) Delete(ctx context.Context, id string) error {
	if _, err := s.userRepo.FindByID(ctx, id); err != nil {
		return errors.New("user not found")
	}
	return s.userRepo.Delete(ctx, id)
}

func (s *userService) UpdateUserRole(ctx context.Context, id, roleID string) (*model.User, error) {
	if _, err := s.userRepo.FindByID(ctx, id); err != nil {
		return nil, errors.New("user not found")
	}
	if err := s.userRepo.UpdateRole(ctx, id, roleID); err != nil {
		return nil, err
	}
	return s.userRepo.FindByID(ctx, id)
}

// ==================== HANDLER METHODS (untuk route bersih) ====================

// @Summary Dapatkan semua user
// @Description Mengambil daftar semua user dari database (dengan pagination, search, sorting)
// @Tags Users
// @Accept json
// @Produce json
// @Param page query int false "Nomor halaman (default 1)"
// @Param limit query int false "Jumlah item per halaman (default 10)"
// @Param sortBy query string false "Kolom untuk sorting (id, username, email, role, created_at)"
// @Param order query string false "Urutan sorting (asc/desc)"
// @Param search query string false "Pencarian user"
// @Success 200 {object} model.UserResponse
// @Failure 500 {object} model.ErrorResponse
// @Security ApiKeyAuth
// @Router /api/v1/users [get]
func (s *userService) ListUsersHandler(c *fiber.Ctx) error {
	users, err := s.GetAll(c.Context())
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(model.ErrorResponse{Message: err.Error()})
	}
	for _, u := range users {
		u.PasswordHash = ""
	}
	// Nanti bisa diganti dengan UserResponse yang berisi pagination jika sudah diimplementasikan
	return c.JSON(fiber.Map{"data": users})
}

// @Summary Dapatkan user berdasarkan ID
// @Description Mengambil detail user dengan ID tertentu
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} model.User
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security ApiKeyAuth
// @Router /api/v1/users/{id} [get]
func (s *userService) GetUserHandler(c *fiber.Ctx) error {
	id := c.Params("id")
	user, err := s.GetByID(c.Context(), id)
	if err != nil {
		if err.Error() == "user not found" {
			return c.Status(http.StatusNotFound).JSON(model.ErrorResponse{Message: err.Error()})
		}
		return c.Status(http.StatusInternalServerError).JSON(model.ErrorResponse{Message: err.Error()})
	}
	user.PasswordHash = ""
	return c.JSON(user)
}

// @Summary Buat user baru
// @Description Membuat user baru dengan data yang diberikan
// @Tags Users
// @Accept json
// @Produce json
// @Param body body service.CreateUserReq true "Data user baru"
// @Success 201 {object} model.User
// @Failure 400 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security ApiKeyAuth
// @Router /api/v1/users [post]
func (s *userService) CreateUserHandler(c *fiber.Ctx) error {
	var req CreateUserReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(model.ErrorResponse{Message: "invalid body"})
	}

	user, err := s.Create(c.Context(), &req)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "missing required fields" {
			status = http.StatusBadRequest
		}
		return c.Status(status).JSON(model.ErrorResponse{Message: err.Error()})
	}
	user.PasswordHash = ""
	return c.Status(http.StatusCreated).JSON(user)
}

// @Summary Update user
// @Description Memperbarui data user dengan ID tertentu
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param body body model.User true "Data user yang diupdate"
// @Success 200 {object} model.User
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security ApiKeyAuth
// @Router /api/v1/users/{id} [put]
func (s *userService) UpdateUserHandler(c *fiber.Ctx) error {
	id := c.Params("id")
	var req model.User
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(model.ErrorResponse{Message: "invalid body"})
	}

	user, err := s.Update(c.Context(), id, &req)
	if err != nil {
		if err.Error() == "user not found" {
			return c.Status(http.StatusNotFound).JSON(model.ErrorResponse{Message: err.Error()})
		}
		return c.Status(http.StatusInternalServerError).JSON(model.ErrorResponse{Message: err.Error()})
	}
	user.PasswordHash = ""
	return c.JSON(user)
}

// @Summary Hapus user
// @Description Menghapus user dengan ID tertentu
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 204 "No Content"
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security ApiKeyAuth
// @Router /api/v1/users/{id} [delete]
func (s *userService) DeleteUserHandler(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := s.Delete(c.Context(), id); err != nil {
		if err.Error() == "user not found" {
			return c.Status(http.StatusNotFound).JSON(model.ErrorResponse{Message: err.Error()})
		}
		return c.Status(http.StatusInternalServerError).JSON(model.ErrorResponse{Message: err.Error()})
	}
	return c.SendStatus(http.StatusNoContent)
}

// @Summary Update role user
// @Description Memperbarui role user dengan ID tertentu
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param role_id query string true "Role ID baru"
// @Success 200 {object} model.User
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security ApiKeyAuth
// @Router /api/v1/users/{id}/role [put]
func (s *userService) UpdateUserRoleHandler(c *fiber.Ctx) error {
	id := c.Params("id")
	roleID := c.Query("role_id")
	if roleID == "" {
		return c.Status(http.StatusBadRequest).JSON(model.ErrorResponse{Message: "role_id query parameter required"})
	}

	user, err := s.UpdateUserRole(c.Context(), id, roleID)
	if err != nil {
		if err.Error() == "user not found" {
			return c.Status(http.StatusNotFound).JSON(model.ErrorResponse{Message: err.Error()})
		}
		return c.Status(http.StatusInternalServerError).JSON(model.ErrorResponse{Message: err.Error()})
	}
	user.PasswordHash = ""
	return c.JSON(user)
}