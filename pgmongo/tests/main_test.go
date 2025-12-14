// tests/main_test.go
package tests

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"
	"unsafe"

	jwtpkg "github.com/golang-jwt/jwt/v5"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"BACKEND-UAS/pgmongo/jwt"
	"BACKEND-UAS/pgmongo/model"
	"BACKEND-UAS/pgmongo/repository"
	"BACKEND-UAS/pgmongo/service"
)

// ======================= MOCK JWT SERVICE =======================
type mockJWTService struct {
	hash                string
	checkPasswordResult bool
	token               string
	err                 error
}

func (m *mockJWTService) HashPassword(password string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	if m.hash != "" {
		return m.hash, nil
	}
	return "hashed", nil
}

func (m *mockJWTService) GenerateToken(userID, roleID, role string) (string, error) {
	return m.token, m.err
}

func (m *mockJWTService) ValidateToken(tokenStr string) (*jwtpkg.Token, error) {
	return &jwtpkg.Token{Valid: true}, nil
}

func (m *mockJWTService) CheckPasswordHash(password, hash string) bool {
	return m.checkPasswordResult
}

var _ jwt.JWTService = (*mockJWTService)(nil)

// ======================= USER REPOSITORY TESTS (dengan sqlmock) =======================

func TestUserRepository_FindByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := repository.NewUserRepository(db)

	userID := uuid.New().String()
	now := time.Now().Truncate(time.Second)

	tests := []struct {
		name    string
		setup   func()
		wantErr bool
		want    *model.User
	}{
		{
			name: "success",
			setup: func() {
				rows := sqlmock.NewRows([]string{
					"id", "username", "email", "password_hash", "full_name",
					"role_id", "is_active", "created_at", "updated_at",
				}).AddRow(userID, "johndoe", "john@example.com", "hashedpass",
					"John Doe", "role1", true, now, now)

				mock.ExpectQuery(regexp.QuoteMeta(
					`SELECT id, username, email, password_hash, full_name, role_id, is_active, created_at, updated_at FROM users WHERE id=$1 LIMIT 1`,
				)).WithArgs(userID).WillReturnRows(rows)
			},
			want: &model.User{
				ID:           userID,
				Username:     "johndoe",
				Email:        "john@example.com",
				PasswordHash: "hashedpass",
				FullName:     "John Doe",
				RoleID:       "role1",
				IsActive:     true,
				CreatedAt:    now,
				UpdatedAt:    now,
			},
		},
		{
			name: "not_found",
			setup: func() {
				mock.ExpectQuery(regexp.QuoteMeta(
					`SELECT id, username, email, password_hash, full_name, role_id, is_active, created_at, updated_at FROM users WHERE id=$1 LIMIT 1`,
				)).WithArgs(userID).WillReturnError(sql.ErrNoRows)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			got, err := repo.FindByID(context.Background(), userID)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := repository.NewUserRepository(db)

	fixedTime := time.Now().Truncate(time.Second)
	user := &model.User{
		ID:           uuid.New().String(),
		Username:     "newuser",
		Email:        "new@example.com",
		PasswordHash: "hash123",
		FullName:     "New User",
		RoleID:       "role2",
		IsActive:     true,
		CreatedAt:    fixedTime,
		UpdatedAt:    fixedTime,
	}

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO users`)).
			WithArgs(
				user.ID, user.Username, user.Email, user.PasswordHash,
				user.FullName, user.RoleID, user.IsActive,
				user.CreatedAt, user.UpdatedAt,
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.Create(context.Background(), user)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// ======================= MOCK USER REPOSITORY UNTUK SERVICE =======================

type mockUserRepo struct {
	users        map[string]*model.User
	createCalled bool
	createdUser  *model.User
	findErr      error
}

func (m *mockUserRepo) FindByID(ctx context.Context, id string) (*model.User, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	user, exists := m.users[id]
	if !exists {
		return nil, sql.ErrNoRows
	}
	return user, nil
}

func (m *mockUserRepo) Create(ctx context.Context, user *model.User) error {
	m.createCalled = true
	m.createdUser = user
	if m.users == nil {
		m.users = make(map[string]*model.User)
	}
	m.users[user.ID] = user
	return nil
}

func (m *mockUserRepo) Delete(ctx context.Context, id string) error {
	delete(m.users, id)
	return nil
}

func (m *mockUserRepo) GetAll(ctx context.Context) ([]*model.User, error) { return nil, nil }
func (m *mockUserRepo) Update(ctx context.Context, id string, user *model.User) error { return nil }
func (m *mockUserRepo) UpdateRole(ctx context.Context, id, roleID string) error { return nil }
func (m *mockUserRepo) GetRoleNameByID(ctx context.Context, roleID string) (string, error) { return "", nil }
func (m *mockUserRepo) GetPermissionsByRoleID(ctx context.Context, roleID string) ([]string, error) {
	return nil, nil
}
func (m *mockUserRepo) FindByUsernameOrEmail(ctx context.Context, identifier string) (*model.User, error) {
	return nil, nil
}

var _ repository.UserRepository = (*mockUserRepo)(nil)

// ======================= USER SERVICE TESTS =======================

func TestUserService_Create(t *testing.T) {
	tests := []struct {
		name    string
		req     *service.CreateUserReq
		wantErr string
	}{
		{
			name: "success",
			req: &service.CreateUserReq{
				Username: "alice",
				Email:    "alice@example.com",
				Password: "secret123",
				FullName: "Alice Wonder",
				RoleID:   "admin-role",
			},
		},
		{
			name: "missing_required_fields",
			req: &service.CreateUserReq{
				Username: "",
				Email:    "bob@example.com",
				Password: "pass",
				FullName: "Bob",
				RoleID:   "role",
			},
			wantErr: "missing required fields",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := &mockUserRepo{}
			jwtSvc := &mockJWTService{hash: "hashed_password_123"}
			svc := service.NewUserService(userRepo, jwtSvc)

			user, err := svc.Create(context.Background(), tt.req)

			if tt.wantErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				assert.Nil(t, user)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, user)
			assert.True(t, userRepo.createCalled)
			assert.NotEmpty(t, user.ID)
			assert.Equal(t, tt.req.Username, user.Username)
			assert.Equal(t, tt.req.Email, user.Email)
			assert.Equal(t, "hashed_password_123", user.PasswordHash)
			assert.Equal(t, tt.req.FullName, user.FullName)
			assert.Equal(t, tt.req.RoleID, user.RoleID)
			assert.True(t, user.IsActive)
		})
	}
}

func TestUserService_GetByID(t *testing.T) {
	userRepo := &mockUserRepo{
		users: map[string]*model.User{
			"existing-id": {
				ID:       "existing-id",
				Username: "testuser",
			},
		},
	}
	jwtSvc := &mockJWTService{}
	svc := service.NewUserService(userRepo, jwtSvc)

	t.Run("success", func(t *testing.T) {
		user, err := svc.GetByID(context.Background(), "existing-id")
		assert.NoError(t, err)
		assert.Equal(t, "existing-id", user.ID)
	})

	t.Run("not_found", func(t *testing.T) {
		_, err := svc.GetByID(context.Background(), "non-existent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")
	})
}

func TestUserService_Delete(t *testing.T) {
	userRepo := &mockUserRepo{
		users: map[string]*model.User{
			"to-delete": {ID: "to-delete"},
		},
	}
	jwtSvc := &mockJWTService{}
	svc := service.NewUserService(userRepo, jwtSvc)

	t.Run("success", func(t *testing.T) {
		err := svc.Delete(context.Background(), "to-delete")
		assert.NoError(t, err)
		assert.Empty(t, userRepo.users)
	})

	t.Run("not_found", func(t *testing.T) {
		err := svc.Delete(context.Background(), "non-existent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")
	})
}

// ======================= AUTH SERVICE TESTS =======================

type mockUserRepositoryForAuth struct {
	userByIdentifier *model.User
	findErr          error
}

var _ repository.UserRepository = (*mockUserRepositoryForAuth)(nil)

func (m *mockUserRepositoryForAuth) FindByUsernameOrEmail(ctx context.Context, identifier string) (*model.User, error) {
	return m.userByIdentifier, m.findErr
}

func (m *mockUserRepositoryForAuth) FindByID(ctx context.Context, id string) (*model.User, error) { return nil, nil }
func (m *mockUserRepositoryForAuth) GetAll(ctx context.Context) ([]*model.User, error) { return nil, nil }
func (m *mockUserRepositoryForAuth) Create(ctx context.Context, user *model.User) error { return nil }
func (m *mockUserRepositoryForAuth) Update(ctx context.Context, id string, user *model.User) error { return nil }
func (m *mockUserRepositoryForAuth) Delete(ctx context.Context, id string) error { return nil }
func (m *mockUserRepositoryForAuth) UpdateRole(ctx context.Context, id, roleID string) error { return nil }
func (m *mockUserRepositoryForAuth) GetRoleNameByID(ctx context.Context, roleID string) (string, error) { return "", nil }
func (m *mockUserRepositoryForAuth) GetPermissionsByRoleID(ctx context.Context, roleID string) ([]string, error) {
	return nil, nil
}

func TestAuthService_LoginHandler(t *testing.T) {
	tests := []struct {
		name           string
		username       string
		password       string
		setupMocks     func(*mockUserRepositoryForAuth, *mockJWTService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:     "Success_Login",
			username: "admin",
			password: "password123",
			setupMocks: func(repo *mockUserRepositoryForAuth, jwtSvc *mockJWTService) {
				repo.userByIdentifier = &model.User{
					ID:           "user-123",
					Username:     "admin",
					PasswordHash: "anyhash",
					RoleID:       "role-admin",
					IsActive:     true,
				}
				jwtSvc.checkPasswordResult = true
				jwtSvc.token = "mock-jwt-token-xyz"
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"token":"mock-jwt-token-xyz"`,
		},
		{
			name:     "Wrong_Password",
			username: "admin",
			password: "wrong",
			setupMocks: func(repo *mockUserRepositoryForAuth, jwtSvc *mockJWTService) {
				repo.userByIdentifier = &model.User{IsActive: true}
				jwtSvc.checkPasswordResult = false
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "invalid credentials",
		},
		{
			name:           "User_Not_Found",
			username:       "unknown",
			password:       "any",
			setupMocks:     func(repo *mockUserRepositoryForAuth, jwtSvc *mockJWTService) { repo.findErr = assert.AnError },
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "invalid credentials",
		},
		{
			name:     "User_Inactive",
			username: "inactive",
			password: "password123",
			setupMocks: func(repo *mockUserRepositoryForAuth, jwtSvc *mockJWTService) {
				repo.userByIdentifier = &model.User{IsActive: false}
				jwtSvc.checkPasswordResult = true
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "account is inactive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := &mockUserRepositoryForAuth{}
			jwtSvc := &mockJWTService{}

			tt.setupMocks(userRepo, jwtSvc)

			authService := service.NewAuthService(userRepo, jwtSvc)

			app := fiber.New()
			app.Post("/api/v1/auth/login", authService.LoginHandler)

			payload := map[string]string{"username": tt.username, "password": tt.password}
			bodyBytes, _ := json.Marshal(payload)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req, -1)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectedBody != "" {
				body, _ := io.ReadAll(resp.Body)
				assert.Contains(t, string(body), tt.expectedBody)
			}
		})
	}
}

