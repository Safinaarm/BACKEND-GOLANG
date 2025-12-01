// File: BACKEND-UAS/pgmongo/repository/achievement_repository.go
package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq" // PASTIKAN IMPORT INI ADA
	"BACKEND-UAS/pgmongo/model"
)

type AchievementRepository struct {
	db *sql.DB
}

func NewAchievementRepository(db *sql.DB) *AchievementRepository {
	return &AchievementRepository{db: db}
}

func parseUUID(s string) uuid.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil
	}
	return id
}

// ===================== SCAN ROWS =====================
func (r *AchievementRepository) scanAchievementRows(rows *sql.Rows) ([]model.AchievementReference, error) {
	var refs []model.AchievementReference
	for rows.Next() {
		var ref model.AchievementReference
		var s model.Student

		var (
			arID             string
			arStudentID      string
			verifiedByStr    sql.NullString
			rejectionNote    sql.NullString
			submittedAt      sql.NullTime
			verifiedAt       sql.NullTime
			sID              string
			sUserID          string
		)

		err := rows.Scan(
			&arID, &arStudentID, &ref.MongoAchievementID, &ref.Status, &submittedAt, &verifiedAt,
			&verifiedByStr, &rejectionNote, &ref.CreatedAt, &ref.UpdatedAt,
			&sID, &sUserID, &s.StudentID, &s.ProgramStudy, &s.AcademicYear, &s.CreatedAt,
			&s.User.ID, &s.User.Username, &s.User.Email, &s.User.FullName, &s.User.RoleID,
			&s.User.IsActive, &s.User.CreatedAt, &s.User.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		ref.ID = parseUUID(arID)
		ref.StudentID = parseUUID(arStudentID)
		s.ID = parseUUID(sID)
		s.UserID = parseUUID(sUserID)
		ref.Student = s

		if submittedAt.Valid {
			ref.SubmittedAt = &submittedAt.Time
		}
		if verifiedAt.Valid {
			ref.VerifiedAt = &verifiedAt.Time
		}
		if verifiedByStr.Valid {
			vb := parseUUID(verifiedByStr.String)
			if vb != uuid.Nil {
				ref.VerifiedBy = &vb
			}
		}
		if rejectionNote.Valid {
			ref.RejectionNote = rejectionNote.String
		}

		refs = append(refs, ref)
	}
	return refs, nil
}

// ===================== LIST BY STUDENT IDS =====================
func (r *AchievementRepository) GetAchievementReferencesByStudentIDs(studentIDs []uuid.UUID, status *string, page, limit int) (*model.PaginatedResponse[model.AchievementReference], error) {
	if len(studentIDs) == 0 {
		return &model.PaginatedResponse[model.AchievementReference]{Data: []model.AchievementReference{}}, nil
	}

	// Convert ke string slice untuk pq.Array
	ids := make([]string, len(studentIDs))
	for i, id := range studentIDs {
		ids[i] = id.String()
	}

	// COUNT — Pakai pq.Array + ::uuid[]
	var total int64
	countQuery := `SELECT COUNT(*) FROM achievement_references WHERE student_id = ANY($1::uuid[]) AND status != 'deleted'`
	countArgs := []interface{}{pq.Array(ids)}
	if status != nil {
		countQuery += " AND status = $2"
		countArgs = append(countArgs, *status)
	}
	err := r.db.QueryRow(countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, err
	}

	// SELECT
	offset := (page - 1) * limit
	query := `
		SELECT ar.id, ar.student_id, ar.mongo_achievement_id, ar.status, ar.submitted_at, ar.verified_at, 
		       ar.verified_by, ar.rejection_note, ar.created_at, ar.updated_at,
		       s.id, s.user_id, s.student_id, s.program_study, s.academic_year, s.created_at,
		       u.id, u.username, u.email, u.full_name, u.role_id, u.is_active, u.created_at, u.updated_at
		FROM achievement_references ar
		JOIN students s ON ar.student_id = s.id
		JOIN users u ON s.user_id = u.id
		WHERE ar.student_id = ANY($1::uuid[]) AND ar.status != 'deleted'
	`
	args := []interface{}{pq.Array(ids)}

	if status != nil {
		query += " AND ar.status = $2"
		args = append(args, *status)
	}

	// Gunakan penomoran parameter manual agar aman
	if status != nil {
		query += fmt.Sprintf(" ORDER BY ar.created_at DESC LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
		args = append(args, limit, offset)
	} else {
		query += " ORDER BY ar.created_at DESC LIMIT $2 OFFSET $3"
		args = append(args, limit, offset)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	data, err := r.scanAchievementRows(rows)
	if err != nil {
		return nil, err
	}

	totalPages := (int(total) + limit - 1) / limit
	return &model.PaginatedResponse[model.AchievementReference]{
		Data:       data,
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

// ===================== LIST ALL (ADMIN) =====================
func (r *AchievementRepository) GetAllAchievementReferences(status *string, page, limit int) (*model.PaginatedResponse[model.AchievementReference], error) {
	var total int64
	countQuery := `SELECT COUNT(*) FROM achievement_references WHERE status != 'deleted'`
	countArgs := []interface{}{}
	if status != nil {
		countQuery += " AND status = $1"
		countArgs = append(countArgs, *status)
	}
	err := r.db.QueryRow(countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, err
	}

	offset := (page - 1) * limit
	query := `
		SELECT ar.id, ar.student_id, ar.mongo_achievement_id, ar.status, ar.submitted_at, ar.verified_at, 
		       ar.verified_by, ar.rejection_note, ar.created_at, ar.updated_at,
		       s.id, s.user_id, s.student_id, s.program_study, s.academic_year, s.created_at,
		       u.id, u.username, u.email, u.full_name, u.role_id, u.is_active, u.created_at, u.updated_at
		FROM achievement_references ar
		JOIN students s ON ar.student_id = s.id
		JOIN users u ON s.user_id = u.id
		WHERE ar.status != 'deleted'
	`
	args := []interface{}{}

	if status != nil {
		query += " AND ar.status = $1"
		args = append(args, *status)
	}

	if status != nil {
		query += fmt.Sprintf(" ORDER BY ar.created_at DESC LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
		args = append(args, limit, offset)
	} else {
		query += " ORDER BY ar.created_at DESC LIMIT $1 OFFSET $2"
		args = append(args, limit, offset)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	data, err := r.scanAchievementRows(rows)
	if err != nil {
		return nil, err
	}

	totalPages := (int(total) + limit - 1) / limit
	return &model.PaginatedResponse[model.AchievementReference]{
		Data:       data,
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}


// ===================== DETAIL =====================
func (r *AchievementRepository) GetAchievementReferenceByID(id uuid.UUID) (*model.AchievementReference, error) {
	query := `
		SELECT ar.id, ar.student_id, ar.mongo_achievement_id, ar.status, ar.submitted_at, ar.verified_at, 
		       ar.verified_by, ar.rejection_note, ar.created_at, ar.updated_at,
		       s.id, s.user_id, s.student_id, s.program_study, s.academic_year, s.created_at,
		       u.id, u.username, u.email, u.full_name, u.role_id, u.is_active, u.created_at, u.updated_at
		FROM achievement_references ar
		JOIN students s ON ar.student_id = s.id
		JOIN users u ON s.user_id = u.id
		WHERE ar.id = $1
	`
	row := r.db.QueryRow(query, id.String())

	var ref model.AchievementReference
	var s model.Student
	var (
		arID          string
		arStudentID   string
		verifiedByStr sql.NullString
		rejectionNote sql.NullString
		submittedAt   sql.NullTime
		verifiedAt    sql.NullTime
		sID           string
		sUserID       string
	)

	err := row.Scan(
		&arID, &arStudentID, &ref.MongoAchievementID, &ref.Status, &submittedAt, &verifiedAt,
		&verifiedByStr, &rejectionNote, &ref.CreatedAt, &ref.UpdatedAt,
		&sID, &sUserID, &s.StudentID, &s.ProgramStudy, &s.AcademicYear, &s.CreatedAt,
		&s.User.ID, &s.User.Username, &s.User.Email, &s.User.FullName, &s.User.RoleID,
		&s.User.IsActive, &s.User.CreatedAt, &s.User.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	ref.ID = parseUUID(arID)
	ref.StudentID = parseUUID(arStudentID)
	s.ID = parseUUID(sID)
	s.UserID = parseUUID(sUserID)
	ref.Student = s

	if submittedAt.Valid {
		ref.SubmittedAt = &submittedAt.Time
	}
	if verifiedAt.Valid {
		ref.VerifiedAt = &verifiedAt.Time
	}
	if verifiedByStr.Valid {
		vb := parseUUID(verifiedByStr.String)
		if vb != uuid.Nil {
			ref.VerifiedBy = &vb
		}
	}
	if rejectionNote.Valid {
		ref.RejectionNote = rejectionNote.String
	}

	return &ref, nil
}

// ===================== CRUD =====================
func (r *AchievementRepository) CreateAchievementReference(ref *model.AchievementReference) error {
	_, err := r.db.Exec(`
		INSERT INTO achievement_references (id, student_id, mongo_achievement_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		ref.ID.String(), ref.StudentID.String(), ref.MongoAchievementID, ref.Status, ref.CreatedAt, ref.UpdatedAt)
	return err
}

func (r *AchievementRepository) SoftDeleteAchievementReference(id uuid.UUID) error {
	_, err := r.db.Exec(`UPDATE achievement_references SET status = 'deleted', updated_at = NOW() WHERE id = $1`, id.String())
	return err
}

func (r *AchievementRepository) SubmitAchievement(id uuid.UUID) error {
	_, err := r.db.Exec(`UPDATE achievement_references SET status = 'submitted', submitted_at = NOW(), updated_at = NOW() WHERE id = $1`, id.String())
	return err
}

func (r *AchievementRepository) VerifyAchievement(id uuid.UUID, verifiedBy uuid.UUID, rejectionNote *string) error {
	now := time.Now()
	if rejectionNote != nil && *rejectionNote != "" {
		_, err := r.db.Exec(`
			UPDATE achievement_references 
			SET status = 'rejected', rejection_note = $1, verified_at = $2, updated_at = $3 
			WHERE id = $4`,
			*rejectionNote, now, now, id.String())
		return err
	}
	_, err := r.db.Exec(`
		UPDATE achievement_references 
		SET status = 'verified', verified_by = $1, verified_at = $2, updated_at = $3 
		WHERE id = $4`,
		verifiedBy.String(), now, now, id.String())
	return err
}

// ===================== USER LOOKUP =====================
func (r *AchievementRepository) GetStudentByUserID(userID uuid.UUID) (*model.Student, error) {
	query := `SELECT s.id, s.user_id, s.student_id, s.program_study, s.academic_year, s.created_at,
		             u.id, u.username, u.email, u.full_name, u.role_id, u.is_active, u.created_at, u.updated_at
		      FROM students s JOIN users u ON s.user_id = u.id WHERE s.user_id = $1`
	row := r.db.QueryRow(query, userID.String())

	var s model.Student
	var idStr, userIDStr string
	err := row.Scan(&idStr, &userIDStr, &s.StudentID, &s.ProgramStudy, &s.AcademicYear, &s.CreatedAt,
		&s.User.ID, &s.User.Username, &s.User.Email, &s.User.FullName, &s.User.RoleID,
		&s.User.IsActive, &s.User.CreatedAt, &s.User.UpdatedAt)
	if err != nil {
		return nil, err
	}
	s.ID = parseUUID(idStr)
	s.UserID = parseUUID(userIDStr)
	return &s, nil
}

func (r *AchievementRepository) GetLecturerByUserID(userID uuid.UUID) (*model.Lecturer, error) {
	query := `SELECT id, user_id, lecturer_id, department, created_at FROM lecturers WHERE user_id = $1`
	row := r.db.QueryRow(query, userID.String())

	var l model.Lecturer
	var idStr, userIDStr string
	err := row.Scan(&idStr, &userIDStr, &l.LecturerID, &l.Department, &l.CreatedAt)
	if err != nil {
		return nil, err
	}
	l.ID = parseUUID(idStr)
	l.UserID = parseUUID(userIDStr)
	return &l, nil
}

func (r *AchievementRepository) GetStudentIDsByAdvisor(advisorID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.db.Query(`SELECT id FROM students WHERE advisor_id = $1`, advisorID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var idStr string
		if err := rows.Scan(&idStr); err != nil {
			return nil, err
		}
		ids = append(ids, parseUUID(idStr))
	}
	return ids, nil
}