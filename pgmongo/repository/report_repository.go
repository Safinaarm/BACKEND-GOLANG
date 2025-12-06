// File: BACKEND-UAS/pgmongo/repository/report_repository.go
package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"BACKEND-UAS/pgmongo/model"
)

type ReportRepository interface {
	GetAchievementStatistics(ctx context.Context, userID uuid.UUID) (*model.AchievementStatistics, error)
	GetStudentAchievementStatistics(ctx context.Context, studentID, userID uuid.UUID) (*model.StudentAchievementStatistics, error)
}

type reportRepository struct {
	db *sql.DB
}

func NewReportRepository(db *sql.DB) ReportRepository {
	return &reportRepository{db: db}
}

// GetAchievementStatistics fetches global stats (all achievements; service filters by role)
func (r *reportRepository) GetAchievementStatistics(ctx context.Context, userID uuid.UUID) (*model.AchievementStatistics, error) {
	stats := &model.AchievementStatistics{
		TotalPerType:   make(map[string]int64),
		TotalPerPeriod: make(map[string]int64),
		TopStudents:    []model.TopStudent{},
		Distribution:   make(map[string]int64),
	}

	// Total per type
	queryType := `
		SELECT achievement_type, COUNT(*) as count
		FROM achievements
		GROUP BY achievement_type
	`
	rowsType, err := r.db.QueryContext(ctx, queryType)
	if err != nil {
		return nil, err
	}
	defer rowsType.Close()
	for rowsType.Next() {
		var typ string
		var count int64
		if err := rowsType.Scan(&typ, &count); err != nil {
			return nil, err
		}
		stats.TotalPerType[typ] = count
	}

	// Total per period (monthly, last 12 months)
	now := time.Now()
	oneYearAgo := now.AddDate(-1, 0, 0)
	queryPeriod := `
		SELECT DATE_TRUNC('month', created_at) as period, COUNT(*) as count
		FROM achievements
		WHERE created_at >= $1
		GROUP BY period
		ORDER BY period
	`
	rowsPeriod, err := r.db.QueryContext(ctx, queryPeriod, oneYearAgo)
	if err != nil {
		return nil, err
	}
	defer rowsPeriod.Close()
	for rowsPeriod.Next() {
		var period time.Time
		var count int64
		if err := rowsPeriod.Scan(&period, &count); err != nil {
			return nil, err
		}
		stats.TotalPerPeriod[period.Format("2006-01")] = count
	}

	// Top students (top 10 by points sum)
	queryTop := `
		SELECT s.student_id, u.full_name, COALESCE(SUM(a.points), 0) as points, COUNT(a.id) as count
		FROM students s
		JOIN users u ON s.user_id = u.id
		LEFT JOIN achievements a ON s.id = a.student_id
		GROUP BY s.id, s.student_id, u.full_name
		ORDER BY points DESC NULLS LAST, count DESC NULLS LAST
		LIMIT 10
	`
	rowsTop, err := r.db.QueryContext(ctx, queryTop)
	if err != nil {
		return nil, err
	}
	defer rowsTop.Close()
	for rowsTop.Next() {
		var ts model.TopStudent
		if err := rowsTop.Scan(&ts.StudentID, &ts.FullName, &ts.Points, &ts.Count); err != nil {
			return nil, err
		}
		stats.TopStudents = append(stats.TopStudents, ts)
	}

	// Distribution by level (assume 'level' field in achievements)
	queryDist := `
		SELECT COALESCE(level, 'unknown') as level, COUNT(*) as count
		FROM achievements
		GROUP BY level
	`
	rowsDist, err := r.db.QueryContext(ctx, queryDist)
	if err != nil {
		return nil, err
	}
	defer rowsDist.Close()
	for rowsDist.Next() {
		var level string
		var count int64
		if err := rowsDist.Scan(&level, &count); err != nil {
			return nil, err
		}
		stats.Distribution[level] = count
	}

	return stats, nil
}

// GetStudentAchievementStatistics fetches stats for specific student
func (r *reportRepository) GetStudentAchievementStatistics(ctx context.Context, studentID, userID uuid.UUID) (*model.StudentAchievementStatistics, error) {
	stats := &model.StudentAchievementStatistics{
		PerType:      make(map[string]int64),
		PerPeriod:    make(map[string]int64),
		Distribution: make(map[string]int64),
	}

	// Total achievements
	queryTotal := `SELECT COUNT(*) FROM achievements WHERE student_id = $1`
	var total int64
	err := r.db.QueryRowContext(ctx, queryTotal, studentID).Scan(&total)
	if err != nil {
		return nil, err
	}
	stats.TotalAchievements = total

	// Per type
	queryType := `
		SELECT achievement_type, COUNT(*) as count
		FROM achievements
		WHERE student_id = $1
		GROUP BY achievement_type
	`
	rowsType, err := r.db.QueryContext(ctx, queryType, studentID)
	if err != nil {
		return nil, err
	}
	defer rowsType.Close()
	for rowsType.Next() {
		var typ string
		var count int64
		if err := rowsType.Scan(&typ, &count); err != nil {
			return nil, err
		}
		stats.PerType[typ] = count
	}

	// Per period (monthly, last 12 months)
	now := time.Now()
	oneYearAgo := now.AddDate(-1, 0, 0)
	queryPeriod := `
		SELECT DATE_TRUNC('month', created_at) as period, COUNT(*) as count
		FROM achievements
		WHERE student_id = $1 AND created_at >= $2
		GROUP BY period
		ORDER BY period
	`
	rowsPeriod, err := r.db.QueryContext(ctx, queryPeriod, studentID, oneYearAgo)
	if err != nil {
		return nil, err
	}
	defer rowsPeriod.Close()
	for rowsPeriod.Next() {
		var period time.Time
		var count int64
		if err := rowsPeriod.Scan(&period, &count); err != nil {
			return nil, err
		}
		stats.PerPeriod[period.Format("2006-01")] = count
	}

	// Distribution by level
	queryDist := `
		SELECT COALESCE(level, 'unknown') as level, COUNT(*) as count
		FROM achievements
		WHERE student_id = $1
		GROUP BY level
	`
	rowsDist, err := r.db.QueryContext(ctx, queryDist, studentID)
	if err != nil {
		return nil, err
	}
	defer rowsDist.Close()
	for rowsDist.Next() {
		var level string
		var count int64
		if err := rowsDist.Scan(&level, &count); err != nil {
			return nil, err
		}
		stats.Distribution[level] = count
	}

	return stats, nil
}