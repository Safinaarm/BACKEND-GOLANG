// postgre/route/auth_route.go
package route

import (
	"net/http"

	"BACKEND-UAS/middleware"
	"BACKEND-UAS/pgmongo/service"

	"github.com/gofiber/fiber/v2"
)

func AuthRoute(app *fiber.App, authSvc service.AuthService, authMiddleware *middleware.AuthMiddlewareConfig) {
	v1 := app.Group("/api/v1")
	auth := v1.Group("/auth")

	// @Summary Login user
	// @Description Melakukan login dengan username/email dan password, menghasilkan access dan refresh token
	// @Tags Auth
	// @Accept json
	// @Produce json
	// @Param login body LoginRequest true "Credentials (username or email and password)"
	// @Success 200 {object} AuthResponse
	// @Failure 400 {object} model.ErrorResponse "Invalid credentials or missing fields"
	// @Failure 401 {object} model.ErrorResponse "Account inactive"
	// @Failure 500 {object} model.ErrorResponse "Internal server error"
	// @Router /api/v1/auth/login [post]
	auth.Post("/login", func(c *fiber.Ctx) error {
		var req struct {
			Identifier string `json:"username"` // allow username or email
			Password   string `json:"password"`
		}
		if err := c.BodyParser(&req); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "invalid body"})
		}
		if req.Identifier == "" || req.Password == "" {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "identifier and password required"})
		}
		access, refresh, user, role, perms, err := authSvc.Login(c.Context(), req.Identifier, req.Password)
		if err != nil {
			if err.Error() == "user not found" || err.Error() == "invalid credentials" {
				return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"status": "error", "message": "invalid credentials"})
			}
			if err.Error() == "account inactive" {
				return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"status": "error", "message": "account inactive"})
			}
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": err.Error()})
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

	// @Summary Refresh token
	// @Description Memperbarui access dan refresh token menggunakan refresh token yang valid
	// @Tags Auth
	// @Accept json
	// @Produce json
	// @Param refresh_token body RefreshRequest true "Refresh token"
	// @Success 200 {object} AuthResponse
	// @Failure 400 {object} model.ErrorResponse "Invalid refresh token"
	// @Failure 401 {object} model.ErrorResponse "Account inactive"
	// @Failure 500 {object} model.ErrorResponse "Internal server error"
	// @Router /api/v1/auth/refresh [post]
	auth.Post("/refresh", func(c *fiber.Ctx) error {
		var req struct {
			RefreshToken string `json:"refreshToken"`
		}
		if err := c.BodyParser(&req); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "invalid body"})
		}
		if req.RefreshToken == "" {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "refreshToken required"})
		}
		access, refresh, user, role, perms, err := authSvc.Refresh(c.Context(), req.RefreshToken)
		if err != nil {
			if err.Error() == "invalid refresh token" {
				return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"status": "error", "message": "invalid refresh token"})
			}
			if err.Error() == "account inactive" {
				return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"status": "error", "message": "account inactive"})
			}
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": err.Error()})
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

	// @Summary Logout user
	// @Description Melakukan logout dengan invalidasi refresh token (client-side untuk demo)
	// @Tags Auth
	// @Accept json
	// @Produce json
	// @Param refresh_token body RefreshRequest true "Refresh token to invalidate"
	// @Success 200 {object} LogoutResponse
	// @Failure 400 {object} model.ErrorResponse "Invalid refresh token"
	// @Failure 500 {object} model.ErrorResponse "Internal server error"
	// @Router /api/v1/auth/logout [post]
	auth.Post("/logout", func(c *fiber.Ctx) error {
		var req struct {
			RefreshToken string `json:"refreshToken"`
		}
		if err := c.BodyParser(&req); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "invalid body"})
		}
		// RefreshToken is optional for demo; client should discard tokens client-side
		err := authSvc.Logout(c.Context(), req.RefreshToken)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": err.Error()})
		}
		return c.Status(http.StatusOK).JSON(fiber.Map{
			"status":  "success",
			"message": "logged out successfully",
		})
	})

	// @Summary Get profile
	// @Description Mengambil data user berdasarkan JWT token
	// @Tags Auth
	// @Accept json
	// @Produce json
	// @Security ApiKeyAuth
	// @Success 200 {object} ProfileResponse
	// @Failure 401 {object} model.ErrorResponse "Unauthorized"
	// @Router /api/v1/auth/profile [get]
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