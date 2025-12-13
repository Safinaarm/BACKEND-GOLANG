package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"BACKEND-UAS/pgmongo/model"
	"BACKEND-UAS/pgmongo/service"
	"BACKEND-UAS/route"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

/* =====================
   MOCK ROUTE SERVICE
===================== */

type mockUserRouteService struct{}

// Create implements service.UserService.
func (m *mockUserRouteService) Create(ctx context.Context, req *service.CreateUserReq) (*model.User, error) {
	panic("unimplemented")
}

// Delete implements service.UserService.
func (m *mockUserRouteService) Delete(ctx context.Context, id string) error {
	panic("unimplemented")
}

// GetAll implements service.UserService.
func (m *mockUserRouteService) GetAll(ctx context.Context) ([]*model.User, error) {
	panic("unimplemented")
}

// GetByID implements service.UserService.
func (m *mockUserRouteService) GetByID(ctx context.Context, id string) (*model.User, error) {
	panic("unimplemented")
}

// Update implements service.UserService.
func (m *mockUserRouteService) Update(ctx context.Context, id string, req *model.User) (*model.User, error) {
	panic("unimplemented")
}

// UpdateUserRole implements service.UserService.
func (m *mockUserRouteService) UpdateUserRole(ctx context.Context, id string, roleID string) (*model.User, error) {
	panic("unimplemented")
}

func (m *mockUserRouteService) ListUsersHandler(c *fiber.Ctx) error {
	return c.Status(200).JSON(fiber.Map{"ok": true})
}
func (m *mockUserRouteService) GetUserHandler(c *fiber.Ctx) error {
	return c.SendStatus(200)
}
func (m *mockUserRouteService) CreateUserHandler(c *fiber.Ctx) error {
	return c.SendStatus(201)
}
func (m *mockUserRouteService) UpdateUserHandler(c *fiber.Ctx) error {
	return c.SendStatus(200)
}
func (m *mockUserRouteService) DeleteUserHandler(c *fiber.Ctx) error {
	return c.SendStatus(204)
}
func (m *mockUserRouteService) UpdateUserRoleHandler(c *fiber.Ctx) error {
	return c.SendStatus(200)
}

/* =====================
   TEST
===================== */

func TestUserRoute_ListUsers_OK(t *testing.T) {
	app := fiber.New()

	userSvc := &mockUserRouteService{}

	route.UserRoute(app, userSvc, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
