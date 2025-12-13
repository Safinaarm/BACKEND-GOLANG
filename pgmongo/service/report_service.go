// File: BACKEND-UAS/pgmongo/service/report_service.go
package service

import (
	"context"
	"fmt"
	"net/http"
	"sort"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"BACKEND-UAS/pgmongo/model"
	"BACKEND-UAS/pgmongo/repository"
)

// ReportService defines the interface for report and statistics operations
type ReportService interface {
	GetAchievementStatistics(ctx context.Context, userID uuid.UUID) (*model.AchievementStatistics, error)
	GetStudentAchievementStatistics(ctx context.Context, studentID, userID uuid.UUID) (*model.StudentAchievementStatistics, error)
	HandleGetAchievementStatistics() fiber.Handler
	HandleGetStudentAchievementStatistics() fiber.Handler
}

type reportService struct {
	reportRepo   repository.ReportRepository
	studentRepo  *repository.StudentRepository
	lecturerRepo repository.LecturerRepository
}

func NewReportService(reportRepo repository.ReportRepository, studentRepo *repository.StudentRepository, lecturerRepo repository.LecturerRepository) ReportService {
	return &reportService{
		reportRepo:   reportRepo,
		studentRepo:  studentRepo,
		lecturerRepo: lecturerRepo,
	}
}

// @Summary Get achievement statistics
// @Description Mengambil statistik prestasi berdasarkan role user (student: own, lecturer: advisees, admin: global)
// @Tags Reports
// @Accept json
// @Produce json
// @Success 200 {object} model.AchievementStatistics
// @Failure 403 {object} model.ErrorResponse "No profile found"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /reports/statistics [get]
func (s *reportService) GetAchievementStatistics(ctx context.Context, userID uuid.UUID) (*model.AchievementStatistics, error) {
	// Access check based on role
	isStudent, err := s.isStudent(userID)
	if err != nil {
		return nil, err
	}
	if isStudent {
		student, err := s.studentRepo.GetStudentByUserID(userID)
		if err != nil {
			return nil, err
		}
		if student == nil {
			return nil, fmt.Errorf("no student profile found")
		}
		// For student: Own stats
		return s.aggregateOwnStats(ctx, student.ID)
	}

	isLecturer, err := s.isLecturer(userID)
	if err != nil {
		return nil, err
	}
	if isLecturer {
		lecturer, err := s.lecturerRepo.GetByUserID(userID)
		if err != nil {
			return nil, err
		}
		if lecturer == nil {
			return nil, fmt.Errorf("no lecturer profile found")
		}
		// For lecturer: Advisees stats
		return s.aggregateAdviseesStats(ctx, lecturer.ID)
	}

	// Admin: Full global stats
	return s.reportRepo.GetAchievementStatistics(ctx, userID)
}

// @Summary Get student achievement statistics
// @Description Mengambil statistik prestasi mahasiswa spesifik (dengan access check: own, advisor, or admin)
// @Tags Reports
// @Accept json
// @Produce json
// @Param student_id path string true "Student ID (UUID)"
// @Success 200 {object} model.StudentAchievementStatistics
// @Failure 403 {object} model.ErrorResponse "Access denied"
// @Failure 404 {object} model.ErrorResponse "Student not found"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /reports/students/{student_id}/statistics [get]
func (s *reportService) GetStudentAchievementStatistics(ctx context.Context, studentID, userID uuid.UUID) (*model.StudentAchievementStatistics, error) {
	// Access check
	student, err := s.studentRepo.GetStudentByID(studentID)
	if err != nil {
		return nil, err
	}
	if student == nil {
		return nil, fmt.Errorf("student not found")
	}

	isStudent, err := s.isStudent(userID)
	if err != nil {
		return nil, err
	}
	if isStudent {
		if student.UserID != userID {
			return nil, fmt.Errorf("access denied")
		}
	}

	isLecturer, err := s.isLecturer(userID)
	if err != nil {
		return nil, err
	}
	if isLecturer {
		lecturerID, err := s.getLecturerID(userID)
		if err != nil {
			return nil, err
		}
		if student.AdvisorID != lecturerID {
			return nil, fmt.Errorf("access denied")
		}
	}

	// Authorized: Fetch student-specific stats
	return s.reportRepo.GetStudentAchievementStatistics(ctx, studentID, userID)
}

func (s *reportService) HandleGetAchievementStatistics() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userIDStr, ok := c.Locals("user_id").(string)
		if !ok {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
		}
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
		}

		stats, err := s.GetAchievementStatistics(c.Context(), userID)
		if err != nil {
			errMsg := err.Error()
			status := http.StatusInternalServerError
			if errMsg == "no student profile found" || errMsg == "no lecturer profile found" {
				status = http.StatusForbidden
			}
			return c.Status(status).JSON(fiber.Map{"error": errMsg})
		}

		return c.JSON(stats)
	}
}

