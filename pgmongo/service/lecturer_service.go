package service

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"BACKEND-UAS/pgmongo/model"
	"BACKEND-UAS/pgmongo/repository"
)

type LecturerService interface {
	GetAllLecturers(ctx context.Context, userID uuid.UUID, page, limit int) (*model.PaginatedResponse[model.Lecturer], error)
	GetAdvisees(ctx context.Context, lecturerID, userID uuid.UUID) ([]model.Student, error)

	// Handler functions
	GetAllLecturersHandler(c *fiber.Ctx) error
	GetAdviseesHandler(c *fiber.Ctx) error
}

type lecturerService struct {
	lecturerRepo repository.LecturerRepository
	studentRepo  *repository.StudentRepository
}

func NewLecturerService(lecturerRepo repository.LecturerRepository, studentRepo *repository.StudentRepository) LecturerService {
	return &lecturerService{
		lecturerRepo: lecturerRepo,
		studentRepo:  studentRepo,
	}
}

// Core business logic
func (s *lecturerService) GetAllLecturers(ctx context.Context, userID uuid.UUID, page, limit int) (*model.PaginatedResponse[model.Lecturer], error) {
	// Lecturer: hanya data sendiri
	lecturer, err := s.lecturerRepo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}
	if lecturer != nil {
		data := []model.Lecturer{*lecturer}
		total := int64(1)
		totalPages := (total + int64(limit) - 1) / int64(limit)
		return &model.PaginatedResponse[model.Lecturer]{
			Data:       data,
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: int(totalPages),
		}, nil
	}

	// Student: hanya advisornya
	student, err := s.studentRepo.GetStudentByUserID(userID)
	if err != nil {
		return nil, err
	}
	if student != nil {
		if student.AdvisorID == uuid.Nil {
			return nil, fiber.NewError(http.StatusBadRequest, "no advisor assigned")
		}
		data := []model.Lecturer{student.Advisor}
		total := int64(1)
		totalPages := (total + int64(limit) - 1) / int64(limit)
		return &model.PaginatedResponse[model.Lecturer]{
			Data:       data,
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: int(totalPages),
		}, nil
	}

	// Admin: semua dengan pagination
	return s.lecturerRepo.GetAll(page, limit)
}

func (s *lecturerService) GetAdvisees(ctx context.Context, lecturerID, userID uuid.UUID) ([]model.Student, error) {
	lecturer, err := s.lecturerRepo.GetByID(lecturerID)
	if err != nil {
		return nil, err
	}
	if lecturer == nil {
		return nil, fiber.NewError(http.StatusNotFound, "lecturer not found")
	}

	// Access check untuk student
	student, err := s.studentRepo.GetStudentByUserID(userID)
	if err != nil {
		return nil, err
	}
	if student != nil && student.AdvisorID != lecturerID {
		return nil, fiber.NewError(http.StatusForbidden, "access denied")
	}

	// Access check untuk lecturer
	ownLecturer, err := s.lecturerRepo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}
	if ownLecturer != nil && ownLecturer.ID != lecturerID {
		return nil, fiber.NewError(http.StatusForbidden, "access denied")
	}

	return s.studentRepo.GetAdviseesByLecturerID(lecturerID)
}

// @Summary Get all lecturers
// @Description Mengambil daftar dosen berdasarkan role user (lecturer: own data, student: advisor data, admin: all with pagination)
// @Tags Lecturers
// @Accept json
// @Produce json
// @Param page query int false "Page number (default 1)"
// @Param limit query int false "Items per page (default 10)"
// @Success 200 {object} model.PaginatedResponse[model.Lecturer] "Paginated lecturers"
// @Failure 400 {object} model.ErrorResponse "No advisor assigned"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /lecturers [get]
func (s *lecturerService) GetAllLecturersHandler(c *fiber.Ctx) error {
	userIDStr, ok := c.Locals("user_id").(string)
	if !ok {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	pageStr := c.Query("page")
	limitStr := c.Query("limit")

	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(limitStr)
	if limit < 1 || limit > 100 {
		limit = 10
	}

	result, err := s.GetAllLecturers(c.Context(), userID, page, limit)
	if err != nil {
		if fiberErr, ok := err.(*fiber.Error); ok {
			return c.Status(fiberErr.Code).JSON(fiber.Map{"error": fiberErr.Message})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(result)
}

// @Summary Get lecturer's advisees
// @Description Mengambil daftar mahasiswa bimbingan dosen (dengan access check: own, advisor, or admin)
// @Tags Lecturers
// @Accept json
// @Produce json
// @Param id path string true "Lecturer ID (UUID)"
// @Success 200 {object} object "data array of model.Student"
// @Failure 400 {object} model.ErrorResponse "Access denied"
// @Failure 404 {object} model.ErrorResponse "Lecturer not found"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /lecturers/{id}/advisees [get]
func (s *lecturerService) GetAdviseesHandler(c *fiber.Ctx) error {
	idStr := c.Params("id")
	lecturerID, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid lecturer ID"})
	}

	userIDStr, ok := c.Locals("user_id").(string)
	if !ok {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	advisees, err := s.GetAdvisees(c.Context(), lecturerID, userID)
	if err != nil {
		if fiberErr, ok := err.(*fiber.Error); ok {
			return c.Status(fiberErr.Code).JSON(fiber.Map{"error": fiberErr.Message})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"data": advisees})
}