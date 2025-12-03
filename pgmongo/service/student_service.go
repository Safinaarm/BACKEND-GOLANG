// File: BACKEND-UAS/pgmongo/service/student_service.go
package service

import (
	"fmt"

	"github.com/google/uuid"
	"BACKEND-UAS/pgmongo/model"
	"BACKEND-UAS/pgmongo/repository"
)

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

// GetAllStudents retrieves a paginated list of students based on user type (determined from DB).
// If user is student, returns only their own data.
// If user is lecturer, returns their advisees.
// If neither (admin), returns all.
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

// GetOwnStudentProfile gets the student profile for the logged-in user (if student).
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

// GetStudentByID retrieves a single student by ID, with access check based on user type (from DB).
// Student can only view own, lecturer own advisees, admin all.
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

// GetStudentAchievements retrieves achievements for a student, with access check based on user type.
func (s *StudentService) GetStudentAchievements(studentID uuid.UUID, userID uuid.UUID, status *string, page, limit int) (*model.PaginatedResponse[model.AchievementReference], error) {
	// First, check access to the student
	_, err := s.GetStudentByID(studentID, userID)
	if err != nil {
		return nil, err
	}

	return s.achRepo.GetAchievementReferencesByStudentIDs([]uuid.UUID{studentID}, status, page, limit)
}

// UpdateStudentAdvisor updates advisor, only for admin (neither student nor lecturer).
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