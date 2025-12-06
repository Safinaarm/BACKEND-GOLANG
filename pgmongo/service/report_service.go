// File: BACKEND-UAS/pgmongo/service/report_service.go
package service

import (
	"context"
	"fmt"
	"sort"

	"github.com/google/uuid"

	"BACKEND-UAS/pgmongo/model"
	"BACKEND-UAS/pgmongo/repository"
)

type ReportService interface {
	GetAchievementStatistics(ctx context.Context, userID uuid.UUID) (*model.AchievementStatistics, error)
	GetStudentAchievementStatistics(ctx context.Context, studentID, userID uuid.UUID) (*model.StudentAchievementStatistics, error)
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
			StudentID: advisee.StudentID,
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