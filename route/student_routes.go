package route

import (
	"BACKEND-UAS/middleware"
	"BACKEND-UAS/pgmongo/service"

	"github.com/gofiber/fiber/v2"
)

func SetupStudentRoutes(app *fiber.App, studentSvc *service.StudentService, authMiddleware *middleware.AuthMiddlewareConfig) {
	v1 := app.Group("/api/v1")
	students := v1.Group("/students")

	// Semua route student butuh autentikasi
	students.Use(authMiddleware.AuthRequired())

	// GET /api/v1/students
	students.Get("/", studentSvc.GetAllStudentsHandler)

	// GET /api/v1/students/me
	students.Get("/me", studentSvc.GetOwnProfileHandler)

	// GET /api/v1/students/{id}
	students.Get("/:id", studentSvc.GetStudentByIDHandler)

	// GET /api/v1/students/{id}/achievements
	students.Get("/:id/achievements", studentSvc.GetAchievementsHandler)

	// PUT /api/v1/students/{id}/advisor
	students.Put("/:id/advisor", studentSvc.UpdateAdvisorHandler)
}