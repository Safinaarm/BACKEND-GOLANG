// File: BACKEND-UAS/pgmongo/repository/lecturer_repository.go
package repository

import (
	"database/sql"

	"github.com/google/uuid"
	"BACKEND-UAS/pgmongo/model"
)

type LecturerRepository interface {
	GetAll(page, limit int) (*model.PaginatedResponse[model.Lecturer], error)
	GetByID(id uuid.UUID) (*model.Lecturer, error)
	GetByUserID(userID uuid.UUID) (*model.Lecturer, error)
}

type LecturerRepositoryImpl struct {
	db *sql.DB
}

func NewLecturerRepository(db *sql.DB) LecturerRepository {
	return &LecturerRepositoryImpl{db: db}
}

// ===================== SCAN LECTURER ROWS =====================
func (r *LecturerRepositoryImpl) scanLecturerRows(rows *sql.Rows) ([]model.Lecturer, error) {
	var lecturers []model.Lecturer
	for rows.Next() {
		var l model.Lecturer
		var (
			lID      string
			lUserID  string
			uID      sql.NullString
			uUsername sql.NullString
			uEmail   sql.NullString
			uFullName sql.NullString
			uRoleID  sql.NullString
			uIsActive sql.NullBool
			uCreatedAt sql.NullTime
			uUpdatedAt sql.NullTime
		)
		err := rows.Scan(
			&lID,
			&lUserID,
			&l.LecturerID,
			&l.Department,
			&l.CreatedAt,
			&uID,
			&uUsername,
			&uEmail,
			&uFullName,
			&uRoleID,
			&uIsActive,
			&uCreatedAt,
			&uUpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		l.ID = uuidFromString(lID)
		l.UserID = uuidFromString(lUserID)
		l.User.ID = getStringOrEmpty(uID)
		l.User.Username = getStringOrEmpty(uUsername)
		l.User.Email = getStringOrEmpty(uEmail)
		l.User.FullName = getStringOrEmpty(uFullName)
		l.User.RoleID = getStringOrEmpty(uRoleID)
		if uIsActive.Valid {
			l.User.IsActive = uIsActive.Bool
		}
		if uCreatedAt.Valid {
			l.User.CreatedAt = uCreatedAt.Time
		}
		if uUpdatedAt.Valid {
			l.User.UpdatedAt = uUpdatedAt.Time
		}
		l.Notifications = []model.Notification{}
		lecturers = append(lecturers, l)
	}
	return lecturers, nil
}

// ===================== LIST ALL =====================
func (r *LecturerRepositoryImpl) GetAll(page, limit int) (*model.PaginatedResponse[model.Lecturer], error) {
	var total int64
	countQuery := `SELECT COUNT(*) FROM lecturers`
	err := r.db.QueryRow(countQuery).Scan(&total)
	if err != nil {
		return nil, err
	}

	offset := (page - 1) * limit
	query := `
SELECT 
    l.id, l.user_id, l.lecturer_id, l.department, l.created_at,
    u.id, u.username, u.email, u.full_name, u.role_id, u.is_active, u.created_at, u.updated_at
FROM lecturers l
JOIN users u ON l.user_id = u.id
ORDER BY l.created_at DESC
LIMIT $1 OFFSET $2
`
	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	data, err := r.scanLecturerRows(rows)
	if err != nil {
		return nil, err
	}

	totalPages := (int(total) + limit - 1) / limit
	return &model.PaginatedResponse[model.Lecturer]{
		Data:       data,
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

// ===================== DETAIL BY ID =====================
func (r *LecturerRepositoryImpl) GetByID(id uuid.UUID) (*model.Lecturer, error) {
	query := `
SELECT 
    l.id, l.user_id, l.lecturer_id, l.department, l.created_at,
    u.id, u.username, u.email, u.full_name, u.role_id, u.is_active, u.created_at, u.updated_at
FROM lecturers l
JOIN users u ON l.user_id = u.id
WHERE l.id = $1
`
	row := r.db.QueryRow(query, id.String())

	var l model.Lecturer
	var (
		lID      string
		lUserID  string
		uID      sql.NullString
		uUsername sql.NullString
		uEmail   sql.NullString
		uFullName sql.NullString
		uRoleID  sql.NullString
		uIsActive sql.NullBool
		uCreatedAt sql.NullTime
		uUpdatedAt sql.NullTime
	)
	err := row.Scan(
		&lID,
		&lUserID,
		&l.LecturerID,
		&l.Department,
		&l.CreatedAt,
		&uID,
		&uUsername,
		&uEmail,
		&uFullName,
		&uRoleID,
		&uIsActive,
		&uCreatedAt,
		&uUpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	l.ID = uuidFromString(lID)
	l.UserID = uuidFromString(lUserID)
	l.User.ID = getStringOrEmpty(uID)
	l.User.Username = getStringOrEmpty(uUsername)
	l.User.Email = getStringOrEmpty(uEmail)
	l.User.FullName = getStringOrEmpty(uFullName)
	l.User.RoleID = getStringOrEmpty(uRoleID)
	if uIsActive.Valid {
		l.User.IsActive = uIsActive.Bool
	}
	if uCreatedAt.Valid {
		l.User.CreatedAt = uCreatedAt.Time
	}
	if uUpdatedAt.Valid {
		l.User.UpdatedAt = uUpdatedAt.Time
	}
	l.Notifications = []model.Notification{}
	return &l, nil
}

// ===================== DETAIL BY USER ID =====================
func (r *LecturerRepositoryImpl) GetByUserID(userID uuid.UUID) (*model.Lecturer, error) {
	query := `
SELECT 
    l.id, l.user_id, l.lecturer_id, l.department, l.created_at,
    u.id, u.username, u.email, u.full_name, u.role_id, u.is_active, u.created_at, u.updated_at
FROM lecturers l
JOIN users u ON l.user_id = u.id
WHERE l.user_id = $1
`
	row := r.db.QueryRow(query, userID.String())

	var l model.Lecturer
	var (
		lID      string
		lUserID  string
		uID      sql.NullString
		uUsername sql.NullString
		uEmail   sql.NullString
		uFullName sql.NullString
		uRoleID  sql.NullString
		uIsActive sql.NullBool
		uCreatedAt sql.NullTime
		uUpdatedAt sql.NullTime
	)
	err := row.Scan(
		&lID,
		&lUserID,
		&l.LecturerID,
		&l.Department,
		&l.CreatedAt,
		&uID,
		&uUsername,
		&uEmail,
		&uFullName,
		&uRoleID,
		&uIsActive,
		&uCreatedAt,
		&uUpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	l.ID = uuidFromString(lID)
	l.UserID = uuidFromString(lUserID)
	l.User.ID = getStringOrEmpty(uID)
	l.User.Username = getStringOrEmpty(uUsername)
	l.User.Email = getStringOrEmpty(uEmail)
	l.User.FullName = getStringOrEmpty(uFullName)
	l.User.RoleID = getStringOrEmpty(uRoleID)
	if uIsActive.Valid {
		l.User.IsActive = uIsActive.Bool
	}
	if uCreatedAt.Valid {
		l.User.CreatedAt = uCreatedAt.Time
	}
	if uUpdatedAt.Valid {
		l.User.UpdatedAt = uUpdatedAt.Time
	}
	l.Notifications = []model.Notification{}
	return &l, nil
}