// ======================= ACHIEVEMENT SERVICE TESTS =======================

type mockAchievementPostgresRepo struct {
	GetAchievementReferenceByIDFunc          func(id uuid.UUID) (*model.AchievementReference, error)
	GetStudentByUserIDFunc                   func(userID uuid.UUID) (*model.Student, error)
	GetLecturerByUserIDFunc                  func(userID uuid.UUID) (*model.Lecturer, error)
	GetStudentIDsByAdvisorFunc               func(advisorID uuid.UUID) ([]uuid.UUID, error)
	GetAchievementReferencesByStudentIDsFunc func(studentIDs []uuid.UUID, status *string, page, limit int) (*model.PaginatedResponse[model.AchievementReference], error)
	GetAllAchievementReferencesFunc          func(status *string, page, limit int) (*model.PaginatedResponse[model.AchievementReference], error)
	CreateAchievementReferenceFunc           func(ref *model.AchievementReference) error
	SoftDeleteAchievementReferenceFunc       func(id uuid.UUID) error
	SubmitAchievementFunc                    func(id uuid.UUID) error
	VerifyAchievementFunc                    func(id uuid.UUID, verifiedBy uuid.UUID, rejectionNote *string) error
}

func (m *mockAchievementPostgresRepo) GetAchievementReferenceByID(id uuid.UUID) (*model.AchievementReference, error) {
	return m.GetAchievementReferenceByIDFunc(id)
}
func (m *mockAchievementPostgresRepo) GetStudentByUserID(userID uuid.UUID) (*model.Student, error) {
	return m.GetStudentByUserIDFunc(userID)
}
func (m *mockAchievementPostgresRepo) GetLecturerByUserID(userID uuid.UUID) (*model.Lecturer, error) {
	return m.GetLecturerByUserIDFunc(userID)
}
func (m *mockAchievementPostgresRepo) GetStudentIDsByAdvisor(advisorID uuid.UUID) ([]uuid.UUID, error) {
	return m.GetStudentIDsByAdvisorFunc(advisorID)
}
func (m *mockAchievementPostgresRepo) GetAchievementReferencesByStudentIDs(studentIDs []uuid.UUID, status *string, page, limit int) (*model.PaginatedResponse[model.AchievementReference], error) {
	return m.GetAchievementReferencesByStudentIDsFunc(studentIDs, status, page, limit)
}
func (m *mockAchievementPostgresRepo) GetAllAchievementReferences(status *string, page, limit int) (*model.PaginatedResponse[model.AchievementReference], error) {
	return m.GetAllAchievementReferencesFunc(status, page, limit)
}
func (m *mockAchievementPostgresRepo) CreateAchievementReference(ref *model.AchievementReference) error {
	return m.CreateAchievementReferenceFunc(ref)
}
func (m *mockAchievementPostgresRepo) SoftDeleteAchievementReference(id uuid.UUID) error {
	return m.SoftDeleteAchievementReferenceFunc(id)
}
func (m *mockAchievementPostgresRepo) SubmitAchievement(id uuid.UUID) error {
	return m.SubmitAchievementFunc(id)
}
func (m *mockAchievementPostgresRepo) VerifyAchievement(id uuid.UUID, verifiedBy uuid.UUID, rejectionNote *string) error {
	return m.VerifyAchievementFunc(id, verifiedBy, rejectionNote)
}

