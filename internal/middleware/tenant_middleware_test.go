package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Roisfaozi/queue-base/internal/modules/organization/entity"
	orgMocks "github.com/Roisfaozi/queue-base/internal/modules/organization/test/mocks"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newMockOrganizationRepository() *orgMocks.MockOrganizationRepository {
	m := new(orgMocks.MockOrganizationRepository)
	m.On("FindByID", mock.Anything, mock.Anything).Maybe().Return(func(_ context.Context, id string) *entity.Organization {
		return &entity.Organization{ID: id}
	}, nil)
	return m
}

func setupTestRouter(middleware *TenantMiddleware) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	return r
}

// ============================================================================
// Table Driven Tests
// ============================================================================

func TestTenantMiddleware_RequireOrganization_Success(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "RequireOrganization_Success",
			category: "positive",
			run: func(t *testing.T) {
				// Setup mocks
				mockOrgRepo := newMockOrganizationRepository()
				mockReader := new(orgMocks.MockIOrganizationReader)
				log := logrus.New()

				middleware := NewTenantMiddleware(mockOrgRepo, mockReader, log)

				// Setup expectations
				orgID := "org-123"
				userID := "user-456"

				// Mock reader validation
				mockReader.On("ValidateMembership", mock.Anything, orgID, userID).Return(true, nil)
				mockReader.On("GetMemberRole", mock.Anything, orgID, userID).Return("admin", nil)

				// Setup router
				r := setupTestRouter(middleware)
				r.Use(func(c *gin.Context) {
					c.Set("user_id", userID) // Simulate AuthMiddleware
					c.Next()
				})
				r.Use(middleware.RequireOrganization())
				r.GET("/test", func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{
						"organization_id": c.GetString("organization_id"),
						"role":            c.GetString("member_role"),
					})
				})

				// Make request
				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				req.Header.Set(OrgIDHeader, orgID)
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)

				// Assertions
				assert.Equal(t, http.StatusOK, w.Code)
				assert.Contains(t, w.Body.String(), orgID)
				assert.Contains(t, w.Body.String(), "admin")
				mockReader.AssertExpectations(t)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestTenantMiddleware_RequireOrganization_MissingOrgHeader(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "RequireOrganization_MissingOrgHeader",
			category: "negative",
			run: func(t *testing.T) {
				mockOrgRepo := newMockOrganizationRepository()
				mockReader := new(orgMocks.MockIOrganizationReader)
				log := logrus.New()

				middleware := NewTenantMiddleware(mockOrgRepo, mockReader, log)

				r := setupTestRouter(middleware)
				r.Use(func(c *gin.Context) {
					c.Set("user_id", "user-123")
					c.Next()
				})
				r.Use(middleware.RequireOrganization())
				r.GET("/test", func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"ok": true})
				})

				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				// No org header set
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)

				assert.Equal(t, http.StatusBadRequest, w.Code)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestTenantMiddleware_RequireOrganization_NotAuthenticated(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "RequireOrganization_NotAuthenticated",
			category: "negative",
			run: func(t *testing.T) {
				mockOrgRepo := newMockOrganizationRepository()
				mockReader := new(orgMocks.MockIOrganizationReader)
				log := logrus.New()

				middleware := NewTenantMiddleware(mockOrgRepo, mockReader, log)

				r := setupTestRouter(middleware)
				// No auth middleware - user_id not set
				r.Use(middleware.RequireOrganization())
				r.GET("/test", func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"ok": true})
				})

				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				req.Header.Set(OrgIDHeader, "org-123")
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)

				assert.Equal(t, http.StatusUnauthorized, w.Code)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestTenantMiddleware_RequireOrganization_NotMember(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "RequireOrganization_NotMember",
			category: "negative",
			run: func(t *testing.T) {
				mockOrgRepo := newMockOrganizationRepository()
				mockReader := new(orgMocks.MockIOrganizationReader)
				log := logrus.New()

				middleware := NewTenantMiddleware(mockOrgRepo, mockReader, log)

				orgID := "org-123"
				userID := "user-456"

				// Mock reader returns not member
				mockReader.On("ValidateMembership", mock.Anything, orgID, userID).Return(false, nil)

				r := setupTestRouter(middleware)
				r.Use(func(c *gin.Context) {
					c.Set("user_id", userID)
					c.Next()
				})
				r.Use(middleware.RequireOrganization())
				r.GET("/test", func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"ok": true})
				})

				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				req.Header.Set(OrgIDHeader, orgID)
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)

				assert.Equal(t, http.StatusForbidden, w.Code)
				mockReader.AssertExpectations(t)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestTenantMiddleware_RequireOrganization_SlugLookup(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "RequireOrganization_SlugLookup",
			category: "positive",
			run: func(t *testing.T) {
				mockOrgRepo := newMockOrganizationRepository()
				mockReader := new(orgMocks.MockIOrganizationReader)
				log := logrus.New()

				middleware := NewTenantMiddleware(mockOrgRepo, mockReader, log)

				orgID := "org-123"
				orgSlug := "my-org"
				userID := "user-456"

				// Slug lookup returns org
				mockOrgRepo.On("FindBySlug", mock.Anything, orgSlug).Return(&entity.Organization{
					ID:   orgID,
					Slug: orgSlug,
				}, nil)
				// Membership check via reader
				mockReader.On("ValidateMembership", mock.Anything, orgID, userID).Return(true, nil)
				mockReader.On("GetMemberRole", mock.Anything, orgID, userID).Return("owner", nil)

				r := setupTestRouter(middleware)
				r.Use(func(c *gin.Context) {
					c.Set("user_id", userID)
					c.Next()
				})
				r.Use(middleware.RequireOrganization())
				r.GET("/test", func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"organization_id": c.GetString("organization_id")})
				})

				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				req.Header.Set(OrgSlugHeader, orgSlug) // Use slug instead of ID
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)

				assert.Equal(t, http.StatusOK, w.Code)
				assert.Contains(t, w.Body.String(), orgID)
				mockOrgRepo.AssertExpectations(t)
				mockReader.AssertExpectations(t)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestTenantMiddleware_RequireOrganization_OrganizationRouteParamID(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "RequireOrganization_OrganizationRouteParamID",
			category: "positive",
			run: func(t *testing.T) {
				mockOrgRepo := newMockOrganizationRepository()
				mockReader := new(orgMocks.MockIOrganizationReader)
				log := logrus.New()

				middleware := NewTenantMiddleware(mockOrgRepo, mockReader, log)

				orgID := "org-123"
				userID := "user-456"

				mockReader.On("ValidateMembership", mock.Anything, orgID, userID).Return(true, nil)
				mockReader.On("GetMemberRole", mock.Anything, orgID, userID).Return("member", nil)

				r := setupTestRouter(middleware)
				r.Use(func(c *gin.Context) {
					c.Set("user_id", userID)
					c.Next()
				})
				r.Use(middleware.RequireOrganization())
				r.GET("/api/v1/organizations/:id/presence", func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"organization_id": c.GetString("organization_id")})
				})

				req := httptest.NewRequest(http.MethodGet, "/api/v1/organizations/"+orgID+"/presence", nil)
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)

				assert.Equal(t, http.StatusOK, w.Code)
				assert.Contains(t, w.Body.String(), orgID)
				mockReader.AssertExpectations(t)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestTenantMiddleware_RequireOrganization_OrgNotFound(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "RequireOrganization_OrgNotFound",
			category: "negative",
			run: func(t *testing.T) {
				mockOrgRepo := newMockOrganizationRepository()
				mockReader := new(orgMocks.MockIOrganizationReader)
				log := logrus.New()

				middleware := NewTenantMiddleware(mockOrgRepo, mockReader, log)

				orgSlug := "non-existent-org"
				userID := "user-456"

				// Slug lookup returns nil (not found)
				mockOrgRepo.On("FindBySlug", mock.Anything, orgSlug).Return(nil, nil)

				r := setupTestRouter(middleware)
				r.Use(func(c *gin.Context) {
					c.Set("user_id", userID)
					c.Next()
				})
				r.Use(middleware.RequireOrganization())
				r.GET("/test", func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"ok": true})
				})

				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				req.Header.Set(OrgSlugHeader, orgSlug)
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)

				assert.Equal(t, http.StatusNotFound, w.Code)
				mockOrgRepo.AssertExpectations(t)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestTenantMiddleware_RequireOrganization_Error(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "RequireOrganization_Error",
			category: "negative",
			run: func(t *testing.T) {
				mockOrgRepo := newMockOrganizationRepository()
				mockReader := new(orgMocks.MockIOrganizationReader)
				log := logrus.New()

				middleware := NewTenantMiddleware(mockOrgRepo, mockReader, log)

				orgID := "org-123"
				userID := "user-456"

				// Reader returns error
				mockReader.On("ValidateMembership", mock.Anything, orgID, userID).Return(false, errors.New("reader error"))

				r := setupTestRouter(middleware)
				r.Use(func(c *gin.Context) {
					c.Set("user_id", userID)
					c.Next()
				})
				r.Use(middleware.RequireOrganization())
				r.GET("/test", func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"ok": true})
				})

				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				req.Header.Set(OrgIDHeader, orgID)
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)

				assert.Equal(t, http.StatusInternalServerError, w.Code)
				mockReader.AssertExpectations(t)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestGetOrganizationIDFromContext(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "GetOrganizationIDFromContext_Tests",
			category: "edge",
			run: func(t *testing.T) {
				gin.SetMode(gin.TestMode)

				t.Run("exists and valid", func(t *testing.T) {
					c, _ := gin.CreateTestContext(httptest.NewRecorder())
					c.Set("organization_id", "org-123")

					orgID, ok := GetOrganizationIDFromContext(c)
					assert.True(t, ok)
					assert.Equal(t, "org-123", orgID)
				})

				t.Run("not exists", func(t *testing.T) {
					c, _ := gin.CreateTestContext(httptest.NewRecorder())

					orgID, ok := GetOrganizationIDFromContext(c)
					assert.False(t, ok)
					assert.Empty(t, orgID)
				})

				t.Run("wrong type", func(t *testing.T) {
					c, _ := gin.CreateTestContext(httptest.NewRecorder())
					c.Set("organization_id", 123) // int instead of string

					orgID, ok := GetOrganizationIDFromContext(c)
					assert.False(t, ok)
					assert.Empty(t, orgID)
				})
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestInvalidateMembershipCache(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "InvalidateMembershipCache",
			category: "positive",
			run: func(t *testing.T) {
				mockOrgRepo := newMockOrganizationRepository()
				mockReader := new(orgMocks.MockIOrganizationReader)
				log := logrus.New()

				middleware := NewTenantMiddleware(mockOrgRepo, mockReader, log)

				orgID := "org-123"
				userID := "user-456"

				mockReader.On("InvalidateMembershipCache", mock.Anything, orgID, userID).Return(nil)

				err := middleware.InvalidateMembershipCache(context.Background(), orgID, userID)
				assert.NoError(t, err)
				mockReader.AssertExpectations(t)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

// Unused import guard
var _ = time.Second

func TestTenantMiddleware_OptionalOrganization(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "OptionalOrganization_Tests",
			category: "edge",
			run: func(t *testing.T) {
				mockOrgRepo := newMockOrganizationRepository()
				mockReader := new(orgMocks.MockIOrganizationReader)
				log := logrus.New()

				middleware := NewTenantMiddleware(mockOrgRepo, mockReader, log)

				t.Run("no user context", func(t *testing.T) {
					r := setupTestRouter(middleware)
					r.Use(middleware.OptionalOrganization())
					r.GET("/test", func(c *gin.Context) {
						_, exists := c.Get("organization_id")
						assert.False(t, exists)
						c.Status(http.StatusOK)
					})

					req := httptest.NewRequest(http.MethodGet, "/test", nil)
					w := httptest.NewRecorder()
					r.ServeHTTP(w, req)
					assert.Equal(t, http.StatusOK, w.Code)
				})

				t.Run("no org specified", func(t *testing.T) {
					r := setupTestRouter(middleware)
					r.Use(func(c *gin.Context) {
						c.Set("user_id", "user123")
						c.Next()
					})
					r.Use(middleware.OptionalOrganization())
					r.GET("/test", func(c *gin.Context) {
						_, exists := c.Get("organization_id")
						assert.False(t, exists)
						c.Status(http.StatusOK)
					})

					req := httptest.NewRequest(http.MethodGet, "/test", nil)
					w := httptest.NewRecorder()
					r.ServeHTTP(w, req)
					assert.Equal(t, http.StatusOK, w.Code)
				})

				t.Run("slug lookup fails", func(t *testing.T) {
					mockOrgRepo := newMockOrganizationRepository()
					mockOrgRepo.On("FindBySlug", mock.Anything, "bad-slug").Return(nil, errors.New("not found"))
					m := NewTenantMiddleware(mockOrgRepo, mockReader, log)

					r := setupTestRouter(m)
					r.Use(func(c *gin.Context) {
						c.Set("user_id", "user123")
						c.Next()
					})
					r.Use(m.OptionalOrganization())
					r.GET("/test", func(c *gin.Context) {
						_, exists := c.Get("organization_id")
						assert.False(t, exists)
						c.Status(http.StatusOK)
					})

					req := httptest.NewRequest(http.MethodGet, "/test", nil)
					req.Header.Set(OrgSlugHeader, "bad-slug")
					w := httptest.NewRecorder()
					r.ServeHTTP(w, req)
					assert.Equal(t, http.StatusInternalServerError, w.Code)
				})

				t.Run("slug lookup returns nil", func(t *testing.T) {
					mockOrgRepo := newMockOrganizationRepository()
					mockOrgRepo.On("FindBySlug", mock.Anything, "bad-slug2").Return(nil, nil)
					m := NewTenantMiddleware(mockOrgRepo, mockReader, log)

					r := setupTestRouter(m)
					r.Use(func(c *gin.Context) {
						c.Set("user_id", "user123")
						c.Next()
					})
					r.Use(m.OptionalOrganization())
					r.GET("/test", func(c *gin.Context) {
						_, exists := c.Get("organization_id")
						assert.False(t, exists)
						c.Status(http.StatusOK)
					})

					req := httptest.NewRequest(http.MethodGet, "/test", nil)
					req.Header.Set(OrgSlugHeader, "bad-slug2")
					w := httptest.NewRecorder()
					r.ServeHTTP(w, req)
					assert.Equal(t, http.StatusNotFound, w.Code)
				})

				t.Run("membership validation fails", func(t *testing.T) {
					mockReader := new(orgMocks.MockIOrganizationReader)
					mockReader.On("ValidateMembership", mock.Anything, "org123", "user123").Return(false, errors.New("error"))
					m := NewTenantMiddleware(mockOrgRepo, mockReader, log)

					r := setupTestRouter(m)
					r.Use(func(c *gin.Context) {
						c.Set("user_id", "user123")
						c.Next()
					})
					r.Use(m.OptionalOrganization())
					r.GET("/test", func(c *gin.Context) {
						_, exists := c.Get("organization_id")
						assert.False(t, exists)
						c.Status(http.StatusOK)
					})

					req := httptest.NewRequest(http.MethodGet, "/test", nil)
					req.Header.Set(OrgIDHeader, "org123")
					w := httptest.NewRecorder()
					r.ServeHTTP(w, req)
					assert.Equal(t, http.StatusInternalServerError, w.Code)
				})

				t.Run("not a member", func(t *testing.T) {
					mockReader := new(orgMocks.MockIOrganizationReader)
					mockReader.On("ValidateMembership", mock.Anything, "org123", "user123").Return(false, nil)
					m := NewTenantMiddleware(mockOrgRepo, mockReader, log)

					r := setupTestRouter(m)
					r.Use(func(c *gin.Context) {
						c.Set("user_id", "user123")
						c.Next()
					})
					r.Use(m.OptionalOrganization())
					r.GET("/test", func(c *gin.Context) {
						_, exists := c.Get("organization_id")
						assert.False(t, exists)
						c.Status(http.StatusOK)
					})

					req := httptest.NewRequest(http.MethodGet, "/test", nil)
					req.Header.Set(OrgIDHeader, "org123")
					w := httptest.NewRecorder()
					r.ServeHTTP(w, req)
					assert.Equal(t, http.StatusForbidden, w.Code)
				})

				t.Run("success with org ID", func(t *testing.T) {
					mockReader := new(orgMocks.MockIOrganizationReader)
					mockReader.On("ValidateMembership", mock.Anything, "org123", "user123").Return(true, nil)
					mockReader.On("GetMemberRole", mock.Anything, "org123", "user123").Return("admin", nil)
					m := NewTenantMiddleware(mockOrgRepo, mockReader, log)

					r := setupTestRouter(m)
					r.Use(func(c *gin.Context) {
						c.Set("user_id", "user123")
						c.Next()
					})
					r.Use(m.OptionalOrganization())
					r.GET("/test", func(c *gin.Context) {
						orgID, _ := c.Get("organization_id")
						role, _ := c.Get("member_role")
						assert.Equal(t, "org123", orgID)
						assert.Equal(t, "admin", role)
						c.Status(http.StatusOK)
					})

					req := httptest.NewRequest(http.MethodGet, "/test", nil)
					req.Header.Set(OrgIDHeader, "org123")
					w := httptest.NewRecorder()
					r.ServeHTTP(w, req)
					assert.Equal(t, http.StatusOK, w.Code)
				})
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestTenantMiddleware_RequireOrgRole(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "RequireOrgRole_Tests",
			category: "edge",
			run: func(t *testing.T) {
				mockOrgRepo := newMockOrganizationRepository()
				mockReader := new(orgMocks.MockIOrganizationReader)
				log := logrus.New()

				middleware := NewTenantMiddleware(mockOrgRepo, mockReader, log)

				t.Run("role not found", func(t *testing.T) {
					r := setupTestRouter(middleware)
					r.Use(middleware.RequireOrgRole("admin"))
					r.GET("/test", func(c *gin.Context) {
						c.Status(http.StatusOK)
					})

					req := httptest.NewRequest(http.MethodGet, "/test", nil)
					w := httptest.NewRecorder()
					r.ServeHTTP(w, req)
					assert.Equal(t, http.StatusForbidden, w.Code)
				})

				t.Run("insufficient permissions", func(t *testing.T) {
					r := setupTestRouter(middleware)
					r.Use(func(c *gin.Context) {
						c.Set("member_role", "member")
						c.Next()
					})
					r.Use(middleware.RequireOrgRole("admin", "owner"))
					r.GET("/test", func(c *gin.Context) {
						c.Status(http.StatusOK)
					})

					req := httptest.NewRequest(http.MethodGet, "/test", nil)
					w := httptest.NewRecorder()
					r.ServeHTTP(w, req)
					assert.Equal(t, http.StatusForbidden, w.Code)
				})

				t.Run("sufficient permissions", func(t *testing.T) {
					r := setupTestRouter(middleware)
					r.Use(func(c *gin.Context) {
						c.Set("member_role", "admin")
						c.Next()
					})
					r.Use(middleware.RequireOrgRole("member", "admin"))
					r.GET("/test", func(c *gin.Context) {
						c.Status(http.StatusOK)
					})

					req := httptest.NewRequest(http.MethodGet, "/test", nil)
					w := httptest.NewRecorder()
					r.ServeHTTP(w, req)
					assert.Equal(t, http.StatusOK, w.Code)
				})

				t.Run("owner permissions (hierarchy)", func(t *testing.T) {
					r := setupTestRouter(middleware)
					r.Use(func(c *gin.Context) {
						c.Set("member_role", "owner")
						c.Next()
					})
					r.Use(middleware.RequireOrgRole("admin"))
					r.GET("/test", func(c *gin.Context) {
						c.Status(http.StatusOK)
					})

					req := httptest.NewRequest(http.MethodGet, "/test", nil)
					w := httptest.NewRecorder()
					r.ServeHTTP(w, req)
					assert.Equal(t, http.StatusOK, w.Code)
				})
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestGetMemberRoleFromContext(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "GetMemberRoleFromContext_Tests",
			category: "edge",
			run: func(t *testing.T) {
				gin.SetMode(gin.TestMode)

				t.Run("exists and valid", func(t *testing.T) {
					c, _ := gin.CreateTestContext(httptest.NewRecorder())
					c.Set("member_role", "admin")

					val, ok := GetMemberRoleFromContext(c)
					assert.True(t, ok)
					assert.Equal(t, "admin", val)
				})

				t.Run("not exists", func(t *testing.T) {
					c, _ := gin.CreateTestContext(httptest.NewRecorder())

					val, ok := GetMemberRoleFromContext(c)
					assert.False(t, ok)
					assert.Empty(t, val)
				})

				t.Run("wrong type", func(t *testing.T) {
					c, _ := gin.CreateTestContext(httptest.NewRecorder())
					c.Set("member_role", 123)

					val, ok := GetMemberRoleFromContext(c)
					assert.False(t, ok)
					assert.Empty(t, val)
				})

				t.Run("empty string", func(t *testing.T) {
					c, _ := gin.CreateTestContext(httptest.NewRecorder())
					c.Set("member_role", "")

					val, ok := GetMemberRoleFromContext(c)
					assert.False(t, ok)
					assert.Empty(t, val)
				})
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}
