// File: BACKEND-UAS/pgmongo/repository/report_repository.go
package repository

import (
	"context"
	"database/sql"
	"sort"
	"time"

	"github.com/google/uuid"
	"BACKEND-UAS/pgmongo/model"
)

type ReportRepository interface {
	GetAchievementStatistics(ctx context.Context, userID uuid.UUID) (*model.AchievementStatistics, error)
	GetStudentAchievementStatistics(ctx context.Context, studentID, userID uuid.UUID) (*model.StudentAchievementStatistics, error)
	GetTotalPointsForStudent(ctx context.Context, studentID uuid.UUID) (int64, error)
}

type reportRepository struct {
	db                   *sql.DB
	achievementMongoRepo *AchievementRepositoryMongo
}

func NewReportRepository(db *sql.DB, mongoRepo *AchievementRepositoryMongo) ReportRepository {
	return &reportRepository{
		db:                   db,
		achievementMongoRepo: mongoRepo,
	}
}

// GetAchievementStatistics fetches global stats
func (r *reportRepository) GetAchievementStatistics(ctx context.Context, userID uuid.UUID) (*model.AchievementStatistics, error) {
	stats := &model.AchievementStatistics{
		TotalPerType:   make(map[string]int64),
		TotalPerPeriod: make(map[string]int64),
		TopStudents:    []model.TopStudent{},
		Distribution:   make(map[string]int64),
	}

	// Get all verified mongo IDs from Postgres
	var verifiedIDs []string
	query := `SELECT mongo_achievement_id FROM achievement_references WHERE status = 'verified'`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		verifiedIDs = append(verifiedIDs, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(verifiedIDs) == 0 {
		return stats, nil
	}

	// Fetch achievements from Mongo one by one (since no batch method, loop)
	var achievements []*model.Achievement
	for _, id := range verifiedIDs {
		ach, err := r.achievementMongoRepo.GetAchievementByID(id)
		if err != nil || ach == nil {
			continue
		}
		achievements = append(achievements, ach)
	}

	// Aggregate in Go
	typeCounts := make(map[string]int64)
	periodCounts := make(map[string]int64)
	levelCounts := make(map[string]int64)
	now := time.Now()
	oneYearAgo := now.AddDate(-1, 0, 0)
	for _, ach := range achievements {
		if ach.CreatedAt.Before(oneYearAgo) {
			continue
		}
		typeCounts[ach.AchievementType]++
		periodKey := ach.CreatedAt.Format("2006-01")
		periodCounts[periodKey]++
		levelKey := ach.Level
		if levelKey == "" {
			levelKey = "unknown"
		}
		levelCounts[levelKey]++
	}

	stats.TotalPerType = typeCounts
	stats.TotalPerPeriod = periodCounts
	stats.Distribution = levelCounts

	// Top students: Fetch refs with student info
	queryTop := `
		SELECT ar.student_id, u.full_name, ar.mongo_achievement_id
		FROM achievement_references ar
		JOIN students s ON ar.student_id = s.id
		JOIN users u ON s.user_id = u.id
		WHERE ar.status = 'verified'
	`
	rowsTop, err := r.db.QueryContext(ctx, queryTop)
	if err != nil {
		return nil, err
	}
	defer rowsTop.Close()

	studentAchMap := make(map[uuid.UUID][]string) // studentID -> list mongoIDs
	for rowsTop.Next() {
		var sidStr, name, mid string
		if err := rowsTop.Scan(&sidStr, &name, &mid); err != nil {
			continue
		}
		sid, _ := uuid.Parse(sidStr)
		studentAchMap[sid] = append(studentAchMap[sid], mid)
	}

	// Calculate points per student
	type studentData struct {
		points int64
		count  int64
		name   string
	}
	studentMap := make(map[uuid.UUID]studentData)
	for sid, mids := range studentAchMap {
		var totalPoints int64
		var achCount int64
		for _, mid := range mids {
			ach, err := r.achievementMongoRepo.GetAchievementByID(mid)
			if err != nil || ach == nil {
				continue
			}
			totalPoints += int64(ach.Points)
			achCount++
		}
		var fullName string
		nameQuery := `SELECT u.full_name FROM students s JOIN users u ON s.user_id = u.id WHERE s.id = $1`
		if err := r.db.QueryRowContext(ctx, nameQuery, sid.String()).Scan(&fullName); err != nil {
			fullName = "Unknown"
		}
		studentMap[sid] = studentData{points: totalPoints, count: achCount, name: fullName}
	}

	// Convert to top students
	var topStudents []model.TopStudent
	for sid, data := range studentMap {
		topStudents = append(topStudents, model.TopStudent{
			StudentID: sid.String(),
			FullName:  data.name,
			Points:    data.points,
			Count:     data.count,
		})
	}

	// Sort by points DESC, then count DESC, limit 10
	sort.Slice(topStudents, func(i, j int) bool {
		if topStudents[i].Points != topStudents[j].Points {
			return topStudents[i].Points > topStudents[j].Points
		}
		return topStudents[i].Count > topStudents[j].Count
	})
	if len(topStudents) > 10 {
		topStudents = topStudents[:10]
	}
	stats.TopStudents = topStudents

	return stats, nil
}

// GetStudentAchievementStatistics similar logic but filter by studentID
func (r *reportRepository) GetStudentAchievementStatistics(ctx context.Context, studentID, userID uuid.UUID) (*model.StudentAchievementStatistics, error) {
	stats := &model.StudentAchievementStatistics{
		PerType:      make(map[string]int64),
		PerPeriod:    make(map[string]int64),
		Distribution: make(map[string]int64),
	}

	// Get verified mongo IDs for student
	var verifiedIDs []string
	query := `SELECT mongo_achievement_id FROM achievement_references WHERE student_id = $1 AND status = 'verified'`
	rows, err := r.db.QueryContext(ctx, query, studentID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		verifiedIDs = append(verifiedIDs, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(verifiedIDs) == 0 {
		return stats, nil
	}

	// Fetch achievements
	var achievements []*model.Achievement
	for _, id := range verifiedIDs {
		ach, err := r.achievementMongoRepo.GetAchievementByID(id)
		if err != nil || ach == nil {
			continue
		}
		achievements = append(achievements, ach)
	}

	// Aggregate
	now := time.Now()
	oneYearAgo := now.AddDate(-1, 0, 0)
	for _, ach := range achievements {
		if ach.CreatedAt.Before(oneYearAgo) {
			continue
		}
		stats.PerType[ach.AchievementType]++
		periodKey := ach.CreatedAt.Format("2006-01")
		stats.PerPeriod[periodKey]++
		levelKey := ach.Level
		if levelKey == "" {
			levelKey = "unknown"
		}
		stats.Distribution[levelKey]++
	}

	stats.TotalAchievements = int64(len(achievements))

	return stats, nil
}

// GetTotalPointsForStudent
func (r *reportRepository) GetTotalPointsForStudent(ctx context.Context, studentID uuid.UUID) (int64, error) {
	var ids []string
	query := `SELECT mongo_achievement_id FROM achievement_references WHERE student_id = $1 AND status = 'verified'`
	rows, err := r.db.QueryContext(ctx, query, studentID.String())
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			continue
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}

	if len(ids) == 0 {
		return 0, nil
	}

	var total int64
	for _, id := range ids {
		ach, err := r.achievementMongoRepo.GetAchievementByID(id)
		if err != nil || ach == nil {
			continue
		}
		total += int64(ach.Points)
	}
	return total, nil
}