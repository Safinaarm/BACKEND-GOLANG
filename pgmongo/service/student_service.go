package service

import (
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"BACKEND-UAS/pgmongo/model"
	"BACKEND-UAS/pgmongo/repository"
)

type StudentService struct {
	studentRepo *repository.StudentRepository
	achRepo     *repository.AchievementRepository
}

func NewStudentService(studentRepo *repository.StudentRepository, achRepo *repository.AchievementRepository) *StudentService {
	return &StudentService{
		studentRepo: studentRepo,
		achRepo:     achRepo,
	}
}

// Core business logic (dipertahankan & diperbaiki sedikit)
func (s *StudentService) GetAllStudents(userID uuid.UUID, page, limit int) (*model.PaginatedResponse[model.Student], error) {
	var data []model.Student
	var total int64

	student, err := s.studentRepo.GetStudentByUserID(userID)
	if err != nil {
		return nil, err
	}
	if student != nil {
		// Student: hanya data sendiri
		data = []model.Student{*student}
		total = 1
	} else {
		lect, err := s.studentRepo.GetLecturerByUserID(userID)
		if err != nil {
			return nil, err
		}
		if lect != nil {
			// Lecturer: hanya advisees
			advisees, err := s.studentRepo.GetAdviseesByLecturerID(lect.ID)
			if err != nil {
				return nil, err
			}
			data = advisees
			total = int64(len(data))
		} else {
			// Admin: semua dengan pagination
			return s.studentRepo.GetAllStudents(page, limit)
		}
	}

	totalPages := (int(total) + limit - 1) / limit
	return &model.PaginatedResponse[model.Student]{
		Data:       data,
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

func (s *StudentService) GetOwnStudentProfile(userID uuid.UUID) (*model.Student, error) {
	student, err := s.studentRepo.GetStudentByUserID(userID)
	if err != nil {
		return nil, err
	}
	if student == nil {
		return nil, fiber.NewError(http.StatusNotFound, "no student profile found")
	}
	return student, nil
}

func (s *StudentService) GetStudentByID(id, userID uuid.UUID) (*model.Student, error) {
	student, err := s.studentRepo.GetStudentByID(id)
	if err != nil {
		return nil, err
	}
	if student == nil {
		return nil, fiber.NewError(http.StatusNotFound, "student not found")
	}

	// Access check
	isStudent, err := s.isStudent(userID)
	if err != nil {
		return nil, err
	}
	if isStudent && student.UserID != userID {
		return nil, fiber.NewError(http.StatusForbidden, "access denied")
	}

	isLecturer, err := s.isLecturer(userID)
	if err != nil {
		return nil, err
	}
	if isLecturer {
		lectID, err := s.getLecturerID(userID)
		if err != nil {
			return nil, err
		}
		if student.AdvisorID != lectID {
			return nil, fiber.NewError(http.StatusForbidden, "access denied")
		}
	}

	// Admin atau yang berhak
	return student, nil
}

func (s *StudentService) GetStudentAchievements(studentID, userID uuid.UUID, status *string, page, limit int) (*model.PaginatedResponse[model.AchievementReference], error) {
	// Access check via GetStudentByID
	_, err := s.GetStudentByID(studentID, userID)
	if err != nil {
		return nil, err
	}

	return s.achRepo.GetAchievementReferencesByStudentIDs([]uuid.UUID{studentID}, status, page, limit)
}

func (s *StudentService) UpdateStudentAdvisor(studentID, advisorID, userID uuid.UUID) error {
	// Hanya admin yang boleh
	isStudent, _ := s.isStudent(userID)
	isLecturer, _ := s.isLecturer(userID)
	if isStudent || isLecturer {
		return fiber.NewError(http.StatusForbidden, "unauthorized (only admin)")
	}

	return s.studentRepo.UpdateStudentAdvisor(studentID, advisorID)
}

// Helper methods (tetap)
func (s *StudentService) isStudent(userID uuid.UUID) (bool, error) {
	student, err := s.studentRepo.GetStudentByUserID(userID)
	if err != nil {
		return false, err
	}
	return student != nil, nil
}

func (s *StudentService) isLecturer(userID uuid.UUID) (bool, error) {
	lect, err := s.studentRepo.GetLecturerByUserID(userID)
	if err != nil {
		return false, err
	}
	return lect != nil, nil
}

func (s *StudentService) getLecturerID(userID uuid.UUID) (uuid.UUID, error) {
	lect, err := s.studentRepo.GetLecturerByUserID(userID)
	if err != nil {
		return uuid.Nil, err
	}
	if lect == nil {
		return uuid.Nil, fiber.NewError(http.StatusForbidden, "not a lecturer")
	}
	return lect.ID, nil
}

// ==================== HANDLERS WITH SWAGGER ====================

// @Summary Get all students
// @Description Mengambil daftar mahasiswa berdasarkan role user (student: own, lecturer: advisees, admin: all with pagination)
// @Tags Students
// @Accept json
// @Produce json
// @Param page query int false "Page number (default 1)"
// @Param limit query int false "Items per page (default 10)"
// @Success 200 {object} model.PaginatedResponse[model.Student]
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /students [get]
func (s *StudentService) GetAllStudentsHandler(c *fiber.Ctx) error {
	userIDStr, _ := c.Locals("user_id").(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	page, _ := strconv.Atoi(c.Query("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(c.Query("limit"))
	if limit < 1 || limit > 100 {
		limit = 10
	}

	result, err := s.GetAllStudents(userID, page, limit)
	if err != nil {
		if fe, ok := err.(*fiber.Error); ok {
			return c.Status(fe.Code).JSON(fiber.Map{"error": fe.Message})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(result)
}

// @Summary Get own student profile
// @Description Mengambil profil mahasiswa untuk user yang login (jika mahasiswa)
// @Tags Students
// @Accept json
// @Produce json
// @Success 200 {object} model.Student
// @Failure 404 {object} model.ErrorResponse "No student profile found"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /students/me [get]
func (s *StudentService) GetOwnProfileHandler(c *fiber.Ctx) error {
	userIDStr, _ := c.Locals("user_id").(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	student, err := s.GetOwnStudentProfile(userID)
	if err != nil {
		if fe, ok := err.(*fiber.Error); ok {
			return c.Status(fe.Code).JSON(fiber.Map{"error": fe.Message})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(student)
}

// @Summary Get student by ID
// @Description Mengambil detail mahasiswa berdasarkan ID, dengan access check (own, advisor, admin)
// @Tags Students
// @Accept json
// @Produce json
// @Param id path string true "Student ID (UUID)"
// @Success 200 {object} model.Student
// @Failure 400 {object} model.ErrorResponse "Access denied"
// @Failure 404 {object} model.ErrorResponse "Student not found"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /students/{id} [get]
func (s *StudentService) GetStudentByIDHandler(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid student ID"})
	}

	userIDStr, _ := c.Locals("user_id").(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	student, err := s.GetStudentByID(id, userID)
	if err != nil {
		if fe, ok := err.(*fiber.Error); ok {
			return c.Status(fe.Code).JSON(fiber.Map{"error": fe.Message})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(student)
}

// @Summary Get student achievements
// @Description Mengambil prestasi mahasiswa, dengan access check dan filter status/pagination
// @Tags Students
// @Accept json
// @Produce json
// @Param id path string true "Student ID (UUID)"
// @Param status query string false "Filter status (draft, submitted, verified, rejected, deleted)"
// @Param page query int false "Page number (default 1)"
// @Param limit query int false "Items per page (default 10)"
// @Success 200 {object} model.PaginatedResponse[model.AchievementReference]
// @Failure 400 {object} model.ErrorResponse "Access denied"
// @Failure 404 {object} model.ErrorResponse "Student not found"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /students/{id}/achievements [get]
func (s *StudentService) GetAchievementsHandler(c *fiber.Ctx) error {
	idStr := c.Params("id")
	studentID, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid student ID"})
	}

	userIDStr, _ := c.Locals("user_id").(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	status := c.Query("status")
	var statusPtr *string
	if status != "" {
		statusPtr = &status
	}

	page, _ := strconv.Atoi(c.Query("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(c.Query("limit"))
	if limit < 1 || limit > 100 {
		limit = 10
	}

	result, err := s.GetStudentAchievements(studentID, userID, statusPtr, page, limit)
	if err != nil {
		if fe, ok := err.(*fiber.Error); ok {
			return c.Status(fe.Code).JSON(fiber.Map{"error": fe.Message})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(result)
}

// @Summary Update student advisor
// @Description Memperbarui dosen wali mahasiswa (hanya untuk admin)
// @Tags Students
// @Accept json
// @Produce json
// @Param id path string true "Student ID (UUID)"
// @Param body body object true "Request body" schema={ "type": "object", "properties": { "advisor_id": { "type": "string" } }, "required": ["advisor_id"] }
// @Success 200 {object} map[string]string "message: Advisor updated successfully"
// @Failure 400 {object} model.ErrorResponse "Invalid request / Unauthorized"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /students/{id}/advisor [put]
func (s *StudentService) UpdateAdvisorHandler(c *fiber.Ctx) error {
	idStr := c.Params("id")
	studentID, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid student ID"})
	}

	userIDStr, _ := c.Locals("user_id").(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	type Req struct {
		AdvisorID string `json:"advisor_id" validate:"required,uuid"`
	}
	var req Req
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	advisorID, err := uuid.Parse(req.AdvisorID)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid advisor ID"})
	}

	err = s.UpdateStudentAdvisor(studentID, advisorID, userID)
	if err != nil {
		if fe, ok := err.(*fiber.Error); ok {
			return c.Status(fe.Code).JSON(fiber.Map{"error": fe.Message})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Advisor updated successfully"})
}