type mockAchievementMongoRepo struct {
	GetAchievementByIDFunc    func(mongoID string) (*model.Achievement, error)
	CreateAchievementFunc     func(ach *model.Achievement) error
	UpdateAchievementFunc     func(mongoID string, ach *model.Achievement) error
	SoftDeleteAchievementFunc func(mongoID string) error
	AddStatusHistoryFunc      func(mongoID string, history model.StatusHistory) error
	AddNotificationFunc       func(mongoID string, notif model.Notification) error
	UploadAttachmentFunc      func(mongoID string, file io.Reader, fileName, fileType string) (*model.Attachment, error)
}

func (m *mockAchievementMongoRepo) GetAchievementByID(mongoID string) (*model.Achievement, error) {
	return m.GetAchievementByIDFunc(mongoID)
}
func (m *mockAchievementMongoRepo) CreateAchievement(ach *model.Achievement) error {
	return m.CreateAchievementFunc(ach)
}
func (m *mockAchievementMongoRepo) UpdateAchievement(mongoID string, ach *model.Achievement) error {
	return m.UpdateAchievementFunc(mongoID, ach)
}
func (m *mockAchievementMongoRepo) SoftDeleteAchievement(mongoID string) error {
	return m.SoftDeleteAchievementFunc(mongoID)
}
func (m *mockAchievementMongoRepo) AddStatusHistory(mongoID string, history model.StatusHistory) error {
	return m.AddStatusHistoryFunc(mongoID, history)
}
func (m *mockAchievementMongoRepo) AddNotification(mongoID string, notif model.Notification) error {
	return m.AddNotificationFunc(mongoID, notif)
}
func (m *mockAchievementMongoRepo) UploadAttachment(mongoID string, file io.Reader, fileName, fileType string) (*model.Attachment, error) {
	if m.UploadAttachmentFunc != nil {
		return m.UploadAttachmentFunc(mongoID, file, fileName, fileType)
	}
	return nil, nil
}

