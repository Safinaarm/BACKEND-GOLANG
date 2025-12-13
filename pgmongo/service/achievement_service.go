// File: BACKEND-UAS/pgmongo/service/achievement_service.go
package service

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"BACKEND-UAS/pgmongo/model"
	"BACKEND-UAS/pgmongo/repository"
)

// AchievementService manages achievement operations across Postgres and MongoDB
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

// @Summary Get user achievements
// @Description Mengambil daftar prestasi user berdasarkan role (mahasiswa, dosen wali, admin), dengan filter status dan pagination
// @Tags Achievements
// @Accept json
// @Produce json
// @Param user_id path string true "User ID (UUID)"
// @Param role query string true "User role (Mahasiswa, Dosen Wali, Admin)"
// @Param status query string false "Filter status (draft, submitted, verified, rejected, deleted)"
// @Param page query int false "Page number (default 1)"
// @Param limit query int false "Items per page (default 10)"
// @Success 200 {array} model.AchievementReference
// @Failure 400 {object} model.ErrorResponse "Invalid role"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /achievements/user/{user_id} [get]
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
		return nil, errors.New("invalid role")
	}
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
// @Router /achievements/{id} [get]
func (s *AchievementService) GetAchievementDetail(id uuid.UUID) (*model.AchievementDetailResponse, error) {
    ref, err := s.postgresRepo.GetAchievementReferenceByID(id)
    if err != nil {
        return nil, fmt.Errorf("reference not found: %w", err)
    }
    if ref.Status == "deleted" {
        return nil, errors.New("achievement not found")
    }

    ach, err := s.mongoRepo.GetAchievementByID(ref.MongoAchievementID)
    if err != nil || ach == nil || ach.DeletedAt != nil {
        return nil, errors.New("achievement details not found or deleted")
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

// @Summary Create achievement
// @Description Membuat prestasi baru untuk mahasiswa (draft status)
// @Tags Achievements
// @Accept json
// @Produce json
// @Param user_id path string true "User ID (UUID) - must be student"
// @Param achievement body model.Achievement true "Achievement data"
// @Success 201 {object} model.AchievementReference
// @Failure 400 {object} model.ErrorResponse "Student not found"
// @Failure 500 {object} model.ErrorResponse "Failed to create in DB"
// @Router /achievements [post]
func (s *AchievementService) CreateAchievement(userID uuid.UUID, ach model.Achievement) (*model.AchievementReference, error) {
	student, err := s.postgresRepo.GetStudentByUserID(userID)
	if err != nil || student == nil {
		return nil, fmt.Errorf("student not found: %w", err)
	}

	ach.StudentID = student.ID
	if err := s.mongoRepo.CreateAchievement(&ach); err != nil {
		return nil, fmt.Errorf("failed to create in MongoDB: %w", err)
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
		return nil, fmt.Errorf("failed to create reference: %w", err)
	}
	return ref, nil
}

// @Summary Update achievement
// @Description Memperbarui prestasi (hanya untuk draft atau rejected status)
// @Tags Achievements
// @Accept json
// @Produce json
// @Param id path string true "Achievement ID (UUID)"
// @Param achievement body model.Achievement true "Updated achievement data"
// @Success 200 {object} nil
// @Failure 400 {object} model.ErrorResponse "Cannot update this achievement (wrong status)"
// @Failure 404 {object} model.ErrorResponse "Reference not found"
// @Failure 500 {object} model.ErrorResponse "Failed to update in DB"
// @Router /achievements/{id} [put]
func (s *AchievementService) UpdateAchievement(id uuid.UUID, updatedAch model.Achievement) error {
	ref, err := s.postgresRepo.GetAchievementReferenceByID(id)
	if err != nil {
		return fmt.Errorf("reference not found: %w", err)
	}
	if ref.Status == "deleted" || (ref.Status != "draft" && ref.Status != "rejected") {
		return errors.New("cannot update this achievement")
	}

	objID, err := primitive.ObjectIDFromHex(ref.MongoAchievementID)
	if err != nil {
		return fmt.Errorf("invalid mongo id: %w", err)
	}
	updatedAch.ID = objID
	updatedAch.StudentID = ref.StudentID
	return s.mongoRepo.UpdateAchievement(ref.MongoAchievementID, &updatedAch)
}

// @Summary Delete achievement
// @Description Soft delete prestasi (hanya untuk draft status, oleh mahasiswa)
// @Tags Achievements
// @Accept json
// @Produce json
// @Param id path string true "Achievement ID (UUID)"
// @Param user_id path string true "User ID (UUID) - must be student owner"
// @Success 200 {object} nil
// @Failure 400 {object} model.ErrorResponse "Can only delete draft achievement"
// @Failure 404 {object} model.ErrorResponse "Achievement not found"
// @Failure 500 {object} model.ErrorResponse "Failed to delete in DB"
// @Router /achievements/{id} [delete]
func (s *AchievementService) DeleteAchievement(id uuid.UUID, userID uuid.UUID) error {
	ref, err := s.postgresRepo.GetAchievementReferenceByID(id)
	if err != nil || ref.Status != "draft" {
		return errors.New("can only delete draft achievement")
	}

	if err := s.mongoRepo.SoftDeleteAchievement(ref.MongoAchievementID); err != nil {
		return fmt.Errorf("failed to soft delete: %w", err)
	}
	if err := s.postgresRepo.SoftDeleteAchievementReference(id); err != nil {
		return fmt.Errorf("failed to delete reference: %w", err)
	}

	history := model.StatusHistory{Status: "deleted", ChangedBy: &userID, ChangedAt: time.Now(), Note: "Dihapus oleh mahasiswa"}
	if err := s.mongoRepo.AddStatusHistory(ref.MongoAchievementID, history); err != nil {
		fmt.Printf("WARNING: Failed to save delete history: %v\n", err)
	}
	return nil
}

// @Summary Submit achievement
// @Description Submit prestasi untuk verifikasi (dari draft atau rejected)
// @Tags Achievements
// @Accept json
// @Produce json
// @Param id path string true "Achievement ID (UUID)"
// @Param user_id path string true "User ID (UUID) - submitter"
// @Success 200 {object} nil
// @Failure 400 {object} model.ErrorResponse "Only draft/rejected can be submitted"
// @Failure 404 {object} model.ErrorResponse "Achievement not found"
// @Failure 500 {object} model.ErrorResponse "Failed to submit"
// @Router /achievements/{id}/submit [patch]
func (s *AchievementService) SubmitAchievement(id uuid.UUID, userID uuid.UUID) error {
	ref, err := s.postgresRepo.GetAchievementReferenceByID(id)
	if err != nil || (ref.Status != "draft" && ref.Status != "rejected") {
		return errors.New("only draft/rejected can be submitted")
	}

	if err := s.postgresRepo.SubmitAchievement(id); err != nil {
		return fmt.Errorf("failed to submit: %w", err)
	}

	history := model.StatusHistory{Status: "submitted", ChangedBy: &userID, ChangedAt: time.Now(), Note: "Disubmit untuk verifikasi"}
	if err := s.mongoRepo.AddStatusHistory(ref.MongoAchievementID, history); err != nil {
		fmt.Printf("WARNING: Failed to save submit history: %v\n", err)
	}
	return nil
}

// @Summary Verify achievement
// @Description Verifikasi prestasi (oleh dosen/admin, dari submitted status)
// @Tags Achievements
// @Accept json
// @Produce json
// @Param id path string true "Achievement ID (UUID)"
// @Param verified_by path string true "Verifier User ID (UUID)"
// @Success 200 {object} nil
// @Failure 400 {object} model.ErrorResponse "Only submitted can be verified"
// @Failure 404 {object} model.ErrorResponse "Achievement not found"
// @Failure 500 {object} model.ErrorResponse "Failed to verify"
// @Router /achievements/{id}/verify [patch]
func (s *AchievementService) VerifyAchievement(id uuid.UUID, verifiedBy uuid.UUID) error {
	ref, err := s.postgresRepo.GetAchievementReferenceByID(id)
	if err != nil || ref.Status != "submitted" {
		return errors.New("only submitted can be verified")
	}

	if err := s.postgresRepo.VerifyAchievement(id, verifiedBy, nil); err != nil {
		return fmt.Errorf("failed to verify: %w", err)
	}

	history := model.StatusHistory{Status: "verified", ChangedBy: &verifiedBy, ChangedAt: time.Now(), Note: "Diverifikasi"}
	if err := s.mongoRepo.AddStatusHistory(ref.MongoAchievementID, history); err != nil {
		fmt.Printf("WARNING: Failed to save verify history: %v\n", err)
	}

	ach, _ := s.mongoRepo.GetAchievementByID(ref.MongoAchievementID)
	title := "Prestasi Anda"
	if ach != nil && ach.Title != "" {
		title = ach.Title
	}
	notif := model.Notification{Type: "achievement_verified", Title: "Disetujui", Message: fmt.Sprintf("%s disetujui", title), Read: false, CreatedAt: time.Now()}
	if err := s.mongoRepo.AddNotification(ref.MongoAchievementID, notif); err != nil {
		fmt.Printf("WARNING: Failed to send notification: %v\n", err)
	}
	return nil
}

// @Summary Reject achievement
// @Description Tolak prestasi dengan note (oleh dosen/admin, dari submitted status)
// @Tags Achievements
// @Accept json
// @Produce json
// @Param id path string true "Achievement ID (UUID)"
// @Param verified_by path string true "Rejector User ID (UUID)"
// @Param note body string true "Rejection note"
// @Success 200 {object} nil
// @Failure 400 {object} model.ErrorResponse "Only submitted can be rejected"
// @Failure 404 {object} model.ErrorResponse "Achievement not found"
// @Failure 500 {object} model.ErrorResponse "Failed to reject"
// @Router /achievements/{id}/reject [patch]
func (s *AchievementService) RejectAchievement(id uuid.UUID, verifiedBy uuid.UUID, note string) error {
	ref, err := s.postgresRepo.GetAchievementReferenceByID(id)
	if err != nil || ref.Status != "submitted" {
		return errors.New("only submitted can be rejected")
	}

	if err := s.postgresRepo.VerifyAchievement(id, verifiedBy, &note); err != nil {
		return fmt.Errorf("failed to reject: %w", err)
	}

	history := model.StatusHistory{Status: "rejected", ChangedBy: &verifiedBy, ChangedAt: time.Now(), Note: "Ditolak: " + note}
	if err := s.mongoRepo.AddStatusHistory(ref.MongoAchievementID, history); err != nil {
		fmt.Printf("WARNING: Failed to save reject history: %v\n", err)
	}

	ach, _ := s.mongoRepo.GetAchievementByID(ref.MongoAchievementID)
	title := "Prestasi Anda"
	if ach != nil && ach.Title != "" {
		title = ach.Title
	}
	notif := model.Notification{Type: "achievement_rejected", Title: "Ditolak", Message: fmt.Sprintf("%s ditolak: %s", title, note), Read: false, CreatedAt: time.Now()}
	if err := s.mongoRepo.AddNotification(ref.MongoAchievementID, notif); err != nil {
		fmt.Printf("WARNING: Failed to send notification: %v\n", err)
	}
	return nil
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
// @Router /achievements/{id}/history [get]
func (s *AchievementService) GetAchievementHistory(id uuid.UUID) ([]model.StatusHistory, error) {
	ref, err := s.postgresRepo.GetAchievementReferenceByID(id)
	if err != nil {
		return nil, fmt.Errorf("reference not found: %w", err)
	}
	ach, err := s.mongoRepo.GetAchievementByID(ref.MongoAchievementID)
	if err != nil || ach == nil {
		return nil, errors.New("achievement not found")
	}
	return ach.StatusHistory, nil
}

// @Summary Upload attachment
// @Description Upload file attachment ke prestasi (hanya untuk non-deleted)
// @Tags Achievements
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "Achievement ID (UUID)"
// @Param file formData file true "Attachment file"
// @Param file_name formData string true "Original file name"
// @Param file_type formData string true "MIME type"
// @Success 200 {object} model.Attachment
// @Failure 400 {object} model.ErrorResponse "Cannot upload to this achievement"
// @Failure 404 {object} model.ErrorResponse "Achievement not found"
// @Failure 500 {object} model.ErrorResponse "Failed to upload"
// @Router /achievements/{id}/attachments [post]
func (s *AchievementService) UploadAttachment(id uuid.UUID, file io.Reader, fileName, fileType string) (*model.Attachment, error) {
	ref, err := s.postgresRepo.GetAchievementReferenceByID(id)
	if err != nil || ref.Status == "deleted" {
		return nil, errors.New("cannot upload to this achievement")
	}
	return s.mongoRepo.UploadAttachment(ref.MongoAchievementID, file, fileName, fileType)
}