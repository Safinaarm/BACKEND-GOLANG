// File: BACKEND-UAS/pgmongo/model/achievement_reference.go
package model

import (
	"time"

	"github.com/google/uuid"
)

type AchievementReference struct {
	ID               uuid.UUID      `json:"id" bson:"id,omitempty"`
	StudentID        uuid.UUID      `json:"student_id" bson:"student_id"`
	MongoAchievementID string       `json:"mongo_achievement_id" bson:"mongo_achievement_id"`
	Status           string         `json:"status" bson:"status"`
	SubmittedAt      *time.Time     `json:"submitted_at" bson:"submitted_at"`
	VerifiedAt       *time.Time     `json:"verified_at" bson:"verified_at"`
	VerifiedBy       *uuid.UUID     `json:"verified_by" bson:"verified_by"`
	RejectionNote    string         `json:"rejection_note" bson:"rejection_note"`
	CreatedAt        time.Time      `json:"created_at" bson:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at" bson:"updated_at"`

	Student    Student      `json:"student" bson:"student"`
	Achievement *Achievement `json:"achievement" bson:"-"`
}