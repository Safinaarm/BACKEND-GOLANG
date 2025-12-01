// middleware/auth.go
package middleware

import (
	"net/http"
	"strings"

	"BACKEND-UAS/pgmongo/jwt"
	"BACKEND-UAS/pgmongo/repository"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type AuthMiddlewareConfig struct {
	JWTService jwt.JWTService
	UserRepo   repository.UserRepository
}

func NewAuthMiddleware(jwtSvc jwt.JWTService, userRepo repository.UserRepository) *AuthMiddlewareConfig {
	return &AuthMiddlewareConfig{
		JWTService: jwtSvc,
		UserRepo:   userRepo,
	}
}

// =======================================================
// 1) VALIDASI TOKEN + SIMPAN USER INFO
// =======================================================
func (m *AuthMiddlewareConfig) AuthRequired() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"status":  "error",
				"message": "missing authorization header",
			})
		}
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"status":  "error",
				"message": "invalid token format",
			})
		}
		tokenString := parts[1]
		token, err := m.JWTService.ValidateToken(tokenString)
		if err != nil || !token.Valid {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"status":  "error",
				"message": "invalid or expired token",
			})
		}
		claims, ok := token.Claims.(*jwt.Claims)
		if !ok {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"status":  "error",
				"message": "invalid claims",
			})
		}

		// Asumsi claims.UserID adalah string UUID
		userIDStr := claims.UserID // Langsung string
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"status":  "error",
				"message": "invalid user ID in claims",
			})
		}

		// simpan ke Fiber Locals (match dengan route: user_id as string, role as string)
		c.Locals("user_id", userIDStr) // String UUID untuk parse di route
		c.Locals("role", claims.Role)  // String role
		c.Locals("userId", userID)     // UUID untuk service/repo

		// Load permissions dari DB via roleID (pakai method existing GetPermissionsByRoleID)
		roleID := claims.RoleID // Asumsi claims.RoleID adalah string
		perms, err := m.UserRepo.GetPermissionsByRoleID(c.Context(), roleID)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": "failed to load permissions",
			})
		}
		c.Locals("permissions", perms)

		return c.Next()
	}
}

// =======================================================
// 2) MIDDLEWARE CHECK PERMISSION ROUTE
// =======================================================
func RequirePermission(permission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		permAny := c.Locals("permissions")
		if permAny == nil {
			return c.Status(http.StatusForbidden).JSON(fiber.Map{
				"status":  "error",
				"message": "permissions not loaded",
			})
		}
		perms := permAny.([]string)
		for _, p := range perms {
			if p == permission {
				return c.Next()
			}
		}
		return c.Status(http.StatusForbidden).JSON(fiber.Map{
			"status":  "error",
			"message": "permission denied",
		})
	}
}