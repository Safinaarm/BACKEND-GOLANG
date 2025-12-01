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

// === SEMUA FUNGSI LAIN SUDAH AMAN ===

func (s *AchievementService) GetAchievementDetail(id uuid.UUID) (*model.AchievementResponse, error) {
	ref, err := s.postgresRepo.GetAchievementReferenceByID(id)
	if err != nil {
		return nil, fmt.Errorf("reference not found: %w", err)
	}
	if ref.Status == "deleted" {
		return nil, errors.New("achievement not found")
	}

	ach, err := s.mongoRepo.GetAchievementByID(ref.MongoAchievementID)
	if err != nil || ach == nil {
		return nil, errors.New("achievement details not found")
	}

	return &model.AchievementResponse{Reference: *ref, Achievement: *ach}, nil
}

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

func (s *AchievementService) UploadAttachment(id uuid.UUID, file io.Reader, fileName, fileType string) (*model.Attachment, error) {
	ref, err := s.postgresRepo.GetAchievementReferenceByID(id)
	if err != nil || ref.Status == "deleted" {
		return nil, errors.New("cannot upload to this achievement")
	}
	return s.mongoRepo.UploadAttachment(ref.MongoAchievementID, file, fileName, fileType)
}