//go:build e2e
// +build e2e

package api

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	userEntity "github.com/Roisfaozi/queue-base/internal/modules/user/entity"
	"github.com/Roisfaozi/queue-base/tests/e2e/setup"
	"github.com/Roisfaozi/queue-base/tests/fixtures"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func loginForStats(t *testing.T, server *setup.TestServer) string {
	f := fixtures.NewUserFactory(server.DB)
	hash, _ := bcrypt.GenerateFromPassword([]byte("StatsPass123!"), bcrypt.DefaultCost)

	uniqueSuffix := fmt.Sprintf("%d", time.Now().UnixNano())

	f.Create(func(u *userEntity.User) {
		u.Username = "stats_user_" + uniqueSuffix
		u.Email = "stats_" + uniqueSuffix + "@test.com"
		u.Password = string(hash)
	})

	resp := server.Client.POST("/api/v1/auth/login", map[string]any{
		"username": "stats_user_" + uniqueSuffix,
		"password": "StatsPass123!",
	})
	assert.Equal(t, 200, resp.StatusCode)

	var loginRes struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	resp.JSON(&loginRes)
	return loginRes.Data.AccessToken
}

func TestStatsE2E(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()

	token := loginForStats(t, server)

	tests := []struct {
		name           string
		endpoint       string
		method         string
		token          string
		expectedStatus int
		validateFunc   func(t *testing.T, resp *setup.Response)
	}{
		{
			name:           "Success_GetSummary",
			endpoint:       "/api/v1/stats/summary?timeframe=week",
			method:         http.MethodGet,
			token:          token,
			expectedStatus: http.StatusOK,
			validateFunc: func(t *testing.T, resp *setup.Response) {
				var res struct {
					Data struct {
						TotalUsers    int     `json:"total_users"`
						ActiveUsers   int     `json:"active_users"`
						Revenue       float64 `json:"revenue"`
						NewSignups    int     `json:"new_signups"`
						PreviousUsers int     `json:"previous_users"`
					} `json:"data"`
				}
				resp.JSON(&res)
				// Basic validation that structure is present
				assert.GreaterOrEqual(t, res.Data.TotalUsers, 0)
				assert.GreaterOrEqual(t, res.Data.ActiveUsers, 0)
			},
		},
		{
			name:           "Error_SummaryInvalidTimeframe",
			endpoint:       "/api/v1/stats/summary?timeframe=invalid",
			method:         http.MethodGet,
			token:          token,
			expectedStatus: http.StatusBadRequest,
			validateFunc: func(t *testing.T, resp *setup.Response) {
				var res map[string]any
				resp.JSON(&res)
				assert.Equal(t, "invalid timeframe", res["message"])
			},
		},
		{
			name:           "Success_GetActivity",
			endpoint:       "/api/v1/stats/activity?timeframe=month",
			method:         http.MethodGet,
			token:          token,
			expectedStatus: http.StatusOK,
			validateFunc: func(t *testing.T, resp *setup.Response) {
				var res struct {
					Data struct {
						TimeSeries []struct {
							Date  string `json:"date"`
							Value int    `json:"value"`
						} `json:"time_series"`
						Trend float64 `json:"trend"`
					} `json:"data"`
				}
				resp.JSON(&res)
				assert.NotNil(t, res.Data.TimeSeries)
			},
		},
		{
			name:           "Error_ActivityInvalidTimeframe",
			endpoint:       "/api/v1/stats/activity?timeframe=invalid",
			method:         http.MethodGet,
			token:          token,
			expectedStatus: http.StatusBadRequest,
			validateFunc: func(t *testing.T, resp *setup.Response) {
				var res map[string]any
				resp.JSON(&res)
				assert.Equal(t, "invalid timeframe", res["message"])
			},
		},
		{
			name:           "Success_GetInsights",
			endpoint:       "/api/v1/stats/insights",
			method:         http.MethodGet,
			token:          token,
			expectedStatus: http.StatusOK,
			validateFunc: func(t *testing.T, resp *setup.Response) {
				var res struct {
					Data []struct {
						Category string `json:"category"`
						Title    string `json:"title"`
						Value    string `json:"value"`
						Type     string `json:"type"`
					} `json:"data"`
				}
				resp.JSON(&res)
				assert.NotNil(t, res.Data)
				if len(res.Data) > 0 {
					assert.NotEmpty(t, res.Data[0].Category)
				}
			},
		},
		{
			name:           "Error_SummaryUnauthorized",
			endpoint:       "/api/v1/stats/summary?timeframe=week",
			method:         http.MethodGet,
			token:          "", // No token
			expectedStatus: http.StatusUnauthorized,
			validateFunc: func(t *testing.T, resp *setup.Response) {
				var res map[string]any
				resp.JSON(&res)
				assert.Equal(t, "unauthorized", res["message"])
			},
		},
		{
			name:           "Error_ActivityUnauthorized",
			endpoint:       "/api/v1/stats/activity?timeframe=month",
			method:         http.MethodGet,
			token:          "", // No token
			expectedStatus: http.StatusUnauthorized,
			validateFunc: func(t *testing.T, resp *setup.Response) {
				var res map[string]any
				resp.JSON(&res)
				assert.Equal(t, "unauthorized", res["message"])
			},
		},
		{
			name:           "Error_InsightsUnauthorized",
			endpoint:       "/api/v1/stats/insights",
			method:         http.MethodGet,
			token:          "", // No token
			expectedStatus: http.StatusUnauthorized,
			validateFunc: func(t *testing.T, resp *setup.Response) {
				var res map[string]any
				resp.JSON(&res)
				assert.Equal(t, "unauthorized", res["message"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headers := map[string]string{}
			if tt.token != "" {
				headers["Authorization"] = "Bearer " + tt.token
			}

			var resp *setup.Response
			if tt.method == http.MethodGet {
				resp = server.Client.GET(tt.endpoint, headers)
			} else {
				t.Fatalf("unsupported method %s", tt.method)
			}

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			if tt.validateFunc != nil {
				tt.validateFunc(t, resp)
			}
		})
	}
}
