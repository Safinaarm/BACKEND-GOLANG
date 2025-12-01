// File: BACKEND-UAS/pgmongo/model/lecturer.go
package model

import (
	"time"

	"github.com/google/uuid"
)

type Lecturer struct {
	ID         uuid.UUID `json:"id" bson:"id,omitempty"`
	UserID     uuid.UUID `json:"user_id" bson:"user_id"`
	LecturerID string    `json:"lecturer_id" bson:"lecturer_id"`
	Department string    `json:"department" bson:"department"`
	CreatedAt  time.Time `json:"created_at" bson:"created_at"`

	User         User       `json:"user" bson:"user"`
	Notifications []Notification `json:"notifications" bson:"notifications"`
}