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
// @Summary List achievements (filtered by role)
// @Description Mengambil daftar prestasi berdasarkan role user (mahasiswa: own, dosen: advisees, admin: all), dengan filter status dan pagination
// @Tags Achievements
// @Accept json
// @Produce json
// @Param page query int false "Page number (default 1)"
// @Param limit query int false "Items per page (default 10)"
// @Param status query string false "Filter status (draft, submitted, verified, rejected, deleted)"
// @Success 200 {object} AchievementListResponse "Success with paginated achievements"
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /achievements [get]
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

	// @Summary Get achievement detail
	// @Description Mengambil detail lengkap prestasi berdasarkan ID, termasuk history status
	// @Tags Achievements
	// @Accept json
	// @Produce json
	// @Param id path string true "Achievement ID (UUID)"
	// @Success 200 {object} model.AchievementDetailResponse "Achievement detail with history"
	// @Failure 400 {object} model.ErrorResponse "Invalid ID"
	// @Failure 404 {object} model.ErrorResponse "Achievement not found or deleted"
	// @Failure 500 {object} model.ErrorResponse "Internal server error"
	// @Security ApiKeyAuth
	// @Router /achievements/{id} [get]
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

	// @Summary Create achievement
	// @Description Membuat prestasi baru (hanya untuk mahasiswa)
	// @Tags Achievements
	// @Accept json
	// @Produce json
	// @Param achievement body model.Achievement true "Achievement data"
// @Success 201 {object} model.AchievementReference "Created achievement reference"
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Failure 500 {object} model.ErrorResponse "Failed to create"
// @Security ApiKeyAuth
	// @Router /achievements [post]
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

	// @Summary Update achievement
	// @Description Memperbarui prestasi (hanya untuk draft atau rejected status, mahasiswa)
// @Tags Achievements
	// @Accept json
	// @Produce json
	// @Param id path string true "Achievement ID (UUID)"
	// @Param achievement body model.Achievement true "Updated achievement data"
// @Success 200 {object} nil "Updated successfully"
// @Failure 400 {object} model.ErrorResponse "Invalid ID or wrong status"
// @Failure 500 {object} model.ErrorResponse "Failed to update"
// @Security ApiKeyAuth
	// @Router /achievements/{id} [put]
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
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"message": "Updated successfully"})
	})

	// @Summary Delete achievement
	// @Description Soft delete prestasi (hanya untuk draft status, mahasiswa)
// @Tags Achievements
	// @Accept json
	// @Produce json
	// @Param id path string true "Achievement ID (UUID)"
	// @Param user_id query string true "User ID (UUID) - must be student owner"
// @Success 200 {object} nil "Deleted successfully"
// @Failure 400 {object} model.ErrorResponse "Invalid ID or wrong status"
// @Failure 500 {object} model.ErrorResponse "Failed to delete"
// @Security ApiKeyAuth
	// @Router /achievements/{id} [delete]
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

	// @Summary Submit achievement
	// @Description Submit prestasi untuk verifikasi (dari draft atau rejected)
// @Tags Achievements
	// @Accept json
	// @Produce json
	// @Param id path string true "Achievement ID (UUID)"
	// @Param user_id query string true "User ID (UUID) - submitter"
// @Success 200 {object} nil "Submitted successfully"
// @Failure 400 {object} model.ErrorResponse "Invalid ID or wrong status"
// @Failure 500 {object} model.ErrorResponse "Failed to submit"
// @Security ApiKeyAuth
	// @Router /achievements/{id}/submit [post]
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

	// @Summary Verify achievement
	// @Description Verifikasi prestasi (oleh dosen wali/admin, dari submitted status)
// @Tags Achievements
	// @Accept json
	// @Produce json
	// @Param id path string true "Achievement ID (UUID)"
	// @Param verified_by query string true "Verifier User ID (UUID)"
// @Success 200 {object} nil "Verified successfully"
// @Failure 400 {object} model.ErrorResponse "Invalid ID or wrong status"
// @Failure 500 {object} model.ErrorResponse "Failed to verify"
// @Security ApiKeyAuth
	// @Router /achievements/{id}/verify [post]
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

	// @Summary Reject achievement
	// @Description Tolak prestasi dengan note (oleh dosen wali/admin, dari submitted status)
// @Tags Achievements
	// @Accept json
	// @Produce json
	// @Param id path string true "Achievement ID (UUID)"
	// @Param verified_by query string true "Rejector User ID (UUID)"
	// @Param rejection_note body RejectionRequest true "Rejection note"
// @Success 200 {object} nil "Rejected successfully"
// @Failure 400 {object} model.ErrorResponse "Invalid ID, note required, or wrong status"
// @Failure 500 {object} model.ErrorResponse "Failed to reject"
// @Security ApiKeyAuth
	// @Router /achievements/{id}/reject [post]
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

	// @Summary Get achievement history
	// @Description Mengambil riwayat status prestasi
	// @Tags Achievements
	// @Accept json
	// @Produce json
	// @Param id path string true "Achievement ID (UUID)"
	// @Success 200 {array} model.StatusHistory "Status history array"
	// @Failure 400 {object} model.ErrorResponse "Invalid ID"
	// @Failure 500 {object} model.ErrorResponse "Internal server error"
	// @Security ApiKeyAuth
	// @Router /achievements/{id}/history [get]
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

	// @Summary Upload attachment
	// @Description Upload file attachment ke prestasi (hanya untuk non-deleted)
// @Tags Achievements
	// @Accept multipart/form-data
	// @Produce json
	// @Param id path string true "Achievement ID (UUID)"
	// @Param file formData file true "Attachment file"
// @Param file_name formData string true "Original file name"
// @Param file_type formData string true "MIME type"
// @Success 200 {object} model.Attachment "Uploaded attachment"
// @Failure 400 {object} model.ErrorResponse "No file uploaded"
// @Failure 404 {object} model.ErrorResponse "Achievement not found"
// @Failure 500 {object} model.ErrorResponse "Failed to upload"
// @Security ApiKeyAuth
	// @Router /achievements/{id}/attachments [post]
	ach.Post("/:id/attachments", func(c *fiber.Ctx) error {
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