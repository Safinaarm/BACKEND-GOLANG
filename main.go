// File: BACKEND-UAS/main.go
package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"

	"BACKEND-UAS/config"
	"BACKEND-UAS/middleware"
	"BACKEND-UAS/pgmongo/jwt"
	"BACKEND-UAS/pgmongo/repository"
	"BACKEND-UAS/pgmongo/service"
	"BACKEND-UAS/route"
)

func main() {
	cfg := config.NewConfig()

	// ============================
	// INIT SERVICES
	// ============================
	userRepo := repository.NewUserRepository(cfg.Connection.PostgresDB)
	jwtSvc := jwt.NewJWTService(cfg.JWTSecret)
	authSvc := service.NewAuthService(userRepo, jwtSvc)
	userSvc := service.NewUserService(userRepo, jwtSvc)
	authMiddleware := middleware.NewAuthMiddleware(jwtSvc, userRepo)

	// Achievement repos
	achievementPgRepo := repository.NewAchievementRepository(cfg.Connection.PostgresDB)
	achievementMongoRepo := repository.NewAchievementRepositoryMongo(cfg.Connection.MongoClient)
	achievementSvc := service.NewAchievementService(achievementPgRepo, achievementMongoRepo)

	// Student repos and services
	studentRepo := repository.NewStudentRepository(cfg.Connection.PostgresDB)
	studentSvc := service.NewStudentService(studentRepo, achievementPgRepo)

	// Lecturer repos and services
	lecturerRepo := repository.NewLecturerRepository(cfg.Connection.PostgresDB)
	lecturerSvc := service.NewLecturerService(lecturerRepo, studentRepo)

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
	route.UserRoute(app, userSvc, authMiddleware)
	route.RegisterAchievementRoutes(app, achievementSvc, authMiddleware)
	route.SetupStudentRoutes(app, studentSvc, authMiddleware) // Pass authMiddleware for student routes
	route.SetupLecturerRoutes(app, lecturerSvc, authMiddleware) // Pass authMiddleware for lecturer routes

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