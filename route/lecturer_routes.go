// File: BACKEND-UAS/route/lecturer_routes.go
package route

import (
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"BACKEND-UAS/middleware"
	"BACKEND-UAS/pgmongo/service"
)

// SetupLecturerRoutes sets up the lecturer routes with auth middleware
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
func SetupLecturerRoutes(app *fiber.App, lecturerSvc service.LecturerService, authM *middleware.AuthMiddlewareConfig) {
	v1 := app.Group("/api/v1")
	{
		lecturers := v1.Group("/lecturers")
		lecturers.Use(authM.AuthRequired()) // Apply auth middleware
		{
			// GET /api/v1/lecturers - Filtered by user type (from DB)
			lecturers.Get("/", func(c *fiber.Ctx) error {
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

				paginated, err := lecturerSvc.GetAllLecturers(c.Context(), userID, page, limit)
				if err != nil {
					return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
				}

				return c.JSON(paginated)
			})

			// @Summary Get lecturer's advisees
			// @Description Mengambil daftar mahasiswa bimbingan dosen (dengan access check: own, advisor, or admin)
			// @Tags Lecturers
			// @Accept json
			// @Produce json
			// @Param lecturer_id path string true "Lecturer ID (UUID)"
			// @Param user_id query string true "User ID (UUID) for access check"
			// @Success 200 {array} model.Student "List of advisees"
			// @Failure 400 {object} model.ErrorResponse "Access denied"
			// @Failure 404 {object} model.ErrorResponse "Lecturer not found"
			// @Failure 500 {object} model.ErrorResponse "Internal server error"
			// @Security ApiKeyAuth
			// @Router /lecturers/{lecturer_id}/advisees [get]
			lecturers.Get("/:id/advisees", func(c *fiber.Ctx) error {
				idStr := c.Params("id")
				id, err := uuid.Parse(idStr)
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

				advisees, err := lecturerSvc.GetAdvisees(c.Context(), id, userID)
				if err != nil {
					if err.Error() == "lecturer not found" || err.Error() == "access denied" {
						return c.Status(http.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
					}
					return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
				}

				return c.JSON(fiber.Map{"data": advisees})
			})
		}
	}
}