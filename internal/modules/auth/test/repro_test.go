package test

import (
	"io"
	"testing"
	"time"

	"github.com/Roisfaozi/queue-base/internal/mocking"
	mock_auth "github.com/Roisfaozi/queue-base/internal/modules/auth/test/mocks"
	"github.com/Roisfaozi/queue-base/internal/modules/auth/usecase"
	mock_org "github.com/Roisfaozi/queue-base/internal/modules/organization/test/mocks"
	mock_user "github.com/Roisfaozi/queue-base/internal/modules/user/test/mocks"
	"github.com/Roisfaozi/queue-base/pkg/jwt"
	"github.com/sirupsen/logrus"
)

func TestRepro(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Negative_Validation",
			category: "negative",
			run: func(t *testing.T) {
				jwtManager := jwt.NewJWTManager("secret", "refresh", 1, 1)
				log := logrus.New()
				log.SetOutput(io.Discard)

				_ = usecase.NewAuthUsecase(
					5,
					30*time.Minute,
					jwtManager,
					new(mock_auth.MockTokenRepository),
					new(mock_user.MockUserRepository),
					new(mock_org.MockOrganizationRepository),
					new(mocking.MockWithTransactionManager),
					log,
					new(mock_auth.MockNotificationPublisher),
					new(mock_auth.MockAuthzManager),
					new(mocking.MockTaskDistributor),
					new(mock_auth.MockTicketManager),
					nil,
				)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}
