package service

import (
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"BACKEND-UAS/pgmongo/model"
	"BACKEND-UAS/pgmongo/repository"
)

type AchievementService struct {
	postgresRepo *repository.AchievementRepository
	mongoRepo    *repository.AchievementRepositoryMongo
}

func NewAchievementService(pgRepo *repository.AchievementRepository, mongoRepo *repository.AchievementRepositoryMongo) *AchievementService {
	return &AchievementService{
		postgresRepo: pgRepo,
		mongoRepo:    mongoRepo,
	}
}

// Core business logic (dipertahankan & sedikit diperbaiki)
func (s *AchievementService) GetUserAchievements(userID uuid.UUID, role string, status *string, page, limit int) (*model.PaginatedResponse[model.AchievementReference], error) {
	empty := &model.PaginatedResponse[model.AchievementReference]{
		Data:       []model.AchievementReference{},
		Page:       page,
		Limit:      limit,
		Total:      0,
		TotalPages: 0,
	}

	switch role {
	case "Mahasiswa":
		student, err := s.postgresRepo.GetStudentByUserID(userID)
		if err != nil || student == nil {
			return empty, nil
		}
		return s.postgresRepo.GetAchievementReferencesByStudentIDs([]uuid.UUID{student.ID}, status, page, limit)

	case "Dosen Wali":
		lecturer, err := s.postgresRepo.GetLecturerByUserID(userID)
		if err != nil || lecturer == nil {
			return empty, nil
		}
		studentIDs, err := s.postgresRepo.GetStudentIDsByAdvisor(lecturer.ID)
		if err != nil || len(studentIDs) == 0 {
			return empty, nil
		}
		return s.postgresRepo.GetAchievementReferencesByStudentIDs(studentIDs, status, page, limit)

	case "Admin":
		return s.postgresRepo.GetAllAchievementReferences(status, page, limit)

	default:
		return nil, fiber.NewError(http.StatusBadRequest, "invalid role")
	}
}

func (s *AchievementService) GetAchievementDetail(id uuid.UUID) (*model.AchievementDetailResponse, error) {
	ref, err := s.postgresRepo.GetAchievementReferenceByID(id)
	if err != nil {
		return nil, err
	}
	if ref.Status == "deleted" {
		return nil, fiber.NewError(http.StatusNotFound, "achievement not found")
	}

	ach, err := s.mongoRepo.GetAchievementByID(ref.MongoAchievementID)
	if err != nil || ach == nil || ach.DeletedAt != nil {
		return nil, fiber.NewError(http.StatusNotFound, "achievement details not found or deleted")
	}

	return &model.AchievementDetailResponse{
		ID:            ref.ID.String(),
		Student:       ref.Student,
		Status:        ref.Status,
		SubmittedAt:   ref.SubmittedAt,
		VerifiedAt:    ref.VerifiedAt,
		VerifiedBy:    ref.VerifiedBy,
		RejectionNote: ref.RejectionNote,
		Achievement:   *ach,
		StatusHistory: ach.StatusHistory,
	}, nil
}

