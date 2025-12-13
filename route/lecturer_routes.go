package route

import (
	"BACKEND-UAS/middleware"
	"BACKEND-UAS/pgmongo/service"

	"github.com/gofiber/fiber/v2"
)

// SetupLecturerRoutes men-setup semua route terkait lecturer
func SetupLecturerRoutes(app *fiber.App, lecturerSvc service.LecturerService, authMiddleware *middleware.AuthMiddlewareConfig) {
	v1 := app.Group("/api/v1")
	lecturers := v1.Group("/lecturers")

	// Semua route lecturer memerlukan autentikasi
	lecturers.Use(authMiddleware.AuthRequired())

	// GET /api/v1/lecturers
	lecturers.Get("/", lecturerSvc.GetAllLecturersHandler)

	// GET /api/v1/lecturers/:id/advisees
	lecturers.Get("/:id/advisees", lecturerSvc.GetAdviseesHandler)
}