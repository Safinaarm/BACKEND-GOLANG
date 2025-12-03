// File: BACKEND-UAS/route/student_routes.go
package route

import (
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"BACKEND-UAS/middleware"
	"BACKEND-UAS/pgmongo/service"
)

// SetupStudentRoutes sets up the student routes with auth middleware
func SetupStudentRoutes(app *fiber.App, studentSvc *service.StudentService, authM *middleware.AuthMiddlewareConfig) {
	v1 := app.Group("/api/v1")
	{
		students := v1.Group("/students")
		students.Use(authM.AuthRequired()) // Apply auth middleware
		{
			// GET /api/v1/students - Filtered by user type (from DB)
			students.Get("/", func(c *fiber.Ctx) error {
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

				paginated, err := studentSvc.GetAllStudents(userID, page, limit)
				if err != nil {
					return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
				}

				return c.JSON(paginated)
			})

			// GET /api/v1/students/me - Get own profile (convenience for student)
			students.Get("/me", func(c *fiber.Ctx) error {
				userIDStr, ok := c.Locals("user_id").(string)
				if !ok {
					return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
				}
				userID, err := uuid.Parse(userIDStr)
				if err != nil {
					return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
				}

				student, err := studentSvc.GetOwnStudentProfile(userID)
				if err != nil {
					if err.Error() == "no student profile found" {
						return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "No student profile found"})
					}
					return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
				}

				return c.JSON(student)
			})

			// GET /api/v1/students/:id - With access check (specific ID)
			students.Get("/:id", func(c *fiber.Ctx) error {
				idStr := c.Params("id")
				id, err := uuid.Parse(idStr)
				if err != nil {
					return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid student ID"})
				}

				userIDStr, ok := c.Locals("user_id").(string)
				if !ok {
					return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
				}
				userID, err := uuid.Parse(userIDStr)
				if err != nil {
					return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
				}

				student, err := studentSvc.GetStudentByID(id, userID)
				if err != nil {
					if err.Error() == "access denied" {
						return c.Status(http.StatusForbidden).JSON(fiber.Map{"error": "Access denied"})
					}
					return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
				}
				if student == nil {
					return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Student not found"})
				}

				return c.JSON(student)
			})

			// GET /api/v1/students/:id/achievements - With access check
			students.Get("/:id/achievements", func(c *fiber.Ctx) error {
				idStr := c.Params("id")
				id, err := uuid.Parse(idStr)
				if err != nil {
					return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid student ID"})
				}

				userIDStr, ok := c.Locals("user_id").(string)
				if !ok {
					return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
				}
				userID, err := uuid.Parse(userIDStr)
				if err != nil {
					return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
				}

				status := c.Query("status")
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

				var statusPtr *string
				if status != "" {
					statusPtr = &status
				}

				paginated, err := studentSvc.GetStudentAchievements(id, userID, statusPtr, page, limit)
				if err != nil {
					if err.Error() == "access denied" {
						return c.Status(http.StatusForbidden).JSON(fiber.Map{"error": "Access denied"})
					}
					return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
				}

				return c.JSON(paginated)
			})

			// PUT /api/v1/students/:id/advisor - Admin only
			students.Put("/:id/advisor", func(c *fiber.Ctx) error {
				idStr := c.Params("id")
				id, err := uuid.Parse(idStr)
				if err != nil {
					return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid student ID"})
				}

				userIDStr, ok := c.Locals("user_id").(string)
				if !ok {
					return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
				}
				userID, err := uuid.Parse(userIDStr)
				if err != nil {
					return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
				}

				type UpdateAdvisorRequest struct {
					AdvisorID string `json:"advisor_id"`
				}
				var req UpdateAdvisorRequest
				if err := c.BodyParser(&req); err != nil {
					return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
				}
				if req.AdvisorID == "" {
					return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Advisor ID is required"})
				}

				advisorID, err := uuid.Parse(req.AdvisorID)
				if err != nil {
					return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid advisor ID"})
				}

				err = studentSvc.UpdateStudentAdvisor(id, advisorID, userID)
				if err != nil {
					if err.Error() == "unauthorized" {
						return c.Status(http.StatusForbidden).JSON(fiber.Map{"error": "Unauthorized to update advisor"})
					}
					return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
				}

				return c.JSON(fiber.Map{"message": "Advisor updated successfully"})
			})
		}
	}
}