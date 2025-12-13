// postgre/route/auth_route.go
package route

import (
	"BACKEND-UAS/middleware"
	"BACKEND-UAS/pgmongo/service"

	"github.com/gofiber/fiber/v2"
)

func AuthRoute(app *fiber.App, authSvc service.AuthService, authMiddleware *middleware.AuthMiddlewareConfig) {
	v1 := app.Group("/api/v1")
	auth := v1.Group("/auth")

	auth.Post("/login", authSvc.LoginHandler)
	auth.Post("/refresh", authSvc.RefreshHandler)
	auth.Post("/logout", authSvc.LogoutHandler)

	profile := auth.Group("/profile")
	profile.Use(authMiddleware.AuthRequired())
	profile.Get("/", authSvc.ProfileHandler)
}