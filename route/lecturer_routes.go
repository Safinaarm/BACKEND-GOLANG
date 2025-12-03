// File: BACKEND-UAS/route/lecturer_routes.go
// Note: Adapted to Fiber v2. No auth for now.

package route

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"BACKEND-UAS/pgmongo/service"
)

func RegisterLecturerRoutes(app *fiber.App, lecturerSvc service.LecturerService) {
	api := app.Group("/api/v1")
	api.Get("/lecturers", getAllLecturers(lecturerSvc))
	api.Get("/lecturers/:id/advisees", getLecturerAdvisees(lecturerSvc))
}

func getAllLecturers(lecturerSvc service.LecturerService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := context.Background()
		lecturers, err := lecturerSvc.FindAll(ctx)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(lecturers)
	}
}

func getLecturerAdvisees(lecturerSvc service.LecturerService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		idStr := c.Params("id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid lecturer ID format"})
		}

		ctx := context.Background()
		advisees, err := lecturerSvc.FindAdvisees(ctx, id)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(advisees)
	}
}