// main.go
package main

import (
	"log"

	"github.com/gofiber/fiber/v2"

	"BACKEND-UAS/config"
	"BACKEND-UAS/middleware"
	"BACKEND-UAS/postgre/jwt"
	"BACKEND-UAS/postgre/repository"
	"BACKEND-UAS/route"
	"BACKEND-UAS/postgre/service"
)

func main() {
	cfg := config.NewConfig()

	// ============================
	// INIT SERVICES
	// ============================
	userRepo := repository.NewUserRepository(cfg.Connection.PostgresDB)
	jwtSvc := jwt.NewJWTService(cfg.JWTSecret)
	authSvc := service.NewAuthService(userRepo, jwtSvc)
	authMiddleware := middleware.NewAuthMiddleware(jwtSvc, userRepo)

	// ============================
	// FIBER APP
	// ============================
	app := fiber.New()
	route.AuthRoute(app, authSvc, authMiddleware)

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "OK"})
	})

	// ============================
	// RUN SERVER
	// ============================
	log.Println("🚀 Server running on port :" + cfg.Port)
	log.Fatal(app.Listen(":" + cfg.Port))
}