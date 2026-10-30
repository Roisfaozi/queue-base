package test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Roisfaozi/queue-base/internal/modules/organization/test/mocks"
	"github.com/Roisfaozi/queue-base/internal/modules/organization/usecase"
	"github.com/go-redis/redismock/v9"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func newTestLogger() *logrus.Logger {
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	return log
}

func TestCachedOrgReader_ValidateMembership_CacheHit(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Positive_CacheHit",
			category: "positive",
			run: func(t *testing.T) {
				// Arrange
				db, mock := redismock.NewClientMock()
				mockRepo := new(mocks.MockOrganizationMemberRepository)
				log := newTestLogger()

				reader := usecase.NewCachedOrgReader(mockRepo, db, log)

				ctx := context.Background()
				orgID := "org-123"
				userID := "user-456"
				cacheKey := "org:member:org-123:user-456"

				// Mock Redis GET returning cached "1" (is member)
				mock.ExpectGet(cacheKey).SetVal("1")

				// Act
				isMember, err := reader.ValidateMembership(ctx, orgID, userID)

				// Assert
				assert.NoError(t, err)
				assert.True(t, isMember)
				assert.NoError(t, mock.ExpectationsWereMet())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestCachedOrgReader_ValidateMembership_CacheHit_NotMember(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Negative_CacheHit_NotMember",
			category: "negative",
			run: func(t *testing.T) {
				// Arrange
				db, mock := redismock.NewClientMock()
				mockRepo := new(mocks.MockOrganizationMemberRepository)
				log := newTestLogger()

				reader := usecase.NewCachedOrgReader(mockRepo, db, log)

				ctx := context.Background()
				orgID := "org-123"
				userID := "user-456"
				cacheKey := "org:member:org-123:user-456"

				// Mock Redis GET returning cached "0" (not member)
				mock.ExpectGet(cacheKey).SetVal("0")

				// Act
				isMember, err := reader.ValidateMembership(ctx, orgID, userID)

				// Assert
				assert.NoError(t, err)
				assert.False(t, isMember)
				assert.NoError(t, mock.ExpectationsWereMet())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestCachedOrgReader_ValidateMembership_CacheMiss_IsMember(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Positive_CacheMiss_IsMember",
			category: "positive",
			run: func(t *testing.T) {
				// Arrange
				db, mock := redismock.NewClientMock()
				ctx := context.Background()
				orgID := "org-123"
				userID := "user-456"
				membershipKey := "org:member:org-123:user-456"

				mockRepo := new(mocks.MockOrganizationMemberRepository)
				mockRepo.On("CheckMembership", ctx, orgID, userID).Return(true, nil)
				log := newTestLogger()

				reader := usecase.NewCachedOrgReader(mockRepo, db, log)

				// Mock Redis cache miss, then SET calls
				mock.ExpectGet(membershipKey).RedisNil()
				mock.ExpectSet(membershipKey, "1", 5*time.Minute).SetVal("OK")

				// Act
				isMember, err := reader.ValidateMembership(ctx, orgID, userID)

				// Assert
				assert.NoError(t, err)
				assert.True(t, isMember)
				assert.NoError(t, mock.ExpectationsWereMet())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestCachedOrgReader_ValidateMembership_CacheMiss_NotMember(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Negative_CacheMiss_NotMember",
			category: "negative",
			run: func(t *testing.T) {
				// Arrange
				db, mock := redismock.NewClientMock()
				ctx := context.Background()
				orgID := "org-123"
				userID := "user-456"
				membershipKey := "org:member:org-123:user-456"

				mockRepo := new(mocks.MockOrganizationMemberRepository)
				mockRepo.On("CheckMembership", ctx, orgID, userID).Return(false, nil)
				log := newTestLogger()

				reader := usecase.NewCachedOrgReader(mockRepo, db, log)

				// Mock Redis cache miss, then SET for negative result
				mock.ExpectGet(membershipKey).RedisNil()
				mock.ExpectSet(membershipKey, "0", 5*time.Minute).SetVal("OK")

				// Act
				isMember, err := reader.ValidateMembership(ctx, orgID, userID)

				// Assert
				assert.NoError(t, err)
				assert.False(t, isMember)
				assert.NoError(t, mock.ExpectationsWereMet())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestCachedOrgReader_ValidateMembership_DBError(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Negative_DBError",
			category: "negative",
			run: func(t *testing.T) {
				// Arrange
				db, mock := redismock.NewClientMock()
				ctx := context.Background()
				orgID := "org-123"
				userID := "user-456"
				membershipKey := "org:member:org-123:user-456"

				dbErr := errors.New("database connection error")
				mockRepo := new(mocks.MockOrganizationMemberRepository)
				mockRepo.On("CheckMembership", ctx, orgID, userID).Return(false, dbErr)
				log := newTestLogger()

				reader := usecase.NewCachedOrgReader(mockRepo, db, log)

				// Mock Redis cache miss
				mock.ExpectGet(membershipKey).RedisNil()

				// Act
				isMember, err := reader.ValidateMembership(ctx, orgID, userID)

				// Assert
				assert.Error(t, err)
				assert.Equal(t, dbErr, err)
				assert.False(t, isMember)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestCachedOrgReader_GetMemberRole_CacheHit(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Positive_CacheHit",
			category: "positive",
			run: func(t *testing.T) {
				// Arrange
				db, mock := redismock.NewClientMock()
				mockRepo := new(mocks.MockOrganizationMemberRepository)
				log := newTestLogger()

				reader := usecase.NewCachedOrgReader(mockRepo, db, log)

				ctx := context.Background()
				orgID := "org-123"
				userID := "user-456"
				roleKey := "org:role:org-123:user-456"

				// Mock Redis GET returning cached role
				mock.ExpectGet(roleKey).SetVal("admin")

				// Act
				role, err := reader.GetMemberRole(ctx, orgID, userID)

				// Assert
				assert.NoError(t, err)
				assert.Equal(t, "admin", role)
				assert.NoError(t, mock.ExpectationsWereMet())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestCachedOrgReader_GetMemberRole_CacheMiss(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Positive_CacheMiss",
			category: "positive",
			run: func(t *testing.T) {
				// Arrange
				db, mock := redismock.NewClientMock()
				ctx := context.Background()
				orgID := "org-123"
				userID := "user-456"
				roleKey := "org:role:org-123:user-456"

				mockRepo := new(mocks.MockOrganizationMemberRepository)
				mockRepo.On("GetMemberRole", ctx, orgID, userID).Return("member", nil)
				log := newTestLogger()

				reader := usecase.NewCachedOrgReader(mockRepo, db, log)

				// Mock Redis cache miss, then SET
				mock.ExpectGet(roleKey).RedisNil()
				mock.ExpectSet(roleKey, "member", usecase.MembershipCacheTTL).SetVal("OK")

				// Act
				role, err := reader.GetMemberRole(ctx, orgID, userID)

				// Assert
				assert.NoError(t, err)
				assert.Equal(t, "member", role)
				assert.NoError(t, mock.ExpectationsWereMet())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestCachedOrgReader_InvalidateMembershipCache(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Positive_InvalidateMembershipCache",
			category: "positive",
			run: func(t *testing.T) {
				// Arrange
				db, mock := redismock.NewClientMock()
				mockRepo := new(mocks.MockOrganizationMemberRepository)
				log := newTestLogger()

				reader := usecase.NewCachedOrgReader(mockRepo, db, log)

				ctx := context.Background()
				orgID := "org-123"
				userID := "user-456"
				membershipKey := "org:member:org-123:user-456"
				roleKey := "org:role:org-123:user-456"

				// Mock Redis pipeline DEL
				mock.ExpectDel(membershipKey).SetVal(1)
				mock.ExpectDel(roleKey).SetVal(1)

				// Act
				err := reader.InvalidateMembershipCache(ctx, orgID, userID)

				// Assert
				assert.NoError(t, err)
				assert.NoError(t, mock.ExpectationsWereMet())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestCachedOrgReader_InvalidateOrganizationCache(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Positive_InvalidateOrganizationCache",
			category: "positive",
			run: func(t *testing.T) {
				// Arrange
				db, mock := redismock.NewClientMock()
				mockRepo := new(mocks.MockOrganizationMemberRepository)
				log := newTestLogger()

				reader := usecase.NewCachedOrgReader(mockRepo, db, log)

				ctx := context.Background()
				orgID := "org-123"
				pattern := "org:*:org-123:*"
				statusKey := "nexusos:org_status:org-123"

				// Mock Redis SCAN - returns keys and cursor 0 (stop)
				mock.ExpectScan(0, pattern, 100).SetVal([]string{"key1", "key2"}, 0)

				// Mock Redis DEL for the found keys and status key
				mock.ExpectDel("key1", "key2").SetVal(2)
				mock.ExpectDel(statusKey).SetVal(1)

				// Act
				err := reader.InvalidateOrganizationCache(ctx, orgID)

				// Assert
				assert.NoError(t, err)
				assert.NoError(t, mock.ExpectationsWereMet())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}
