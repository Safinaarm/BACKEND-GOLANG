// postgre/route/user_route.go (updated with CreateUserReq for POST)
package route

import (
	"net/http"

	"BACKEND-UAS/middleware"
	"BACKEND-UAS/pgmongo/model"
	"BACKEND-UAS/pgmongo/service"

	"github.com/gofiber/fiber/v2"
)

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

	// GET /api/v1/users/:id (require read:users)
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

	// POST /api/v1/users (create, require create:users)
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

	// PUT /api/v1/users/:id (update, require update:users)
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

	// DELETE /api/v1/users/:id (delete, require delete:users)
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

	// PUT /api/v1/users/:id/role (update role, require update_role:users)
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