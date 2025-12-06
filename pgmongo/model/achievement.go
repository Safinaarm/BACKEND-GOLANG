// File: BACKEND-UAS/pgmongo/model/achievement.go
package model

import (
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Achievement struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	StudentID      uuid.UUID          `bson:"studentId" json:"studentId"`
	AchievementType string            `bson:"achievementType" json:"achievementType"`
	Title          string             `bson:"title" json:"title"`
	Description    string             `bson:"description" json:"description"`
	Details        bson.M             `bson:"details" json:"details"`
	Attachments    []Attachment       `bson:"attachments" json:"attachments"`
	Tags           []string           `bson:"tags" json:"tags"`
	Points         int                `bson:"points" json:"points"`
	CreatedAt      time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt      time.Time          `bson:"updatedAt" json:"updatedAt"`
	DeletedAt      *time.Time         `bson:"deletedAt,omitempty" json:"deletedAt"`
	Level          string             `bson:"level,omitempty" json:"level"` // Added for competition level distribution

	StatusHistory []StatusHistory `bson:"statusHistory" json:"statusHistory"`
}

type Attachment struct {
	FileName   string    `bson:"fileName" json:"fileName"`
	FileURL    string    `bson:"fileUrl" json:"fileUrl"`
	FileType   string    `bson:"fileType" json:"fileType"`
	UploadedAt time.Time `bson:"uploadedAt" json:"uploadedAt"`
}

type StatusHistory struct {
	ID           uuid.UUID  `bson:"id" json:"id"`
	Status       string     `bson:"status" json:"status"`
	ChangedBy    *uuid.UUID `bson:"changedBy,omitempty" json:"changedBy"`
	ChangedAt    time.Time  `bson:"changedAt" json:"changedAt"`
	Note         string     `bson:"note" json:"note"`
}

type Notification struct {
	ID        uuid.UUID `bson:"id" json:"id"`
	Type      string    `bson:"type" json:"type"`
	Title     string    `bson:"title" json:"title"`
	Message   string    `bson:"message" json:"message"`
	Read      bool      `bson:"read" json:"read"`
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
}

// AchievementStatistics for global reports
type AchievementStatistics struct {
	TotalPerType   map[string]int64    `json:"total_per_type"`
	TotalPerPeriod map[string]int64    `json:"total_per_period"`
	TopStudents    []TopStudent        `json:"top_students"`
	Distribution   map[string]int64    `json:"distribution"`
}

type TopStudent struct {
	StudentID string `json:"student_id"`
	FullName  string `json:"full_name"`
	Points    int64  `json:"points"`
	Count     int64  `json:"count"`
}

// StudentAchievementStatistics for specific student reports
type StudentAchievementStatistics struct {
	TotalAchievements int64            `json:"total_achievements"`
	PerType           map[string]int64 `json:"per_type"`
	PerPeriod         map[string]int64 `json:"per_period"`
	Distribution      map[string]int64 `json:"distribution"`
}