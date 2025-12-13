// File: BACKEND-UAS/route/report_routes.go
package route

import (
	"BACKEND-UAS/middleware"
	"BACKEND-UAS/pgmongo/service"

	"github.com/gofiber/fiber/v2"
)

// SetupReportRoutes sets up the report routes with auth middleware
func SetupReportRoutes(app *fiber.App, reportSvc service.ReportService, authM *middleware.AuthMiddlewareConfig) {
	v1 := app.Group("/api/v1")
	{
		reports := v1.Group("/reports")
		reports.Use(authM.AuthRequired()) // Apply auth middleware
		{
			// GET /api/v1/reports/statistics - Achievement statistics based on user role
			reports.Get("/statistics", reportSvc.HandleGetAchievementStatistics())

			// GET /api/v1/reports/students/{student_id}/statistics - Student-specific achievement statistics with access check
			reports.Get("/students/:student_id/statistics", reportSvc.HandleGetStudentAchievementStatistics())
		}
	}
}