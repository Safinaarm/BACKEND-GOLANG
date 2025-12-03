// File: BACKEND-UAS/pgmongo/repository/lecturer_repository.go
package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"BACKEND-UAS/pgmongo/model"
)

type LecturerRepository interface {
	FindAll(ctx context.Context) ([]model.Lecturer, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.Lecturer, error)
	FindAdvisees(ctx context.Context, id uuid.UUID) ([]model.Student, error)
}

type lecturerRepository struct {
	coll *mongo.Collection
	studentColl *mongo.Collection
}

func NewLecturerRepository(db *mongo.Database) LecturerRepository {
	return &lecturerRepository{
		coll: db.Collection("lecturers"),
		studentColl: db.Collection("students"),
	}
}

func (r *lecturerRepository) FindAll(ctx context.Context) ([]model.Lecturer, error) {
	filter := bson.M{}
	opts := options.Find()
	cur, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var lecturers []model.Lecturer
	if err = cur.All(ctx, &lecturers); err != nil {
		return nil, err
	}
	return lecturers, nil
}

func (r *lecturerRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Lecturer, error) {
	var lecturer model.Lecturer
	filter := bson.M{"id": id}
	err := r.coll.FindOne(ctx, filter).Decode(&lecturer)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("lecturer not found")
		}
		return nil, err
	}
	return &lecturer, nil
}

func (r *lecturerRepository) FindAdvisees(ctx context.Context, id uuid.UUID) ([]model.Student, error) {
	filter := bson.M{"advisor_id": id}
	opts := options.Find()
	cur, err := r.studentColl.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var students []model.Student
	if err = cur.All(ctx, &students); err != nil {
		return nil, err
	}
	return students, nil
}