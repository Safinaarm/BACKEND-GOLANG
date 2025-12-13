// postgre/route/user_route.go
package route

import (
	"BACKEND-UAS/middleware"
	"BACKEND-UAS/pgmongo/service"

	"github.com/gofiber/fiber/v2"
)

func UserRoute(app *fiber.App, userSvc service.UserService, authMiddleware *middleware.AuthMiddlewareConfig) {
	v1 := app.Group("/api/v1")
	users := v1.Group("/users")

	// Semua route user butuh autentikasi
	users.Use(authMiddleware.AuthRequired())

	// List users
	users.Get("/", middleware.RequirePermission("read:users"), userSvc.ListUsersHandler)

	// Get user by ID
	users.Get("/:id", middleware.RequirePermission("read:users"), userSvc.GetUserHandler)

	// Create user
	users.Post("/", middleware.RequirePermission("create:users"), userSvc.CreateUserHandler)

	// Update user
	users.Put("/:id", middleware.RequirePermission("update:users"), userSvc.UpdateUserHandler)

	// Delete user
	users.Delete("/:id", middleware.RequirePermission("delete:users"), userSvc.DeleteUserHandler)

	// Update user role (query param role_id)
	users.Put("/:id/role", middleware.RequirePermission("update_role:users"), userSvc.UpdateUserRoleHandler)
}