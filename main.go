package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"

	"BACKEND-UAS/config"
	"BACKEND-UAS/middleware"
	"BACKEND-UAS/pgmongo/jwt"
	"BACKEND-UAS/pgmongo/repository"
	"BACKEND-UAS/route"
	"BACKEND-UAS/pgmongo/service"
)

func main() {
	cfg := config.NewConfig()

	// ============================
	// INIT SERVICES
	// ============================
	userRepo := repository.NewUserRepository(cfg.Connection.PostgresDB)
	jwtSvc := jwt.NewJWTService(cfg.JWTSecret)
	authSvc := service.NewAuthService(userRepo, jwtSvc)
	userSvc := service.NewUserService(userRepo, jwtSvc) // Tambah ini buat users CRUD
	authMiddleware := middleware.NewAuthMiddleware(jwtSvc, userRepo)

	// ============================
	// INIT FIBER APP
	// ============================
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(500).JSON(fiber.Map{"status": "error", "message": err.Error()})
		},
	})

	// CORS middleware (buat Postman/browser)
	app.Use(cors.New())

	// Routes
	route.AuthRoute(app, authSvc, authMiddleware)
	route.UserRoute(app, userSvc, authMiddleware) // Tambah ini! Buat /api/v1/users

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "OK"})
	})

	// ============================
	// RUN SERVER
	// ============================
	log.Println("🚀 Server running on port :" + cfg.Port)
	log.Fatal(app.Listen(":" + cfg.Port))
}