// File: BACKEND-UAS/pgmongo/service/lecturer_service.go
package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"BACKEND-UAS/pgmongo/model"
	"BACKEND-UAS/pgmongo/repository"
)

// LecturerService defines the interface for lecturer-related operations
type LecturerService interface {
	GetAllLecturers(ctx context.Context, userID uuid.UUID, page, limit int) (*model.PaginatedResponse[model.Lecturer], error)
	GetAdvisees(ctx context.Context, lecturerID, userID uuid.UUID) ([]model.Student, error)
}

type lecturerService struct {
	lecturerRepo repository.LecturerRepository
	studentRepo  *repository.StudentRepository
}

func NewLecturerService(lecturerRepo repository.LecturerRepository, studentRepo *repository.StudentRepository) LecturerService {
	return &lecturerService{
		lecturerRepo: lecturerRepo,
		studentRepo:  studentRepo,
	}
}

// @Summary Get all lecturers
// @Description Mengambil daftar dosen berdasarkan role user (lecturer: own data, student: advisor data, admin: all with pagination)
// @Tags Lecturers
// @Accept json
// @Produce json
// @Param user_id path string true "User ID (UUID)"
// @Param page query int false "Page number (default 1)"
// @Param limit query int false "Items per page (default 10)"
// @Success 200 {array} model.Lecturer
// @Failure 400 {object} model.ErrorResponse "No advisor assigned"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /lecturers [get]
func (s *lecturerService) GetAllLecturers(ctx context.Context, userID uuid.UUID, page, limit int) (*model.PaginatedResponse[model.Lecturer], error) {
	// Check if user is lecturer
	lecturer, err := s.lecturerRepo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}
	if lecturer != nil {
		data := []model.Lecturer{*lecturer}
		total := int64(1)
		totalPages := (int(total) + limit - 1) / limit
		return &model.PaginatedResponse[model.Lecturer]{
			Data:       data,
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		}, nil
	}

	// Check if user is student
	student, err := s.studentRepo.GetStudentByUserID(userID)
	if err != nil {
		return nil, err
	}
	if student != nil {
		if student.AdvisorID == uuid.Nil {
			return nil, fmt.Errorf("no advisor assigned")
		}
		data := []model.Lecturer{student.Advisor}
		total := int64(1)
		totalPages := (int(total) + limit - 1) / limit
		return &model.PaginatedResponse[model.Lecturer]{
			Data:       data,
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		}, nil
	}

	// Admin: return paginated all lecturers
	return s.lecturerRepo.GetAll(page, limit)
}

// @Summary Get lecturer's advisees
// @Description Mengambil daftar mahasiswa bimbingan dosen (dengan access check: own, advisor, or admin)
// @Tags Lecturers
// @Accept json
// @Produce json
// @Param lecturer_id path string true "Lecturer ID (UUID)"
// @Param user_id path string true "User ID (UUID) for access check"
// @Success 200 {array} model.Student
// @Failure 400 {object} model.ErrorResponse "Access denied"
// @Failure 404 {object} model.ErrorResponse "Lecturer not found"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /lecturers/{lecturer_id}/advisees [get]
func (s *lecturerService) GetAdvisees(ctx context.Context, lecturerID, userID uuid.UUID) ([]model.Student, error) {
	// Verify lecturer exists
	lecturer, err := s.lecturerRepo.GetByID(lecturerID)
	if err != nil {
		return nil, err
	}
	if lecturer == nil {
		return nil, fmt.Errorf("lecturer not found")
	}

	// Access checks
	// Check if user is student
	student, err := s.studentRepo.GetStudentByUserID(userID)
	if err != nil {
		return nil, err
	}
	if student != nil {
		if student.AdvisorID != lecturerID {
			return nil, fmt.Errorf("access denied")
		}
	}

	// Check if user is lecturer
	ownLecturer, err := s.lecturerRepo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}
	if ownLecturer != nil {
		if ownLecturer.ID != lecturerID {
			return nil, fmt.Errorf("access denied")
		}
	}

	// Admin or authorized: fetch advisees
	advisees, err := s.studentRepo.GetAdviseesByLecturerID(lecturerID)
	if err != nil {
		return nil, err
	}
	return advisees, nil
}