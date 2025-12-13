// model/user.go
package model

import "time"

// User represents a user in the system
// @Schema description: User model for CRUD and auth operations
type User struct {
	ID           string    `db:"id" json:"id" schema:"id,required"`
	Username     string    `db:"username" json:"username" schema:"username,required,minLength=3,maxLength=50"`
	Email        string    `db:"email" json:"email" schema:"email,required,email"`
	PasswordHash string    `db:"password_hash" json:"-" schema:"-"` // Hashed password (never exposed)
	FullName     string    `db:"full_name" json:"full_name" schema:"full_name,required,minLength=2"`
	RoleID       string    `db:"role_id" json:"role_id" schema:"role_id,required"`
	IsActive     bool      `db:"is_active" json:"is_active" schema:"is_active,default=true"`
	CreatedAt    time.Time `db:"created_at" json:"created_at" schema:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at" schema:"updated_at"`
}

// LoginRequest uses User fields for simplicity (only username/email + password)
// @Schema description: Login credentials using User model subset
type LoginRequest struct {
	Username string `json:"username" schema:"username,required"` // or email
	Password string `json:"password" schema:"password,required,minLength=6"`
}

// AuthResponse wraps login response using User model
// @Schema description: Auth success response with User
type AuthResponse struct {
	Status string `json:"status" schema:"status"`
	Data   struct {
		Token       string `json:"token" schema:"token"`
		RefreshToken string `json:"refreshToken" schema:"refreshToken"`
		User         User   `json:"user" schema:"user"`
		Role         string `json:"role" schema:"role"`
		Permissions  []string `json:"permissions" schema:"permissions"`
	} `json:"data" schema:"data"`
}
// ErrorResponse adalah format standar response saat terjadi error
type ErrorResponse struct {
	Message string `json:"message"`
}

// UserResponse adalah wrapper untuk response list user (untuk pagination nanti)
type UserResponse struct {
	Data       []*User `json:"data"`
	Total      int64   `json:"total,omitempty"`
	Page       int     `json:"page,omitempty"`
	Limit      int     `json:"limit,omitempty"`
	TotalPages int     `json:"total_pages,omitempty"`
}