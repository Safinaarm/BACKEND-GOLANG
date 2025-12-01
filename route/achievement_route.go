// File: BACKEND-UAS/route/achievement_route.go
package route

import (
	"log"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"BACKEND-UAS/middleware"
	"BACKEND-UAS/pgmongo/model"
	"BACKEND-UAS/pgmongo/service"
)

// RegisterAchievementRoutes registers all achievement routes directly
func RegisterAchievementRoutes(app *fiber.App, svc *service.AchievementService, authMiddleware *middleware.AuthMiddlewareConfig) {
	ach := app.Group("/api/v1/achievements")
	ach.Use(authMiddleware.AuthRequired()) // Apply auth ke semua achievement routes

	// Permission examples: "achievement:read", "achievement:create", etc. (sesuaikan dengan DB)
	ach.Get("/", middleware.RequirePermission("achievement:read"), func(c *fiber.Ctx) error {
		// GET /api/v1/achievements - List (filtered by role)
		userIDStr, ok := c.Locals("user_id").(string)
		if !ok {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user ID"})
		}
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user ID"})
		}
		role, ok := c.Locals("role").(string)
		if !ok {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid role"})
		}
		log.Printf("Debug: Querying for userID=%s (string: %s), role=%s", userID, userIDStr, role)
		page, err := strconv.Atoi(c.Query("page", "1"))
		if err != nil {
			page = 1
		}
		limit, err := strconv.Atoi(c.Query("limit", "10"))
		if err != nil {
			limit = 10
		}
		status := c.Query("status")
		var statusPtr *string
		if status != "" {
			statusPtr = &status
		}

		resp, err := svc.GetUserAchievements(userID, role, statusPtr, page, limit)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(resp)
	})

	ach.Get("/:id", middleware.RequirePermission("achievement:read"), func(c *fiber.Ctx) error {
		// GET /api/v1/achievements/:id - Detail
		idStr := c.Params("id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
		}

		resp, err := svc.GetAchievementDetail(id)
		if err != nil {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(resp)
	})

	ach.Post("/", middleware.RequirePermission("achievement:create"), func(c *fiber.Ctx) error {
		// POST /api/v1/achievements - Create (Mahasiswa)
		userIDStr, ok := c.Locals("user_id").(string)
		if !ok {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user ID"})
		}
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user ID"})
		}

		var ach model.Achievement
		if err := c.BodyParser(&ach); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		ref, err := svc.CreateAchievement(userID, ach)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(http.StatusCreated).JSON(ref)
	})

	ach.Put("/:id", middleware.RequirePermission("achievement:update"), func(c *fiber.Ctx) error {
		// PUT /api/v1/achievements/:id - Update (Mahasiswa)
		idStr := c.Params("id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
		}

		var updatedAch model.Achievement
		if err := c.BodyParser(&updatedAch); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		err = svc.UpdateAchievement(id, updatedAch)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"message": "Updated successfully"})
	})

	ach.Delete("/:id", middleware.RequirePermission("achievement:delete"), func(c *fiber.Ctx) error {
		// DELETE /api/v1/achievements/:id - Delete (Mahasiswa)
		idStr := c.Params("id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
		}

		userIDStr, ok := c.Locals("user_id").(string)
		if !ok {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user ID"})
		}
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user ID"})
		}

		err = svc.DeleteAchievement(id, userID)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"message": "Deleted successfully"})
	})

	ach.Post("/:id/submit", func(c *fiber.Ctx) error {
		// POST /api/v1/achievements/:id/submit - Submit for verification
		idStr := c.Params("id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
		}

		userIDStr, ok := c.Locals("user_id").(string)
		if !ok {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user ID"})
		}
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user ID"})
		}

		err = svc.SubmitAchievement(id, userID)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"status": "submitted"})
	})

	ach.Post("/:id/verify", func(c *fiber.Ctx) error {
		// POST /api/v1/achievements/:id/verify - Verify (Dosen Wali)
		idStr := c.Params("id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
		}

		userIDStr, ok := c.Locals("user_id").(string)
		if !ok {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user ID"})
		}
		verifiedBy, err := uuid.Parse(userIDStr)
		if err != nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user ID"})
		}

		err = svc.VerifyAchievement(id, verifiedBy)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"status": "verified"})
	})

	ach.Post("/:id/reject", func(c *fiber.Ctx) error {
		// POST /api/v1/achievements/:id/reject - Reject (Dosen Wali)
		idStr := c.Params("id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
		}

		var req struct {
			RejectionNote string `json:"rejection_note"`
		}
		if err := c.BodyParser(&req); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		if req.RejectionNote == "" {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Rejection note is required"})
		}

		userIDStr, ok := c.Locals("user_id").(string)
		if !ok {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user ID"})
		}
		verifiedBy, err := uuid.Parse(userIDStr)
		if err != nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user ID"})
		}

		err = svc.RejectAchievement(id, verifiedBy, req.RejectionNote)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"status": "rejected"})
	})

	ach.Get("/:id/history", middleware.RequirePermission("achievement:read"), func(c *fiber.Ctx) error {
		// GET /api/v1/achievements/:id/history - Status history
		idStr := c.Params("id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
		}

		histories, err := svc.GetAchievementHistory(id)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(histories)
	})

	ach.Post("/:id/attachments", middleware.RequirePermission("achievement:upload"), func(c *fiber.Ctx) error {
		// POST /api/v1/achievements/:id/attachments - Upload files
		idStr := c.Params("id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
		}

		file, err := c.FormFile("file")
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "No file uploaded"})
		}

		src, err := file.Open()
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		defer src.Close()

		ext := filepath.Ext(file.Filename)
		fileName := uuid.New().String() + ext

		attachment, err := svc.UploadAttachment(id, src, fileName, file.Header.Get("Content-Type"))
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(attachment)
	})
}