type AchievementServiceTestSuite struct {
	suite.Suite
	service       *service.AchievementService
	pgRepo        *mockAchievementPostgresRepo
	mongoRepo     *mockAchievementMongoRepo
	studentID     uuid.UUID
	userID        uuid.UUID
	achievementID uuid.UUID
	mongoID       primitive.ObjectID
}

func (s *AchievementServiceTestSuite) SetupTest() {
	s.studentID = uuid.New()
	s.userID = uuid.New()
	s.achievementID = uuid.New()
	s.mongoID = primitive.NewObjectID()

	s.pgRepo = &mockAchievementPostgresRepo{}
	s.mongoRepo = &mockAchievementMongoRepo{}

	s.service = service.NewAchievementService(
		(*repository.AchievementRepository)(unsafe.Pointer(s.pgRepo)),
		(*repository.AchievementRepositoryMongo)(unsafe.Pointer(s.mongoRepo)),
	)
}

func TestRunAchievementServiceSuite(t *testing.T) {
	suite.Run(t, new(AchievementServiceTestSuite))
}

func (s *AchievementServiceTestSuite) TestGetUserAchievements_Mahasiswa() {
	status := "draft"
	page, limit := 1, 10

	s.pgRepo.GetStudentByUserIDFunc = func(userID uuid.UUID) (*model.Student, error) {
		assert.Equal(s.T(), s.userID, userID)
		return &model.Student{ID: s.studentID}, nil
	}

	expected := &model.PaginatedResponse[model.AchievementReference]{
		Data:       []model.AchievementReference{},
		Total:      5,
		Page:       page,
		Limit:      limit,
		TotalPages: 1,
	}

	s.pgRepo.GetAchievementReferencesByStudentIDsFunc = func(studentIDs []uuid.UUID, st *string, p, l int) (*model.PaginatedResponse[model.AchievementReference], error) {
		assert.Equal(s.T(), []uuid.UUID{s.studentID}, studentIDs)
		assert.Equal(s.T(), &status, st)
		return expected, nil
	}

	resp, err := s.service.GetUserAchievements(s.userID, "Mahasiswa", &status, page, limit)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), expected, resp)
}

