// File: BACKEND-UAS/pgmongo/service/lecturer_service.go
package service

import (
	"context"

	"github.com/google/uuid"

	"BACKEND-UAS/pgmongo/model"
	"BACKEND-UAS/pgmongo/repository"
)

type LecturerService interface {
	FindAll(ctx context.Context) ([]model.Lecturer, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.Lecturer, error)
	FindAdvisees(ctx context.Context, id uuid.UUID) ([]model.Student, error)
}

type lecturerService struct {
	repo repository.LecturerRepository
}

func NewLecturerService(repo repository.LecturerRepository) LecturerService {
	return &lecturerService{repo: repo}
}

func (s *lecturerService) FindAll(ctx context.Context) ([]model.Lecturer, error) {
	return s.repo.FindAll(ctx)
}

func (s *lecturerService) FindByID(ctx context.Context, id uuid.UUID) (*model.Lecturer, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *lecturerService) FindAdvisees(ctx context.Context, id uuid.UUID) ([]model.Student, error) {
	return s.repo.FindAdvisees(ctx, id)
}