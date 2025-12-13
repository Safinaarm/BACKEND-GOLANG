package route

import (
	"BACKEND-UAS/middleware"
	"BACKEND-UAS/pgmongo/service"

	"github.com/gofiber/fiber/v2"
)

func SetupAchievementRoutes(app *fiber.App, svc *service.AchievementService, authMiddleware *middleware.AuthMiddlewareConfig) {
	v1 := app.Group("/api/v1")
	achievements := v1.Group("/achievements")

	// Semua route achievement butuh autentikasi
	achievements.Use(authMiddleware.AuthRequired())

	// List achievements (dengan role dari locals)
	achievements.Get("/", svc.ListHandler)

	// Detail
	achievements.Get("/:id", svc.DetailHandler)

	// Create
	achievements.Post("/", svc.CreateHandler)

	// Update
	achievements.Put("/:id", svc.UpdateHandler)

	// Delete
	achievements.Delete("/:id", svc.DeleteHandler)

	// Submit
	achievements.Post("/:id/submit", svc.SubmitHandler)

	// Verify
	achievements.Post("/:id/verify", svc.VerifyHandler)

	// Reject
	achievements.Post("/:id/reject", svc.RejectHandler)

	// History
	achievements.Get("/:id/history", svc.HistoryHandler)

	// Upload attachment
	achievements.Post("/:id/attachments", svc.UploadAttachmentHandler)
}