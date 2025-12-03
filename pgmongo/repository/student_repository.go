// File: BACKEND-UAS/pgmongo/repository/student_repository.go
package repository

import (
	"database/sql"

	"github.com/google/uuid"
	"BACKEND-UAS/pgmongo/model"
)

type StudentRepository struct {
	db *sql.DB
}

func NewStudentRepository(db *sql.DB) *StudentRepository {
	return &StudentRepository{db: db}
}

func uuidFromString(s string) uuid.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil
	}
	return id
}

// ===================== SCAN ROWS =====================
func (r *StudentRepository) scanStudentRows(rows *sql.Rows) ([]model.Student, error) {
	var students []model.Student
	for rows.Next() {
		var s model.Student
		var (
			sID         string
			sUserID     string
			advIDStr    sql.NullString
			aID         sql.NullString
			aUserID     sql.NullString
			aLecturerID sql.NullString
			aDepartment sql.NullString
			aCreatedAt  sql.NullTime
			auID        sql.NullString
			auUsername  sql.NullString
			auEmail     sql.NullString
			auFullName  sql.NullString
			auRoleID    sql.NullString
			auIsActive  sql.NullBool
			auCreatedAt sql.NullTime
			auUpdatedAt sql.NullTime
		)
		err := rows.Scan(
			&sID,
			&sUserID,
			&s.StudentID,
			&s.ProgramStudy,
			&s.AcademicYear,
			&advIDStr,
			&s.CreatedAt,
			&s.User.ID,
			&s.User.Username,
			&s.User.Email,
			&s.User.FullName,
			&s.User.RoleID,
			&s.User.IsActive,
			&s.User.CreatedAt,
			&s.User.UpdatedAt,
			&aID,
			&aUserID,
			&aLecturerID,
			&aDepartment,
			&aCreatedAt,
			&auID,
			&auUsername,
			&auEmail,
			&auFullName,
			&auRoleID,
			&auIsActive,
			&auCreatedAt,
			&auUpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		s.ID = uuidFromString(sID)
		s.UserID = uuidFromString(sUserID)
		if advIDStr.Valid {
			s.AdvisorID = uuidFromString(advIDStr.String)
		} else {
			s.AdvisorID = uuid.Nil
		}
		// Populate advisor only if AdvisorID is not Nil and fields are valid
		if s.AdvisorID != uuid.Nil && aID.Valid {
			s.Advisor.ID = uuidFromString(aID.String)
			s.Advisor.UserID = uuidFromString(aUserID.String)
			s.Advisor.LecturerID = getStringOrEmpty(aLecturerID)
			s.Advisor.Department = getStringOrEmpty(aDepartment)
			if aCreatedAt.Valid {
				s.Advisor.CreatedAt = aCreatedAt.Time
			}
			// Advisor User
			s.Advisor.User.ID = getStringOrEmpty(auID)
			s.Advisor.User.Username = getStringOrEmpty(auUsername)
			s.Advisor.User.Email = getStringOrEmpty(auEmail)
			s.Advisor.User.FullName = getStringOrEmpty(auFullName)
			s.Advisor.User.RoleID = getStringOrEmpty(auRoleID)
			if auIsActive.Valid {
				s.Advisor.User.IsActive = auIsActive.Bool
			}
			if auCreatedAt.Valid {
				s.Advisor.User.CreatedAt = auCreatedAt.Time
			}
			if auUpdatedAt.Valid {
				s.Advisor.User.UpdatedAt = auUpdatedAt.Time
			}
		}
		// Notifications default to empty slice
		s.Notifications = []model.Notification{}
		students = append(students, s)
	}
	return students, nil
}

func getStringOrEmpty(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

// ===================== LIST ALL =====================
func (r *StudentRepository) GetAllStudents(page, limit int) (*model.PaginatedResponse[model.Student], error) {
	var total int64
	countQuery := `SELECT COUNT(*) FROM students`
	err := r.db.QueryRow(countQuery).Scan(&total)
	if err != nil {
		return nil, err
	}

	offset := (page - 1) * limit
	query := `
SELECT 
    s.id, s.user_id, s.student_id, s.program_study, s.academic_year, s.advisor_id, s.created_at,
    u.id, u.username, u.email, u.full_name, u.role_id, u.is_active, u.created_at, u.updated_at,
    l.id, l.user_id, l.lecturer_id, l.department, l.created_at,
    lu.id, lu.username, lu.email, lu.full_name, lu.role_id, lu.is_active, lu.created_at, lu.updated_at
FROM students s
JOIN users u ON s.user_id = u.id
LEFT JOIN lecturers l ON s.advisor_id = l.id
LEFT JOIN users lu ON l.user_id = lu.id
ORDER BY s.created_at DESC
LIMIT $1 OFFSET $2
`
	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	data, err := r.scanStudentRows(rows)
	if err != nil {
		return nil, err
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

// ===================== DETAIL =====================
func (r *StudentRepository) GetStudentByID(id uuid.UUID) (*model.Student, error) {
	query := `
SELECT 
    s.id, s.user_id, s.student_id, s.program_study, s.academic_year, s.advisor_id, s.created_at,
    u.id, u.username, u.email, u.full_name, u.role_id, u.is_active, u.created_at, u.updated_at,
    l.id, l.user_id, l.lecturer_id, l.department, l.created_at,
    lu.id, lu.username, lu.email, lu.full_name, lu.role_id, lu.is_active, lu.created_at, lu.updated_at
FROM students s
JOIN users u ON s.user_id = u.id
LEFT JOIN lecturers l ON s.advisor_id = l.id
LEFT JOIN users lu ON l.user_id = lu.id
WHERE s.id = $1
`
	row := r.db.QueryRow(query, id.String())

	var s model.Student
	var (
		sID         string
		sUserID     string
		advIDStr    sql.NullString
		aID         sql.NullString
		aUserID     sql.NullString
		aLecturerID sql.NullString
		aDepartment sql.NullString
		aCreatedAt  sql.NullTime
		auID        sql.NullString
		auUsername  sql.NullString
		auEmail     sql.NullString
		auFullName  sql.NullString
		auRoleID    sql.NullString
		auIsActive  sql.NullBool
		auCreatedAt sql.NullTime
		auUpdatedAt sql.NullTime
	)
	err := row.Scan(
		&sID,
		&sUserID,
		&s.StudentID,
		&s.ProgramStudy,
		&s.AcademicYear,
		&advIDStr,
		&s.CreatedAt,
		&s.User.ID,
		&s.User.Username,
		&s.User.Email,
		&s.User.FullName,
		&s.User.RoleID,
		&s.User.IsActive,
		&s.User.CreatedAt,
		&s.User.UpdatedAt,
		&aID,
		&aUserID,
		&aLecturerID,
		&aDepartment,
		&aCreatedAt,
		&auID,
		&auUsername,
		&auEmail,
		&auFullName,
		&auRoleID,
		&auIsActive,
		&auCreatedAt,
		&auUpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	s.ID = uuidFromString(sID)
	s.UserID = uuidFromString(sUserID)
	if advIDStr.Valid {
		s.AdvisorID = uuidFromString(advIDStr.String)
	} else {
		s.AdvisorID = uuid.Nil
	}
	if s.AdvisorID != uuid.Nil && aID.Valid {
		s.Advisor.ID = uuidFromString(aID.String)
		s.Advisor.UserID = uuidFromString(aUserID.String)
		s.Advisor.LecturerID = getStringOrEmpty(aLecturerID)
		s.Advisor.Department = getStringOrEmpty(aDepartment)
		if aCreatedAt.Valid {
			s.Advisor.CreatedAt = aCreatedAt.Time
		}
		s.Advisor.User.ID = getStringOrEmpty(auID)
		s.Advisor.User.Username = getStringOrEmpty(auUsername)
		s.Advisor.User.Email = getStringOrEmpty(auEmail)
		s.Advisor.User.FullName = getStringOrEmpty(auFullName)
		s.Advisor.User.RoleID = getStringOrEmpty(auRoleID)
		if auIsActive.Valid {
			s.Advisor.User.IsActive = auIsActive.Bool
		}
		if auCreatedAt.Valid {
			s.Advisor.User.CreatedAt = auCreatedAt.Time
		}
		if auUpdatedAt.Valid {
			s.Advisor.User.UpdatedAt = auUpdatedAt.Time
		}
	}
	s.Notifications = []model.Notification{}
	return &s, nil
}

// GetStudentByUserID gets student by user ID (for own access)
func (r *StudentRepository) GetStudentByUserID(userID uuid.UUID) (*model.Student, error) {
	query := `
SELECT 
    s.id, s.user_id, s.student_id, s.program_study, s.academic_year, s.advisor_id, s.created_at,
    u.id, u.username, u.email, u.full_name, u.role_id, u.is_active, u.created_at, u.updated_at,
    l.id, l.user_id, l.lecturer_id, l.department, l.created_at,
    lu.id, lu.username, lu.email, lu.full_name, lu.role_id, lu.is_active, lu.created_at, lu.updated_at
FROM students s
JOIN users u ON s.user_id = u.id
LEFT JOIN lecturers l ON s.advisor_id = l.id
LEFT JOIN users lu ON l.user_id = lu.id
WHERE s.user_id = $1
`
	row := r.db.QueryRow(query, userID.String())

	var s model.Student
	var (
		sID         string
		sUserID     string
		advIDStr    sql.NullString
		aID         sql.NullString
		aUserID     sql.NullString
		aLecturerID sql.NullString
		aDepartment sql.NullString
		aCreatedAt  sql.NullTime
		auID        sql.NullString
		auUsername  sql.NullString
		auEmail     sql.NullString
		auFullName  sql.NullString
		auRoleID    sql.NullString
		auIsActive  sql.NullBool
		auCreatedAt sql.NullTime
		auUpdatedAt sql.NullTime
	)
	err := row.Scan(
		&sID,
		&sUserID,
		&s.StudentID,
		&s.ProgramStudy,
		&s.AcademicYear,
		&advIDStr,
		&s.CreatedAt,
		&s.User.ID,
		&s.User.Username,
		&s.User.Email,
		&s.User.FullName,
		&s.User.RoleID,
		&s.User.IsActive,
		&s.User.CreatedAt,
		&s.User.UpdatedAt,
		&aID,
		&aUserID,
		&aLecturerID,
		&aDepartment,
		&aCreatedAt,
		&auID,
		&auUsername,
		&auEmail,
		&auFullName,
		&auRoleID,
		&auIsActive,
		&auCreatedAt,
		&auUpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	s.ID = uuidFromString(sID)
	s.UserID = uuidFromString(sUserID)
	if advIDStr.Valid {
		s.AdvisorID = uuidFromString(advIDStr.String)
	} else {
		s.AdvisorID = uuid.Nil
	}
	if s.AdvisorID != uuid.Nil && aID.Valid {
		s.Advisor.ID = uuidFromString(aID.String)
		s.Advisor.UserID = uuidFromString(aUserID.String)
		s.Advisor.LecturerID = getStringOrEmpty(aLecturerID)
		s.Advisor.Department = getStringOrEmpty(aDepartment)
		if aCreatedAt.Valid {
			s.Advisor.CreatedAt = aCreatedAt.Time
		}
		s.Advisor.User.ID = getStringOrEmpty(auID)
		s.Advisor.User.Username = getStringOrEmpty(auUsername)
		s.Advisor.User.Email = getStringOrEmpty(auEmail)
		s.Advisor.User.FullName = getStringOrEmpty(auFullName)
		s.Advisor.User.RoleID = getStringOrEmpty(auRoleID)
		if auIsActive.Valid {
			s.Advisor.User.IsActive = auIsActive.Bool
		}
		if auCreatedAt.Valid {
			s.Advisor.User.CreatedAt = auCreatedAt.Time
		}
		if auUpdatedAt.Valid {
			s.Advisor.User.UpdatedAt = auUpdatedAt.Time
		}
	}
	s.Notifications = []model.Notification{}
	return &s, nil
}

// GetLecturerByUserID gets lecturer by user ID
func (r *StudentRepository) GetLecturerByUserID(userID uuid.UUID) (*model.Lecturer, error) {
	query := `SELECT id, user_id, lecturer_id, department, created_at FROM lecturers WHERE user_id = $1`
	row := r.db.QueryRow(query, userID.String())
	var l model.Lecturer
	var idStr, userIDStr string
	err := row.Scan(&idStr, &userIDStr, &l.LecturerID, &l.Department, &l.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	l.ID = uuidFromString(idStr)
	l.UserID = uuidFromString(userIDStr)
	return &l, nil
}

// GetAdviseesByLecturerID gets all students advised by a lecturer
func (r *StudentRepository) GetAdviseesByLecturerID(lecturerID uuid.UUID) ([]model.Student, error) {
	query := `
SELECT 
    s.id, s.user_id, s.student_id, s.program_study, s.academic_year, s.advisor_id, s.created_at,
    u.id, u.username, u.email, u.full_name, u.role_id, u.is_active, u.created_at, u.updated_at,
    l.id, l.user_id, l.lecturer_id, l.department, l.created_at,
    lu.id, lu.username, lu.email, lu.full_name, lu.role_id, lu.is_active, lu.created_at, lu.updated_at
FROM students s
JOIN users u ON s.user_id = u.id
LEFT JOIN lecturers l ON s.advisor_id = l.id
LEFT JOIN users lu ON l.user_id = lu.id
WHERE s.advisor_id = $1
ORDER BY s.created_at DESC
`
	rows, err := r.db.Query(query, lecturerID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanStudentRows(rows)
}

// ===================== UPDATE ADVISOR =====================
func (r *StudentRepository) UpdateStudentAdvisor(studentID, advisorID uuid.UUID) error {
	_, err := r.db.Exec(
		`UPDATE students SET advisor_id = $1 WHERE id = $2`,
		advisorID.String(),
		studentID.String(),
	)
	return err
}