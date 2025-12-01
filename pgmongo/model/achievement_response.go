// File: BACKEND-UAS/pgmongo/model/achievement_response.go
package model

type AchievementResponse struct {
	Reference   AchievementReference `json:"reference"`
	Achievement Achievement          `json:"achievement"`
}