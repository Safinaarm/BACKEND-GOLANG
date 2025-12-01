// File: BACKEND-UAS/pgmongo/model/permission.go
package model

import (
	"time"

	"github.com/google/uuid"
)

type Permission struct {
	ID          uuid.UUID `json:"id" bson:"id,omitempty"`
	Name        string    `json:"name" bson:"name"`
	Resource    string    `json:"resource" bson:"resource"`
	Action      string    `json:"action" bson:"action"`
	Description string    `json:"description" bson:"description"`
	CreatedAt   time.Time `json:"created_at" bson:"created_at"`
}