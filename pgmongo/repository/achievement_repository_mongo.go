// File: BACKEND-UAS/pgmongo/repository/achievement_repository_mongo.go
package repository

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"BACKEND-UAS/pgmongo/model"
)

type AchievementRepositoryMongo struct {
	coll *mongo.Collection
}

func NewAchievementRepositoryMongo(client *mongo.Client) *AchievementRepositoryMongo {
	coll := client.Database("your_db").Collection("achievements")
	return &AchievementRepositoryMongo{coll: coll}
}

// GetAchievementByID gets a single achievement from Mongo

// File: BACKEND-UAS/pgmongo/repository/achievement_repository_mongo.go
func (r *AchievementRepositoryMongo) GetAchievementByID(mongoID string) (*model.Achievement, error) {
	objID, err := primitive.ObjectIDFromHex(mongoID)
	if err != nil {
		return nil, fmt.Errorf("invalid mongoID: %v", err)
	}
	var ach model.Achievement
	err = r.coll.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&ach)
	if err == mongo.ErrNoDocuments {
		return nil, nil // Tidak ditemukan, bukan error
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch achievement: %v", err)
	}
	if ach.DeletedAt != nil {
		return nil, nil // Soft-deleted, return nil
	}
	return &ach, nil
}

// CreateAchievement creates a new achievement in Mongo with initial history
func (r *AchievementRepositoryMongo) CreateAchievement(ach *model.Achievement) error {
	ach.CreatedAt = time.Now()
	ach.UpdatedAt = time.Now()
	ach.StatusHistory = []model.StatusHistory{
		{
			ID:        uuid.New(),
			Status:    "draft",
			ChangedAt: time.Now(),
			Note:      "Prestasi dibuat",
		},
	}
	res, err := r.coll.InsertOne(context.Background(), ach)
	if err != nil {
		return err
	}
	ach.ID = res.InsertedID.(primitive.ObjectID)
	return nil
}

// UpdateAchievement updates an achievement in Mongo
func (r *AchievementRepositoryMongo) UpdateAchievement(mongoID string, ach *model.Achievement) error {
	objID, err := primitive.ObjectIDFromHex(mongoID)
	if err != nil {
		return err
	}
	ach.UpdatedAt = time.Now()
	_, err = r.coll.UpdateOne(context.Background(), bson.M{"_id": objID}, bson.M{"$set": ach})
	return err
}

// SoftDeleteAchievement soft deletes an achievement in Mongo
func (r *AchievementRepositoryMongo) SoftDeleteAchievement(mongoID string) error {
	objID, err := primitive.ObjectIDFromHex(mongoID)
	if err != nil {
		return err
	}
	now := time.Now()
	_, err = r.coll.UpdateOne(context.Background(), bson.M{"_id": objID}, bson.M{"$set": bson.M{"deletedAt": now, "updatedAt": now}})
	return err
}

// GetAchievementsByStudentIDs gets achievements by student IDs from Mongo
func (r *AchievementRepositoryMongo) GetAchievementsByStudentIDs(studentIDs []uuid.UUID) ([]model.Achievement, error) {
	filter := bson.M{"studentId": bson.M{"$in": studentIDs}, "deletedAt": bson.M{"$exists": false}}
	var achievements []model.Achievement
	cursor, err := r.coll.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())
	err = cursor.All(context.Background(), &achievements)
	return achievements, err
}

// AddStatusHistory adds a status history entry to an achievement in Mongo
func (r *AchievementRepositoryMongo) AddStatusHistory(mongoID string, history model.StatusHistory) error {
	objID, err := primitive.ObjectIDFromHex(mongoID)
	if err != nil {
		return err
	}
	history.ID = uuid.New()
	history.ChangedAt = time.Now()
	_, err = r.coll.UpdateOne(context.Background(), bson.M{"_id": objID}, bson.M{"$push": bson.M{"statusHistory": history}})
	return err
}

// AddNotification adds a notification to an achievement in Mongo
func (r *AchievementRepositoryMongo) AddNotification(mongoID string, notif model.Notification) error {
	objID, err := primitive.ObjectIDFromHex(mongoID)
	if err != nil {
		return err
	}
	notif.ID = uuid.New()
	notif.CreatedAt = time.Now()
	_, err = r.coll.UpdateOne(context.Background(), bson.M{"_id": objID}, bson.M{"$push": bson.M{"notifications": notif}})
	return err
}

// UploadAttachment adds an attachment to an achievement in Mongo
func (r *AchievementRepositoryMongo) UploadAttachment(mongoID string, file io.Reader, fileName, fileType string) (*model.Attachment, error) {
	objID, err := primitive.ObjectIDFromHex(mongoID)
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll("./uploads", 0755); err != nil {
		return nil, err
	}

	filePath := filepath.Join("./uploads", fileName)
	out, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}
	defer out.Close()
	if _, err := io.Copy(out, file); err != nil {
		return nil, err
	}

	attachment := &model.Attachment{
		FileName:   fileName,
		FileURL:    "/uploads/" + fileName,
		FileType:   fileType,
		UploadedAt: time.Now(),
	}

	update := bson.M{"$push": bson.M{"attachments": attachment}}
	_, err = r.coll.UpdateOne(context.Background(), bson.M{"_id": objID}, update)
	if err != nil {
		os.Remove(filePath)
		return nil, err
	}

	return attachment, nil
}