func (s *AchievementServiceTestSuite) TestGetUserAchievements_Admin() {
	page, limit := 1, 10

	expected := &model.PaginatedResponse[model.AchievementReference]{
		Data:       []model.AchievementReference{},
		Total:      100,
		Page:       page,
		Limit:      limit,
		TotalPages: 10,
	}

	s.pgRepo.GetAllAchievementReferencesFunc = func(status *string, p, l int) (*model.PaginatedResponse[model.AchievementReference], error) {
		return expected, nil
	}

	resp, err := s.service.GetUserAchievements(s.userID, "Admin", nil, page, limit)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), expected, resp)
}

func (s *AchievementServiceTestSuite) TestCreateAchievement_Success() {
	ach := model.Achievement{Title: "Juara 1 Lomba"}

	s.pgRepo.GetStudentByUserIDFunc = func(userID uuid.UUID) (*model.Student, error) {
		return &model.Student{ID: s.studentID}, nil
	}

	s.mongoRepo.CreateAchievementFunc = func(a *model.Achievement) error {
		assert.Equal(s.T(), s.studentID, a.StudentID)
		a.ID = s.mongoID
		return nil
	}

	s.pgRepo.CreateAchievementReferenceFunc = func(ref *model.AchievementReference) error {
		assert.Equal(s.T(), s.studentID, ref.StudentID)
		assert.Equal(s.T(), s.mongoID.Hex(), ref.MongoAchievementID)
		assert.Equal(s.T(), "draft", ref.Status)
		return nil
	}

	ref, err := s.service.CreateAchievement(s.userID, ach)
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), ref)
	assert.Equal(s.T(), "draft", ref.Status)
}