func (s *reportService) HandleGetStudentAchievementStatistics() fiber.Handler {
	return func(c *fiber.Ctx) error {
		studentIDStr := c.Params("student_id")
		studentID, err := uuid.Parse(studentIDStr)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid student ID"})
		}

		userIDStr, ok := c.Locals("user_id").(string)
		if !ok {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
		}
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
		}

		stats, err := s.GetStudentAchievementStatistics(c.Context(), studentID, userID)
		if err != nil {
			errMsg := err.Error()
			status := http.StatusInternalServerError
			switch {
			case errMsg == "access denied":
				status = http.StatusForbidden
			case errMsg == "student not found":
				status = http.StatusNotFound
			}
			return c.Status(status).JSON(fiber.Map{"error": errMsg})
		}

		return c.JSON(stats)
	}
}

// Helper methods
func (s *reportService) isStudent(userID uuid.UUID) (bool, error) {
	student, err := s.studentRepo.GetStudentByUserID(userID)
	return student != nil, err
}

func (s *reportService) isLecturer(userID uuid.UUID) (bool, error) {
	lecturer, err := s.lecturerRepo.GetByUserID(userID)
	return lecturer != nil, err
}

func (s *reportService) getLecturerID(userID uuid.UUID) (uuid.UUID, error) {
	lecturer, err := s.lecturerRepo.GetByUserID(userID)
	if err != nil {
		return uuid.Nil, err
	}
	if lecturer == nil {
		return uuid.Nil, fmt.Errorf("not a lecturer")
	}
	return lecturer.ID, nil
}

// Aggregate own stats
func (s *reportService) aggregateOwnStats(ctx context.Context, studentID uuid.UUID) (*model.AchievementStatistics, error) {
	studentStats, err := s.reportRepo.GetStudentAchievementStatistics(ctx, studentID, uuid.Nil)
	if err != nil {
		return nil, err
	}

	totalPoints, err := s.reportRepo.GetTotalPointsForStudent(ctx, studentID)
	if err != nil {
		return nil, err
	}

	student, err := s.studentRepo.GetStudentByID(studentID)
	if err != nil {
		return nil, err
	}
	fullName := "Own Profile"
	if student != nil && student.User.FullName != "" {
		fullName = student.User.FullName
	}

	stats := &model.AchievementStatistics{
		TotalPerType:   studentStats.PerType,
		TotalPerPeriod: studentStats.PerPeriod,
		Distribution:   studentStats.Distribution,
		TopStudents: []model.TopStudent{
			{
				StudentID: studentID.String(),
				FullName:  fullName,
				Points:    totalPoints,
				Count:     studentStats.TotalAchievements,
			},
		},
	}

	return stats, nil
}

// Aggregate advisees stats
func (s *reportService) aggregateAdviseesStats(ctx context.Context, lecturerID uuid.UUID) (*model.AchievementStatistics, error) {
	advisees, err := s.studentRepo.GetAdviseesByLecturerID(lecturerID)
	if err != nil {
		return nil, err
	}

	var totalType map[string]int64 = make(map[string]int64)
	var totalPeriod map[string]int64 = make(map[string]int64)
	var totalDist map[string]int64 = make(map[string]int64)
	var topStudents []model.TopStudent

	for _, advisee := range advisees {
		studentStats, err := s.reportRepo.GetStudentAchievementStatistics(ctx, advisee.ID, uuid.Nil)
		if err != nil {
			continue
		}

		// Sum
		for typ, count := range studentStats.PerType {
			totalType[typ] += count
		}
		for period, count := range studentStats.PerPeriod {
			totalPeriod[period] += count
		}
		for level, count := range studentStats.Distribution {
			totalDist[level] += count
		}

		points, err := s.reportRepo.GetTotalPointsForStudent(ctx, advisee.ID)
		if err != nil {
			points = 0
		}

		topStudents = append(topStudents, model.TopStudent{
			StudentID: advisee.ID.String(),
			FullName:  advisee.User.FullName,
			Points:    points,
			Count:     studentStats.TotalAchievements,
		})
	}

	// Sort topStudents by points DESC, then count DESC
	sort.Slice(topStudents, func(i, j int) bool {
		if topStudents[i].Points != topStudents[j].Points {
			return topStudents[i].Points > topStudents[j].Points
		}
		return topStudents[i].Count > topStudents[j].Count
	})

	// Limit to top 10
	if len(topStudents) > 10 {
		topStudents = topStudents[:10]
	}

	stats := &model.AchievementStatistics{
		TotalPerType:   totalType,
		TotalPerPeriod: totalPeriod,
		TopStudents:    topStudents,
		Distribution:   totalDist,
	}

	return stats, nil
}