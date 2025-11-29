// postgre/repository/user_repository.go
package repository

import (
	"context"
	"database/sql"

	"BACKEND-UAS/postgre/model"
)

type UserRepository interface {
	FindByUsernameOrEmail(ctx context.Context, identifier string) (*model.User, error)
	FindByID(ctx context.Context, id string) (*model.User, error)
	GetRoleNameByID(ctx context.Context, roleID string) (string, error)
	GetPermissionsByRoleID(ctx context.Context, roleID string) ([]string, error)
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindByUsernameOrEmail(ctx context.Context, identifier string) (*model.User, error) {
	q := `SELECT id, username, email, password_hash, full_name, role_id, is_active, created_at, updated_at
	      FROM users
	      WHERE username=$1 OR email=$1 LIMIT 1`
	u := &model.User{}
	row := r.db.QueryRowContext(ctx, q, identifier)
	var createdAt, updatedAt sql.NullTime
	err := row.Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.FullName, &u.RoleID, &u.IsActive, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}
	if createdAt.Valid {
		u.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		u.UpdatedAt = updatedAt.Time
	}
	return u, nil
}

func (r *userRepository) FindByID(ctx context.Context, id string) (*model.User, error) {
	q := `SELECT id, username, email, password_hash, full_name, role_id, is_active, created_at, updated_at
	      FROM users
	      WHERE id=$1 LIMIT 1`
	u := &model.User{}
	row := r.db.QueryRowContext(ctx, q, id)
	var createdAt, updatedAt sql.NullTime
	err := row.Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.FullName, &u.RoleID, &u.IsActive, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}
	if createdAt.Valid {
		u.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		u.UpdatedAt = updatedAt.Time
	}
	return u, nil
}

func (r *userRepository) GetRoleNameByID(ctx context.Context, roleID string) (string, error) {
	var name string
	err := r.db.QueryRowContext(ctx, `SELECT name FROM roles WHERE id=$1`, roleID).Scan(&name)
	if err != nil {
		return "", err
	}
	return name, nil
}

func (r *userRepository) GetPermissionsByRoleID(ctx context.Context, roleID string) ([]string, error) {
	q := `
	  SELECT p.name
	  FROM role_permissions rp
	  JOIN permissions p ON rp.permission_id = p.id
	  WHERE rp.role_id = $1
	`
	rows, err := r.db.QueryContext(ctx, q, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var perms []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, err
		}
		perms = append(perms, p)
	}
	return perms, nil
}