package test

import (
	"context"
	"github.com/stretchr/testify/mock"
	"io"
	"testing"
	"time"

	mocking "github.com/Roisfaozi/queue-base/internal/mocking"
	"github.com/Roisfaozi/queue-base/internal/modules/auth/usecase"
	"github.com/Roisfaozi/queue-base/internal/modules/user/entity"
	"github.com/Roisfaozi/queue-base/pkg/jwt"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"

	mock_auth "github.com/Roisfaozi/queue-base/internal/modules/auth/test/mocks"
	mock_org "github.com/Roisfaozi/queue-base/internal/modules/organization/test/mocks"
	mock_user "github.com/Roisfaozi/queue-base/internal/modules/user/test/mocks"
	"github.com/Roisfaozi/queue-base/pkg/sso"
)

const (
	TestAccessSecret  = "test-access-secret"
	TestRefreshSecret = "test-refresh-secret"
	TestUserID        = "user-test-id"
	TestUsername      = "testuser"
	TestRole          = "role:user"
)

type testDependencies struct {
	jwtManager      *jwt.JWTManager
	tokenRepo       *mock_auth.MockTokenRepository
	userRepo        *mock_user.MockUserRepository
	orgRepo         *mock_org.MockOrganizationRepository
	tm              *mocking.MockWithTransactionManager
	publisher       *mock_auth.MockNotificationPublisher
	authz           *mock_auth.MockAuthzManager
	validate        *validator.Validate
	log             *logrus.Logger
	taskDistributor *mocking.MockTaskDistributor
	ticketManager   *mock_auth.MockTicketManager
	ssoProviders    map[string]sso.Provider
}

func setupTest(t *testing.T) (usecase.AuthUseCase, *testDependencies) {
	jwtManager := jwt.NewJWTManager(TestAccessSecret, TestRefreshSecret, 15*time.Minute, 24*time.Hour)

	deps := &testDependencies{
		jwtManager:      jwtManager,
		tokenRepo:       new(mock_auth.MockTokenRepository),
		userRepo:        new(mock_user.MockUserRepository),
		orgRepo:         new(mock_org.MockOrganizationRepository),
		tm:              new(mocking.MockWithTransactionManager),
		publisher:       new(mock_auth.MockNotificationPublisher),
		authz:           new(mock_auth.MockAuthzManager),
		validate:        validator.New(),
		log:             logrus.New(),
		taskDistributor: new(mocking.MockTaskDistributor),
		ticketManager:   new(mock_auth.MockTicketManager),
		ssoProviders:    make(map[string]sso.Provider),
	}

	deps.log.SetOutput(io.Discard)

	authService := usecase.NewAuthUsecase(
		5,
		30*time.Minute,
		deps.jwtManager,
		deps.tokenRepo,
		deps.userRepo,
		deps.orgRepo,
		deps.tm,
		deps.log,
		deps.publisher,
		deps.authz,
		deps.taskDistributor,
		deps.ticketManager,
		deps.ssoProviders,
	)

	return authService, deps
}

func createTestUser(password string) (*entity.User, string) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return &entity.User{
		ID:       TestUserID,
		Username: TestUsername,
		Name:     "Test User",
		Password: string(hashedPassword),
		Email:    "test@example.com",
		Status:   entity.UserStatusActive,
	}, password
}

func simulateWithinTransaction(deps *testDependencies, errToReturn error) {
	deps.tm.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			ctx := args.Get(0).(context.Context)
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(errToReturn)
}
