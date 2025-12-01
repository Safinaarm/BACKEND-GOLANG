// File: BACKEND-UAS/pgmongo/model/student.go
package model

import (
	"time"

	"github.com/google/uuid"
)

type Student struct {
	ID           uuid.UUID `json:"id" bson:"id,omitempty"`
	UserID       uuid.UUID `json:"user_id" bson:"user_id"`
	StudentID    string    `json:"student_id" bson:"student_id"`
	ProgramStudy string    `json:"program_study" bson:"program_study"`
	AcademicYear string    `json:"academic_year" bson:"academic_year"`
	AdvisorID    uuid.UUID `json:"advisor_id" bson:"advisor_id"`
	CreatedAt    time.Time `json:"created_at" bson:"created_at"`

	User         User       `json:"user" bson:"user"`
	Advisor      Lecturer   `json:"advisor" bson:"advisor"`
	Notifications []Notification `json:"notifications" bson:"notifications"`
}