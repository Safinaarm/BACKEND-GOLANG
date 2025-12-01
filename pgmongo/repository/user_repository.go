// postgre/repository/user_repository.go
package repository

import (
	"context"
	"database/sql"
	"time"

	"BACKEND-UAS/pgmongo/model"
)

type UserRepository interface {
	FindByUsernameOrEmail(ctx context.Context, identifier string) (*model.User, error)
	FindByID(ctx context.Context, id string) (*model.User, error)
	GetAll(ctx context.Context) ([]*model.User, error)
	Create(ctx context.Context, user *model.User) error
	Update(ctx context.Context, id string, user *model.User) error
	Delete(ctx context.Context, id string) error
	UpdateRole(ctx context.Context, id, roleID string) error
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

func (r *userRepository) GetAll(ctx context.Context) ([]*model.User, error) {
	q := `SELECT id, username, email, full_name, role_id, is_active, created_at, updated_at
	      FROM users
	      ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		u := &model.User{}
		var createdAt, updatedAt sql.NullTime
		err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.FullName, &u.RoleID, &u.IsActive, &createdAt, &updatedAt)
		if err != nil {
			return nil, err
		}
		if createdAt.Valid {
			u.CreatedAt = createdAt.Time
		}
		if updatedAt.Valid {
			u.UpdatedAt = updatedAt.Time
		}
		users = append(users, u)
	}
	return users, nil
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	q := `INSERT INTO users (id, username, email, password_hash, full_name, role_id, is_active, created_at, updated_at)
	      VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := r.db.ExecContext(ctx, q, user.ID, user.Username, user.Email, user.PasswordHash, user.FullName, user.RoleID, user.IsActive, user.CreatedAt, user.UpdatedAt)
	return err
}

func (r *userRepository) Update(ctx context.Context, id string, user *model.User) error {
	q := `UPDATE users SET username=$1, email=$2, password_hash=$3, full_name=$4, role_id=$5, is_active=$6, updated_at=$7
	      WHERE id=$8`
	_, err := r.db.ExecContext(ctx, q, user.Username, user.Email, user.PasswordHash, user.FullName, user.RoleID, user.IsActive, user.UpdatedAt, id)
	return err
}

func (r *userRepository) Delete(ctx context.Context, id string) error {
	q := `DELETE FROM users WHERE id=$1`
	_, err := r.db.ExecContext(ctx, q, id)
	return err
}

func (r *userRepository) UpdateRole(ctx context.Context, id, roleID string) error {
	q := `UPDATE users SET role_id=$1, updated_at=$2 WHERE id=$3`
	_, err := r.db.ExecContext(ctx, q, roleID, time.Now(), id)
	return err
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
