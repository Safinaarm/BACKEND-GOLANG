// File: BACKEND-UAS/pgmongo/service/student_service.go
package service

import (
	"fmt"

	"github.com/google/uuid"
	"BACKEND-UAS/pgmongo/model"
	"BACKEND-UAS/pgmongo/repository"
)

// StudentService manages student-related operations with role-based access
type StudentService struct {
	studentRepo *repository.StudentRepository
	achRepo     *repository.AchievementRepository
}

func NewStudentService(studentRepo *repository.StudentRepository, achRepo *repository.AchievementRepository) *StudentService {
	return &StudentService{
		studentRepo: studentRepo,
		achRepo:     achRepo,
	}
}

// @Summary Get all students
// @Description Mengambil daftar mahasiswa berdasarkan role user (student: own, lecturer: advisees, admin: all with pagination)
// @Tags Students
// @Accept json
// @Produce json
// @Param user_id path string true "User ID (UUID)"
// @Param page query int false "Page number (default 1)"
// @Param limit query int false "Items per page (default 10)"
// @Success 200 {object} model.PaginatedResponse{model.Student}
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /students [get]
func (s *StudentService) GetAllStudents(userID uuid.UUID, page, limit int) (*model.PaginatedResponse[model.Student], error) {
	var data []model.Student
	var total int64

	// Check if user is student
	student, err := s.studentRepo.GetStudentByUserID(userID)
	if err != nil {
		return nil, err
	}
	if student != nil {
		// Only own data
		data = []model.Student{*student}
		total = 1
	} else {
		// Check if user is lecturer
		lect, err := s.studentRepo.GetLecturerByUserID(userID)
		if err != nil {
			return nil, err
		}
		if lect != nil {
			// Get advisees
			advisees, err := s.studentRepo.GetAdviseesByLecturerID(lect.ID)
			if err != nil {
				return nil, err
			}
			data = advisees
			total = int64(len(data))
		} else {
			// Admin: all students
			paginated, err := s.studentRepo.GetAllStudents(page, limit)
			if err != nil {
				return nil, err
			}
			return paginated, nil
		}
	}

	totalPages := (int(total) + limit - 1) / limit
	return &model.PaginatedResponse[model.Student]{
		Data:       data,
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

// @Summary Get own student profile
// @Description Mengambil profil mahasiswa untuk user yang login (jika mahasiswa)
// @Tags Students
// @Accept json
// @Produce json
// @Param user_id path string true "User ID (UUID)"
// @Success 200 {object} model.Student
// @Failure 404 {object} model.ErrorResponse "No student profile found"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /students/me [get]
func (s *StudentService) GetOwnStudentProfile(userID uuid.UUID) (*model.Student, error) {
	student, err := s.studentRepo.GetStudentByUserID(userID)
	if err != nil {
		return nil, err
	}
	if student == nil {
		return nil, fmt.Errorf("no student profile found")
	}
	return student, nil
}

// @Summary Get student by ID
// @Description Mengambil detail mahasiswa berdasarkan ID, dengan access check (own, advisor, admin)
// @Tags Students
// @Accept json
// @Produce json
// @Param id path string true "Student ID (UUID)"
// @Param user_id path string true "User ID (UUID) for access check"
// @Success 200 {object} model.Student
// @Failure 400 {object} model.ErrorResponse "Access denied"
// @Failure 404 {object} model.ErrorResponse "Student not found"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /students/{id} [get]
func (s *StudentService) GetStudentByID(id uuid.UUID, userID uuid.UUID) (*model.Student, error) {
	student, err := s.studentRepo.GetStudentByID(id)
	if err != nil {
		return nil, err
	}
	if student == nil {
		return nil, nil
	}

	// Determine user type
	isStudent, err := s.isStudent(userID)
	if err != nil {
		return nil, err
	}
	if isStudent {
		if student.UserID != userID {
			return nil, fmt.Errorf("access denied")
		}
		return student, nil
	}

	isLecturer, err := s.isLecturer(userID)
	if err != nil {
		return nil, err
	}
	if isLecturer {
		lectID, err := s.getLecturerID(userID)
		if err != nil {
			return nil, err
		}
		if student.AdvisorID != lectID {
			return nil, fmt.Errorf("access denied")
		}
		return student, nil
	}

	// Admin: allowed
	return student, nil
}

// @Summary Get student achievements
// @Description Mengambil prestasi mahasiswa, dengan access check dan filter status/pagination
// @Tags Students
// @Accept json
// @Produce json
// @Param student_id path string true "Student ID (UUID)"
// @Param user_id path string true "User ID (UUID) for access check"
// @Param status query string false "Filter status (draft, submitted, verified, rejected, deleted)"
// @Param page query int false "Page number (default 1)"
// @Param limit query int false "Items per page (default 10)"
// @Success 200 {object} model.PaginatedResponse{model.AchievementReference}
// @Failure 400 {object} model.ErrorResponse "Access denied"
// @Failure 404 {object} model.ErrorResponse "Student not found"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /students/{student_id}/achievements [get]
func (s *StudentService) GetStudentAchievements(studentID uuid.UUID, userID uuid.UUID, status *string, page, limit int) (*model.PaginatedResponse[model.AchievementReference], error) {
	// First, check access to the student
	_, err := s.GetStudentByID(studentID, userID)
	if err != nil {
		return nil, err
	}

	return s.achRepo.GetAchievementReferencesByStudentIDs([]uuid.UUID{studentID}, status, page, limit)
}

// @Summary Update student advisor
// @Description Memperbarui dosen wali mahasiswa (hanya untuk admin)
// @Tags Students
// @Accept json
// @Produce json
// @Param student_id path string true "Student ID (UUID)"
// @Param advisor_id path string true "New Advisor ID (UUID)"
// @Param user_id path string true "User ID (UUID) - must be admin"
// @Success 200 {object} nil
// @Failure 400 {object} model.ErrorResponse "Unauthorized (only admin)"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /students/{student_id}/advisor [patch]
func (s *StudentService) UpdateStudentAdvisor(studentID, advisorID, userID uuid.UUID) error {
	isStudent, err := s.isStudent(userID)
	if err != nil {
		return err
	}
	if isStudent {
		return fmt.Errorf("unauthorized")
	}

	isLecturer, err := s.isLecturer(userID)
	if err != nil {
		return err
	}
	if isLecturer {
		return fmt.Errorf("unauthorized")
	}

	// Admin
	return s.studentRepo.UpdateStudentAdvisor(studentID, advisorID)
}

// Helper methods
func (s *StudentService) isStudent(userID uuid.UUID) (bool, error) {
	student, err := s.studentRepo.GetStudentByUserID(userID)
	if err != nil {
		return false, err
	}
	return student != nil, nil
}

func (s *StudentService) isLecturer(userID uuid.UUID) (bool, error) {
	lect, err := s.studentRepo.GetLecturerByUserID(userID)
	if err != nil {
		return false, err
	}
	return lect != nil, nil
}

func (s *StudentService) getLecturerID(userID uuid.UUID) (uuid.UUID, error) {
	lect, err := s.studentRepo.GetLecturerByUserID(userID)
	if err != nil {
		return uuid.Nil, err
	}
	if lect == nil {
		return uuid.Nil, fmt.Errorf("not a lecturer")
	}
	return lect.ID, nil
}