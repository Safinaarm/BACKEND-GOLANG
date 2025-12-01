// File: BACKEND-UAS/pgmongo/model/achievement_detail_response.go
package model
import (
	"time"

	"github.com/google/uuid"
)

type AchievementDetailResponse struct {
    ID            string          `json:"id"`
    Student       Student         `json:"student"`
    Status        string          `json:"status"`
    SubmittedAt   *time.Time      `json:"submitted_at,omitempty"`
    VerifiedAt    *time.Time      `json:"verified_at,omitempty"`
    VerifiedBy    *uuid.UUID      `json:"verified_by,omitempty"`
    RejectionNote string          `json:"rejection_note,omitempty"`
    Achievement   Achievement      `json:"achievement"`
    StatusHistory []StatusHistory `json:"statusHistory"`
}