func (s *AchievementService) CreateAchievement(userID uuid.UUID, ach model.Achievement) (*model.AchievementReference, error) {
	student, err := s.postgresRepo.GetStudentByUserID(userID)
	if err != nil || student == nil {
		return nil, fiber.NewError(http.StatusBadRequest, "student not found")
	}

	ach.StudentID = student.ID
	if err := s.mongoRepo.CreateAchievement(&ach); err != nil {
		return nil, err
	}

	ref := &model.AchievementReference{
		ID:                 uuid.New(),
		StudentID:          student.ID,
		MongoAchievementID: ach.ID.Hex(),
		Status:             "draft",
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
	if err := s.postgresRepo.CreateAchievementReference(ref); err != nil {
		return nil, err
	}
	return ref, nil
}

func (s *AchievementService) UpdateAchievement(id uuid.UUID, updatedAch model.Achievement) error {
	ref, err := s.postgresRepo.GetAchievementReferenceByID(id)
	if err != nil {
		return err
	}
	if ref.Status == "deleted" || (ref.Status != "draft" && ref.Status != "rejected") {
		return fiber.NewError(http.StatusBadRequest, "cannot update this achievement")
	}

	objID, err := primitive.ObjectIDFromHex(ref.MongoAchievementID)
	if err != nil {
		return err
	}
	updatedAch.ID = objID
	updatedAch.StudentID = ref.StudentID
	return s.mongoRepo.UpdateAchievement(ref.MongoAchievementID, &updatedAch)
}

func (s *AchievementService) DeleteAchievement(id uuid.UUID, userID uuid.UUID) error {
	ref, err := s.postgresRepo.GetAchievementReferenceByID(id)
	if err != nil || ref.Status != "draft" {
		return fiber.NewError(http.StatusBadRequest, "can only delete draft achievement")
	}

	if err := s.mongoRepo.SoftDeleteAchievement(ref.MongoAchievementID); err != nil {
		return err
	}
	if err := s.postgresRepo.SoftDeleteAchievementReference(id); err != nil {
		return err
	}

	history := model.StatusHistory{Status: "deleted", ChangedBy: &userID, ChangedAt: time.Now(), Note: "Dihapus oleh mahasiswa"}
	_ = s.mongoRepo.AddStatusHistory(ref.MongoAchievementID, history) // ignore error
	return nil
}

func (s *AchievementService) SubmitAchievement(id uuid.UUID, userID uuid.UUID) error {
	ref, err := s.postgresRepo.GetAchievementReferenceByID(id)
	if err != nil || (ref.Status != "draft" && ref.Status != "rejected") {
		return fiber.NewError(http.StatusBadRequest, "only draft/rejected can be submitted")
	}

	if err := s.postgresRepo.SubmitAchievement(id); err != nil {
		return err
	}

	history := model.StatusHistory{Status: "submitted", ChangedBy: &userID, ChangedAt: time.Now(), Note: "Disubmit untuk verifikasi"}
	_ = s.mongoRepo.AddStatusHistory(ref.MongoAchievementID, history)
	return nil
}

func (s *AchievementService) VerifyAchievement(id uuid.UUID, verifiedBy uuid.UUID) error {
	ref, err := s.postgresRepo.GetAchievementReferenceByID(id)
	if err != nil || ref.Status != "submitted" {
		return fiber.NewError(http.StatusBadRequest, "only submitted can be verified")
	}

	if err := s.postgresRepo.VerifyAchievement(id, verifiedBy, nil); err != nil {
		return err
	}

	history := model.StatusHistory{Status: "verified", ChangedBy: &verifiedBy, ChangedAt: time.Now(), Note: "Diverifikasi"}
	_ = s.mongoRepo.AddStatusHistory(ref.MongoAchievementID, history)

	ach, _ := s.mongoRepo.GetAchievementByID(ref.MongoAchievementID)
	title := "Prestasi Anda"
	if ach != nil && ach.Title != "" {
		title = ach.Title
	}
	notif := model.Notification{Type: "achievement_verified", Title: "Disetujui", Message: title + " disetujui", Read: false, CreatedAt: time.Now()}
	_ = s.mongoRepo.AddNotification(ref.MongoAchievementID, notif)
	return nil
}

func (s *AchievementService) RejectAchievement(id uuid.UUID, verifiedBy uuid.UUID, note string) error {
	ref, err := s.postgresRepo.GetAchievementReferenceByID(id)
	if err != nil || ref.Status != "submitted" {
		return fiber.NewError(http.StatusBadRequest, "only submitted can be rejected")
	}

	if err := s.postgresRepo.VerifyAchievement(id, verifiedBy, &note); err != nil {
		return err
	}

	history := model.StatusHistory{Status: "rejected", ChangedBy: &verifiedBy, ChangedAt: time.Now(), Note: "Ditolak: " + note}
	_ = s.mongoRepo.AddStatusHistory(ref.MongoAchievementID, history)

	ach, _ := s.mongoRepo.GetAchievementByID(ref.MongoAchievementID)
	title := "Prestasi Anda"
	if ach != nil && ach.Title != "" {
		title = ach.Title
	}
	notif := model.Notification{Type: "achievement_rejected", Title: "Ditolak", Message: title + " ditolak: " + note, Read: false, CreatedAt: time.Now()}
	_ = s.mongoRepo.AddNotification(ref.MongoAchievementID, notif)
	return nil
}

func (s *AchievementService) GetAchievementHistory(id uuid.UUID) ([]model.StatusHistory, error) {
	ref, err := s.postgresRepo.GetAchievementReferenceByID(id)
	if err != nil {
		return nil, err
	}
	ach, err := s.mongoRepo.GetAchievementByID(ref.MongoAchievementID)
	if err != nil || ach == nil {
		return nil, fiber.NewError(http.StatusNotFound, "achievement not found")
	}
	return ach.StatusHistory, nil
}

func (s *AchievementService) UploadAttachment(id uuid.UUID, file io.Reader, fileName, fileType string) (*model.Attachment, error) {
	ref, err := s.postgresRepo.GetAchievementReferenceByID(id)
	if err != nil || ref.Status == "deleted" {
		return nil, fiber.NewError(http.StatusBadRequest, "cannot upload to this achievement")
	}
	return s.mongoRepo.UploadAttachment(ref.MongoAchievementID, file, fileName, fileType)
}

// ==================== HANDLERS WITH SWAGGER ====================

// @Summary List achievements (filtered by role)
// @Description Mengambil daftar prestasi berdasarkan role user (mahasiswa: own, dosen: advisees, admin: all), dengan filter status dan pagination
// @Tags Achievements
// @Accept json
// @Produce json
// @Param page query int false "Page number (default 1)"
// @Param limit query int false "Items per page (default 10)"
// @Param status query string false "Filter status (draft, submitted, verified, rejected, deleted)"
// @Success 200 {object} model.PaginatedResponse[model.AchievementReference]
// @Failure 400 {object} model.ErrorResponse "Invalid role"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /achievements [get]
func (s *AchievementService) ListHandler(c *fiber.Ctx) error {
	userIDStr := c.Locals("user_id").(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	role := c.Locals("role").(string)

	page, _ := strconv.Atoi(c.Query("page", "1"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	if limit < 1 || limit > 100 {
		limit = 10
	}

	status := c.Query("status")
	var statusPtr *string
	if status != "" {
		statusPtr = &status
	}

	resp, err := s.GetUserAchievements(userID, role, statusPtr, page, limit)
	if err != nil {
		if fe, ok := err.(*fiber.Error); ok {
			return c.Status(fe.Code).JSON(fiber.Map{"error": fe.Message})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(resp)
}

// @Summary Get achievement detail
// @Description Mengambil detail lengkap prestasi berdasarkan ID, termasuk history status
// @Tags Achievements
// @Accept json
// @Produce json
// @Param id path string true "Achievement ID (UUID)"
// @Success 200 {object} model.AchievementDetailResponse
// @Failure 404 {object} model.ErrorResponse "Achievement not found or deleted"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /achievements/{id} [get]
func (s *AchievementService) DetailHandler(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid achievement ID"})
	}

	resp, err := s.GetAchievementDetail(id)
	if err != nil {
		if fe, ok := err.(*fiber.Error); ok {
			return c.Status(fe.Code).JSON(fiber.Map{"error": fe.Message})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(resp)
}

// @Summary Create achievement
// @Description Membuat prestasi baru untuk mahasiswa (draft status)
// @Tags Achievements
// @Accept json
// @Produce json
// @Param achievement body model.Achievement true "Achievement data"
// @Success 201 {object} model.AchievementReference
// @Failure 400 {object} model.ErrorResponse "Student not found"
// @Failure 500 {object} model.ErrorResponse "Failed to create"
// @Security ApiKeyAuth
// @Router /achievements [post]
func (s *AchievementService) CreateHandler(c *fiber.Ctx) error {
	userIDStr := c.Locals("user_id").(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	var ach model.Achievement
	if err := c.BodyParser(&ach); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	ref, err := s.CreateAchievement(userID, ach)
	if err != nil {
		if fe, ok := err.(*fiber.Error); ok {
			return c.Status(fe.Code).JSON(fiber.Map{"error": fe.Message})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(http.StatusCreated).JSON(ref)
}

// @Summary Update achievement
// @Description Memperbarui prestasi (hanya untuk draft atau rejected status)
// @Tags Achievements
// @Accept json
// @Produce json
// @Param id path string true "Achievement ID (UUID)"
// @Param achievement body model.Achievement true "Updated achievement data"
// @Success 200 {object} map[string]string "message: Updated successfully"
// @Failure 400 {object} model.ErrorResponse "Cannot update this achievement"
// @Failure 500 {object} model.ErrorResponse "Failed to update"
// @Security ApiKeyAuth
// @Router /achievements/{id} [put]
func (s *AchievementService) UpdateHandler(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid achievement ID"})
	}

	var updatedAch model.Achievement
	if err := c.BodyParser(&updatedAch); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	err = s.UpdateAchievement(id, updatedAch)
	if err != nil {
		if fe, ok := err.(*fiber.Error); ok {
			return c.Status(fe.Code).JSON(fiber.Map{"error": fe.Message})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Updated successfully"})
}

// @Summary Delete achievement
// @Description Soft delete prestasi (hanya untuk draft status, oleh mahasiswa)
// @Tags Achievements
// @Accept json
// @Produce json
// @Param id path string true "Achievement ID (UUID)"
// @Success 200 {object} map[string]string "message: Deleted successfully"
// @Failure 400 {object} model.ErrorResponse "Can only delete draft achievement"
// @Failure 500 {object} model.ErrorResponse "Failed to delete"
// @Security ApiKeyAuth
// @Router /achievements/{id} [delete]
func (s *AchievementService) DeleteHandler(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid achievement ID"})
	}

	userIDStr := c.Locals("user_id").(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	err = s.DeleteAchievement(id, userID)
	if err != nil {
		if fe, ok := err.(*fiber.Error); ok {
			return c.Status(fe.Code).JSON(fiber.Map{"error": fe.Message})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Deleted successfully"})
}

// @Summary Submit achievement
// @Description Submit prestasi untuk verifikasi (dari draft atau rejected)
// @Tags Achievements
// @Accept json
// @Produce json
// @Param id path string true "Achievement ID (UUID)"
// @Success 200 {object} map[string]string "status: submitted"
// @Failure 400 {object} model.ErrorResponse "Only draft/rejected can be submitted"
// @Failure 500 {object} model.ErrorResponse "Failed to submit"
// @Security ApiKeyAuth
// @Router /achievements/{id}/submit [post]
func (s *AchievementService) SubmitHandler(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid achievement ID"})
	}

	userIDStr := c.Locals("user_id").(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	err = s.SubmitAchievement(id, userID)
	if err != nil {
		if fe, ok := err.(*fiber.Error); ok {
			return c.Status(fe.Code).JSON(fiber.Map{"error": fe.Message})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": "submitted"})
}

// @Summary Verify achievement
// @Description Verifikasi prestasi (oleh dosen/admin, dari submitted status)
// @Tags Achievements
// @Accept json
// @Produce json
// @Param id path string true "Achievement ID (UUID)"
// @Success 200 {object} map[string]string "status: verified"
// @Failure 400 {object} model.ErrorResponse "Only submitted can be verified"
// @Failure 500 {object} model.ErrorResponse "Failed to verify"
// @Security ApiKeyAuth
// @Router /achievements/{id}/verify [post]
func (s *AchievementService) VerifyHandler(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid achievement ID"})
	}

	userIDStr := c.Locals("user_id").(string)
	verifiedBy, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	err = s.VerifyAchievement(id, verifiedBy)
	if err != nil {
		if fe, ok := err.(*fiber.Error); ok {
			return c.Status(fe.Code).JSON(fiber.Map{"error": fe.Message})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": "verified"})
}

// @Summary Reject achievement
// @Description Tolak prestasi dengan note (oleh dosen/admin, dari submitted status)
// @Tags Achievements
// @Accept json
// @Produce json
// @Param id path string true "Achievement ID (UUID)"
// @Param rejection_note body object true "Rejection note" schema={"type":"object","properties":{"rejection_note":{"type":"string"}}}
// @Success 200 {object} map[string]string "status: rejected"
// @Failure 400 {object} model.ErrorResponse "Note required or wrong status"
// @Failure 500 {object} model.ErrorResponse "Failed to reject"
// @Security ApiKeyAuth
// @Router /achievements/{id}/reject [post]
func (s *AchievementService) RejectHandler(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid achievement ID"})
	}

	type Req struct {
		RejectionNote string `json:"rejection_note" validate:"required"`
	}
	var req Req
	if err := c.BodyParser(&req); err != nil || req.RejectionNote == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Rejection note is required"})
	}

	userIDStr := c.Locals("user_id").(string)
	verifiedBy, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	err = s.RejectAchievement(id, verifiedBy, req.RejectionNote)
	if err != nil {
		if fe, ok := err.(*fiber.Error); ok {
			return c.Status(fe.Code).JSON(fiber.Map{"error": fe.Message})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": "rejected"})
}

// @Summary Get achievement history
// @Description Mengambil riwayat status prestasi
// @Tags Achievements
// @Accept json
// @Produce json
// @Param id path string true "Achievement ID (UUID)"
// @Success 200 {array} model.StatusHistory
// @Failure 404 {object} model.ErrorResponse "Achievement not found"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /achievements/{id}/history [get]
func (s *AchievementService) HistoryHandler(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid achievement ID"})
	}

	histories, err := s.GetAchievementHistory(id)
	if err != nil {
		if fe, ok := err.(*fiber.Error); ok {
			return c.Status(fe.Code).JSON(fiber.Map{"error": fe.Message})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(histories)
}

// @Summary Upload attachment
// @Description Upload file attachment ke prestasi (hanya untuk non-deleted)
// @Tags Achievements
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "Achievement ID (UUID)"
// @Param file formData file true "Attachment file"
// @Success 200 {object} model.Attachment
// @Failure 400 {object} model.ErrorResponse "No file or invalid achievement"
// @Failure 500 {object} model.ErrorResponse "Failed to upload"
// @Security ApiKeyAuth
// @Router /achievements/{id}/attachments [post]
func (s *AchievementService) UploadAttachmentHandler(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid achievement ID"})
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

	attachment, err := s.UploadAttachment(id, src, fileName, file.Header.Get("Content-Type"))
	if err != nil {
		if fe, ok := err.(*fiber.Error); ok {
			return c.Status(fe.Code).JSON(fiber.Map{"error": fe.Message})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(attachment)
}
