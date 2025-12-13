package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"BACKEND-UAS/pgmongo/model"
	"BACKEND-UAS/pgmongo/service"

	jwtpkg "github.com/golang-jwt/jwt/v5"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

/* =======================
   MOCK USER REPOSITORY
======================= */

type mockUserRepo struct{}

func (m *mockUserRepo) FindByUsernameOrEmail(ctx context.Context, identifier string) (*model.User, error) {
	return &model.User{
		ID:           "user-123",
		Username:     "admin",
		Email:        "admin@test.com",
		PasswordHash: "hashed-password",
		RoleID:       "role-1",
		IsActive:     true,
	}, nil
}

func (m *mockUserRepo) FindByID(ctx context.Context, id string) (*model.User, error) {
	return nil, nil
}
func (m *mockUserRepo) GetAll(ctx context.Context) ([]*model.User, error) {
	return nil, nil
}
func (m *mockUserRepo) Create(ctx context.Context, user *model.User) error {
	return nil
}
func (m *mockUserRepo) Update(ctx context.Context, id string, user *model.User) error {
	return nil
}
func (m *mockUserRepo) Delete(ctx context.Context, id string) error {
	return nil
}
func (m *mockUserRepo) UpdateRole(ctx context.Context, id, roleID string) error {
	return nil
}
func (m *mockUserRepo) GetRoleNameByID(ctx context.Context, roleID string) (string, error) {
	return "admin", nil
}
func (m *mockUserRepo) GetPermissionsByRoleID(ctx context.Context, roleID string) ([]string, error) {
	return []string{"read:users"}, nil
}

/* =======================
   MOCK JWT SERVICE
======================= */

type mockJWTService struct{}

func (m *mockJWTService) GenerateToken(userID, roleID, role string) (string, error) {
	return "mock-token", nil
}

func (m *mockJWTService) ValidateToken(tokenStr string) (*jwtpkg.Token, error) {
	return &jwtpkg.Token{}, nil
}

func (m *mockJWTService) CheckPasswordHash(password, hash string) bool {
	return password == "password"
}

func (m *mockJWTService) HashPassword(password string) (string, error) {
	return password, nil
}

/* =======================
   TEST LOGIN ROUTE
======================= */

func TestAuthLoginRoute_Success(t *testing.T) {
	app := fiber.New()

	userRepo := &mockUserRepo{}
	jwtSvc := &mockJWTService{}

	authSvc := service.NewAuthService(userRepo, jwtSvc)

	app.Post("/api/v1/auth/login", authSvc.LoginHandler)

	body := map[string]string{
	"username": "admin",
	"password": "password",
}


	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/auth/login",
		bytes.NewBuffer(payload),
	)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