func (s *AchievementServiceTestSuite) TestSubmitAchievement_Success() {
	ref := &model.AchievementReference{
		ID:                 s.achievementID,
		MongoAchievementID: s.mongoID.Hex(),
		Status:             "draft",
	}

	s.pgRepo.GetAchievementReferenceByIDFunc = func(id uuid.UUID) (*model.AchievementReference, error) {
		return ref, nil
	}

	s.pgRepo.SubmitAchievementFunc = func(id uuid.UUID) error {
		assert.Equal(s.T(), s.achievementID, id)
		return nil
	}

	s.mongoRepo.AddStatusHistoryFunc = func(mongoID string, history model.StatusHistory) error {
		assert.Equal(s.T(), s.mongoID.Hex(), mongoID)
		assert.Equal(s.T(), "submitted", history.Status)
		return nil
	}

	err := s.service.SubmitAchievement(s.achievementID, s.userID)
	assert.NoError(s.T(), err)
}

func (s *AchievementServiceTestSuite) TestVerifyAchievement_Success() {
	ref := &model.AchievementReference{
		ID:                 s.achievementID,
		MongoAchievementID: s.mongoID.Hex(),
		Status:             "submitted",
	}

	ach := &model.Achievement{Title: "Test Achievement", ID: s.mongoID}

	s.pgRepo.GetAchievementReferenceByIDFunc = func(id uuid.UUID) (*model.AchievementReference, error) {
		return ref, nil
	}

	s.pgRepo.VerifyAchievementFunc = func(id uuid.UUID, verifiedBy uuid.UUID, note *string) error {
		assert.Nil(s.T(), note)
		return nil
	}

	s.mongoRepo.AddStatusHistoryFunc = func(mongoID string, history model.StatusHistory) error {
		assert.Equal(s.T(), "verified", history.Status)
		return nil
	}

	s.mongoRepo.GetAchievementByIDFunc = func(mongoID string) (*model.Achievement, error) {
		return ach, nil
	}

	s.mongoRepo.AddNotificationFunc = func(mongoID string, notif model.Notification) error {
		assert.Contains(s.T(), notif.Message, "disetujui")
		return nil
	}

	err := s.service.VerifyAchievement(s.achievementID, s.userID)
	assert.NoError(s.T(), err)
}

func (s *AchievementServiceTestSuite) TestDeleteAchievement_OnlyDraft() {
	ref := &model.AchievementReference{
		ID:                 s.achievementID,
		MongoAchievementID: s.mongoID.Hex(),
		Status:             "draft",
	}

	s.pgRepo.GetAchievementReferenceByIDFunc = func(id uuid.UUID) (*model.AchievementReference, error) {
		return ref, nil
	}

	s.mongoRepo.SoftDeleteAchievementFunc = func(mongoID string) error { return nil }
	s.pgRepo.SoftDeleteAchievementReferenceFunc = func(id uuid.UUID) error { return nil }
	s.mongoRepo.AddStatusHistoryFunc = func(mongoID string, history model.StatusHistory) error { return nil }

	err := s.service.DeleteAchievement(s.achievementID, s.userID)
	assert.NoError(s.T(), err)
}

func (s *AchievementServiceTestSuite) TestDeleteAchievement_NonDraft_ReturnsError() {
	ref := &model.AchievementReference{
		ID:     s.achievementID,
		Status: "submitted",
	}

	s.pgRepo.GetAchievementReferenceByIDFunc = func(id uuid.UUID) (*model.AchievementReference, error) {
		return ref, nil
	}

	err := s.service.DeleteAchievement(s.achievementID, s.userID)
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "can only delete draft")
}