// postgre/route/user_route.go (updated with CreateUserReq for POST)
package route

import (
	"net/http"

	"BACKEND-UAS/middleware"
	"BACKEND-UAS/pgmongo/model"
	"BACKEND-UAS/pgmongo/service"

	"github.com/gofiber/fiber/v2"
)

// @Summary Get all users
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
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /api/v1/users [get]
func UserRoute(app *fiber.App, userSvc service.UserService, authMiddleware *middleware.AuthMiddlewareConfig) {
	v1 := app.Group("/api/v1")
	users := v1.Group("/users")

	// Protected routes for admin
	users.Use(authMiddleware.AuthRequired())

	// GET /api/v1/users (list all, require read:users)
	users.Get("/", middleware.RequirePermission("read:users"), func(c *fiber.Ctx) error {
		usersList, err := userSvc.GetAll(c.Context())
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": err.Error()})
		}
		// Hide passwords in list
		for _, u := range usersList {
			u.PasswordHash = ""
		}
		return c.Status(http.StatusOK).JSON(fiber.Map{
			"status": "success",
			"data":   usersList,
		})
	})

	// @Summary Get user by ID
	// @Description Mengambil detail user dengan ID tertentu
	// @Tags Users
	// @Accept json
	// @Produce json
	// @Param id path string true "User ID"
// @Success 200 {object} model.User
	// @Failure 404 {object} model.ErrorResponse "User not found"
	// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
	// @Router /api/v1/users/{id} [get]
	users.Get("/:id", middleware.RequirePermission("read:users"), func(c *fiber.Ctx) error {
		id := c.Params("id")
		user, err := userSvc.GetByID(c.Context(), id)
		if err != nil {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"status": "error", "message": err.Error()})
		}
		// Hide password
		user.PasswordHash = ""
		return c.Status(http.StatusOK).JSON(fiber.Map{
			"status": "success",
			"data":   user,
		})
	})

	// @Summary Create user
	// @Description Membuat user baru dengan data yang diberikan
	// @Tags Users
	// @Accept json
	// @Produce json
	// @Param user body service.CreateUserReq true "Data user baru"
	// @Success 201 {object} model.User
	// @Failure 400 {object} model.ErrorResponse "Invalid body"
	// @Failure 500 {object} model.ErrorResponse "Failed to create"
// @Security ApiKeyAuth
	// @Router /api/v1/users [post]
	users.Post("/", middleware.RequirePermission("create:users"), func(c *fiber.Ctx) error {
		var req service.CreateUserReq // Use CreateUserReq for input (plain password)
		if err := c.BodyParser(&req); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "invalid body"})
		}
		user, err := userSvc.Create(c.Context(), &req)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": err.Error()})
		}
		// Hide password
		user.PasswordHash = ""
		return c.Status(http.StatusCreated).JSON(fiber.Map{
			"status": "success",
			"data":   user,
		})
	})

	// @Summary Update user
	// @Description Memperbarui data user dengan ID tertentu
	// @Tags Users
	// @Accept json
	// @Produce json
	// @Param id path string true "User ID"
	// @Param user body model.User true "Data user yang diupdate"
	// @Success 200 {object} model.User
	// @Failure 400 {object} model.ErrorResponse "Invalid body"
	// @Failure 500 {object} model.ErrorResponse "Failed to update"
// @Security ApiKeyAuth
	// @Router /api/v1/users/{id} [put]
	users.Put("/:id", middleware.RequirePermission("update:users"), func(c *fiber.Ctx) error {
		id := c.Params("id")
		var req model.User
		if err := c.BodyParser(&req); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "invalid body"})
		}
		user, err := userSvc.Update(c.Context(), id, &req)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": err.Error()})
		}
		// Hide password
		user.PasswordHash = ""
		return c.Status(http.StatusOK).JSON(fiber.Map{
			"status": "success",
			"data":   user,
		})
	})

	// @Summary Delete user
	// @Description Menghapus user dengan ID tertentu
	// @Tags Users
	// @Accept json
	// @Produce json
	// @Param id path string true "User ID"
	// @Success 200 {object} DeleteResponse
	// @Failure 500 {object} model.ErrorResponse "Failed to delete"
// @Security ApiKeyAuth
	// @Router /api/v1/users/{id} [delete]
	users.Delete("/:id", middleware.RequirePermission("delete:users"), func(c *fiber.Ctx) error {
		id := c.Params("id")
		err := userSvc.Delete(c.Context(), id)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": err.Error()})
		}
		return c.Status(http.StatusOK).JSON(fiber.Map{
			"status":  "success",
			"message": "user deleted successfully",
		})
	})

	// @Summary Update user role
	// @Description Memperbarui role user dengan ID tertentu
	// @Tags Users
	// @Accept json
	// @Produce json
	// @Param id path string true "User ID"
	// @Param role body RoleUpdateRequest true "Role ID baru"
	// @Success 200 {object} model.User
	// @Failure 400 {object} model.ErrorResponse "Role ID required"
	// @Failure 500 {object} model.ErrorResponse "Failed to update role"
// @Security ApiKeyAuth
	// @Router /api/v1/users/{id}/role [put]
	users.Put("/:id/role", middleware.RequirePermission("update_role:users"), func(c *fiber.Ctx) error {
		id := c.Params("id")
		var req struct {
			RoleID string `json:"role_id"`
		}
		if err := c.BodyParser(&req); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "invalid body"})
		}
		if req.RoleID == "" {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "role_id required"})
		}
		user, err := userSvc.UpdateUserRole(c.Context(), id, req.RoleID)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": err.Error()})
		}
		// Hide password
		user.PasswordHash = ""
		return c.Status(http.StatusOK).JSON(fiber.Map{
			"status": "success",
			"data":   user,
		})
	})
}