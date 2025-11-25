// postgre/route/auth_route.go
package route

import (
	"net/http"

	"BACKEND-UAS/middleware"
	"BACKEND-UAS/postgre/service"

	"github.com/gofiber/fiber/v2"
)

func AuthRoute(app *fiber.App, authSvc service.AuthService, authMiddleware *middleware.AuthMiddlewareConfig) {
	v1 := app.Group("/api/v1")
	auth := v1.Group("/auth")

	auth.Post("/login", func(c *fiber.Ctx) error {
		var req struct {
			Identifier string `json:"username"` // allow username or email
			Password   string `json:"password"`
		}
		if err := c.BodyParser(&req); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "invalid body"})
		}
		access, refresh, user, role, perms, err := authSvc.Login(c.Context(), req.Identifier, req.Password)
		if err != nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"status": "error", "message": err.Error()})
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
	})

	// Protected profile endpoint as example
	profile := auth.Group("/profile")
	profile.Use(authMiddleware.AuthRequired())
	profile.Get("/", func(c *fiber.Ctx) error {
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
	})
}