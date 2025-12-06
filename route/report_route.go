// File: BACKEND-UAS/route/report_routes.go
package route

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"BACKEND-UAS/middleware"
	"BACKEND-UAS/pgmongo/service"
)

// SetupReportRoutes sets up the report routes with auth middleware
func SetupReportRoutes(app *fiber.App, reportSvc service.ReportService, authM *middleware.AuthMiddlewareConfig) {
	v1 := app.Group("/api/v1")
	{
		reports := v1.Group("/reports")
		reports.Use(authM.AuthRequired()) // Apply auth middleware
		{
			// GET /api/v1/reports/statistics - Achievement statistics based on user role
			reports.Get("/statistics", func(c *fiber.Ctx) error {
				userIDStr, ok := c.Locals("user_id").(string)
				if !ok {
					return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
				}
				userID, err := uuid.Parse(userIDStr)
				if err != nil {
					return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
				}

				stats, err := reportSvc.GetAchievementStatistics(c.Context(), userID)
				if err != nil {
					if err.Error() == "access denied" || err.Error() == "no student profile found" || err.Error() == "no lecturer profile found" {
						return c.Status(http.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
					}
					return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
				}

				return c.JSON(stats)
			})

			// GET /api/v1/reports/student/:id - Student-specific achievement statistics with access check
			reports.Get("/student/:id", func(c *fiber.Ctx) error {
				idStr := c.Params("id")
				studentID, err := uuid.Parse(idStr)
				if err != nil {
					return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid student ID"})
				}

				userIDStr, ok := c.Locals("user_id").(string)
				if !ok {
					return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
				}
				userID, err := uuid.Parse(userIDStr)
				if err != nil {
					return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
				}

				stats, err := reportSvc.GetStudentAchievementStatistics(c.Context(), studentID, userID)
				if err != nil {
					if err.Error() == "access denied" || err.Error() == "student not found" {
						return c.Status(http.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
					}
					return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
				}

				return c.JSON(stats)
			})
		}
	}
}