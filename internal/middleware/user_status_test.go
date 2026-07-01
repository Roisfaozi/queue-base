package middleware

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Roisfaozi/queue-base/internal/modules/user/entity"
	"github.com/Roisfaozi/queue-base/internal/modules/user/test/mocks"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupUserStatusTest() (*gin.Engine, *mocks.MockUserRepository, *logrus.Logger) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockRepo := new(mocks.MockUserRepository)
	log := logrus.New()
	log.SetOutput(io.Discard)

	return router, mockRepo, log
}

// ============================================================================
// Table Driven Tests
// ============================================================================

func TestUserStatusMiddleware_ActiveUser(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "ActiveUser",
			category: "positive",
			run: func(t *testing.T) {
				router, mockRepo, log := setupUserStatusTest()

				userID := "user-123"
				activeUser := &entity.User{
					ID:       userID,
					Username: "activeuser",
					Status:   entity.UserStatusActive,
				}

				// Mock repository
				mockRepo.On("FindByID", mock.Anything, userID).Return(activeUser, nil)

				router.Use(func(c *gin.Context) {
					c.Set("user_id", userID)
					c.Next()
				})
				router.Use(UserStatusMiddleware(mockRepo, log))

				router.GET("/test", func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"message": "success"})
				})

				req := httptest.NewRequest("GET", "/test", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				// Assert
				assert.Equal(t, http.StatusOK, w.Code)
				assert.Contains(t, w.Body.String(), "success")
				mockRepo.AssertExpectations(t)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestUserStatusMiddleware_SuspendedUser(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "SuspendedUser",
			category: "negative",
			run: func(t *testing.T) {
				router, mockRepo, log := setupUserStatusTest()

				userID := "user-456"
				suspendedUser := &entity.User{
					ID:       userID,
					Username: "suspendeduser",
					Status:   entity.UserStatusSuspended,
				}

				// Mock repository
				mockRepo.On("FindByID", mock.Anything, userID).Return(suspendedUser, nil)

				router.Use(func(c *gin.Context) {
					c.Set("user_id", userID)
					c.Next()
				})
				router.Use(UserStatusMiddleware(mockRepo, log))

				router.GET("/test", func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"message": "success"})
				})

				req := httptest.NewRequest("GET", "/test", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				// Assert
				assert.Equal(t, http.StatusForbidden, w.Code)
				assert.Contains(t, w.Body.String(), "suspended")
				mockRepo.AssertExpectations(t)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestUserStatusMiddleware_BannedUser(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "BannedUser",
			category: "negative",
			run: func(t *testing.T) {
				router, mockRepo, log := setupUserStatusTest()

				userID := "user-789"
				bannedUser := &entity.User{
					ID:       userID,
					Username: "banneduser",
					Status:   entity.UserStatusBanned,
				}

				// Mock repository
				mockRepo.On("FindByID", mock.Anything, userID).Return(bannedUser, nil)

				router.Use(func(c *gin.Context) {
					c.Set("user_id", userID)
					c.Next()
				})
				router.Use(UserStatusMiddleware(mockRepo, log))

				router.GET("/test", func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"message": "success"})
				})

				req := httptest.NewRequest("GET", "/test", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				// Assert
				assert.Equal(t, http.StatusForbidden, w.Code)
				assert.Contains(t, w.Body.String(), "banned")
				mockRepo.AssertExpectations(t)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestUserStatusMiddleware_NoUserContext(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "NoUserContext",
			category: "negative",
			run: func(t *testing.T) {
				router, mockRepo, log := setupUserStatusTest()

				// Don't set user_id in context
				router.Use(UserStatusMiddleware(mockRepo, log))

				router.GET("/test", func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"message": "success"})
				})

				req := httptest.NewRequest("GET", "/test", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				// Assert
				assert.Equal(t, http.StatusUnauthorized, w.Code)
				assert.Contains(t, w.Body.String(), "unauthorized")
				mockRepo.AssertNotCalled(t, "FindByID")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestUserStatusMiddleware_UserNotFoundInDB(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "UserNotFoundInDB",
			category: "edge",
			run: func(t *testing.T) {
				router, mockRepo, log := setupUserStatusTest()

				userID := "nonexistent-user"

				// Mock repository - User not found
				mockRepo.On("FindByID", mock.Anything, userID).Return(nil, errors.New("user not found"))

				router.Use(func(c *gin.Context) {
					c.Set("user_id", userID)
					c.Next()
				})
				router.Use(UserStatusMiddleware(mockRepo, log))

				router.GET("/test", func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"message": "success"})
				})

				req := httptest.NewRequest("GET", "/test", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				// Assert
				assert.Equal(t, http.StatusInternalServerError, w.Code)
				assert.Contains(t, w.Body.String(), "internal server error")
				mockRepo.AssertExpectations(t)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestUserStatusMiddleware_DatabaseError(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "DatabaseError",
			category: "edge",
			run: func(t *testing.T) {
				router, mockRepo, log := setupUserStatusTest()

				userID := "user-101"

				// Mock repository - Database error
				mockRepo.On("FindByID", mock.Anything, userID).Return(nil, errors.New("database connection lost"))

				router.Use(func(c *gin.Context) {
					c.Set("user_id", userID)
					c.Next()
				})
				router.Use(UserStatusMiddleware(mockRepo, log))

				router.GET("/test", func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"message": "success"})
				})

				req := httptest.NewRequest("GET", "/test", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				// Assert
				assert.Equal(t, http.StatusInternalServerError, w.Code)
				mockRepo.AssertExpectations(t)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}
