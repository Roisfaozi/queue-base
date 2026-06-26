//go:build e2e
// +build e2e

package realtime

import (
	"encoding/json"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Roisfaozi/queue-base/tests/e2e/setup"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func timestamp() string {
	return strconv.FormatInt(time.Now().UnixNano(), 10)
}

func TestWebSocketE2E_NotificationFlow(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()

	// Initial shared context
	registerPayload := map[string]any{
		"name":     "WS Test User",
		"username": "wstestuser_" + timestamp(),
		"email":    "wstest_" + timestamp() + "@example.com",
		"password": "password123",
	}

	var accessToken string
	var userID string
	var orgID string
	var ticket string
	var conn *websocket.Conn

	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Positive_RegisterUserAndGetToken",
			category: "positive",
			run: func(t *testing.T) {
				w := server.Client.POST("/api/v1/auth/register", registerPayload)
				require.Equal(t, 201, w.StatusCode)

				var registerResp struct {
					Data struct {
						AccessToken string `json:"access_token"`
						User        struct {
							ID string `json:"id"`
						} `json:"user"`
					} `json:"data"`
				}
				err := json.Unmarshal(w.BodyBytes, &registerResp)
				require.NoError(t, err)
				accessToken = registerResp.Data.AccessToken
				userID = registerResp.Data.User.ID
			},
		},
		{
			name:     "Positive_GetUserOrganization",
			category: "positive",
			run: func(t *testing.T) {
				wOrg := server.Client.GET("/api/v1/organizations/me", setup.WithAuth(accessToken))
				require.Equal(t, 200, wOrg.StatusCode)

				var orgResp struct {
					Data struct {
						Organizations []struct {
							ID string `json:"id"`
						} `json:"organizations"`
					} `json:"data"`
				}
				err := json.Unmarshal(wOrg.BodyBytes, &orgResp)
				require.NoError(t, err)
				require.NotEmpty(t, orgResp.Data.Organizations, "User should have at least one organization")
				orgID = orgResp.Data.Organizations[0].ID
			},
		},
		{
			name:     "Positive_RequestWSTicket",
			category: "positive",
			run: func(t *testing.T) {
				wTicket := server.Client.POST("/api/v1/auth/ticket?org_id="+orgID, nil, setup.WithAuth(accessToken))
				require.Equal(t, 200, wTicket.StatusCode)

				var ticketResp struct {
					Data struct {
						Ticket string `json:"ticket"`
					} `json:"data"`
				}
				err := json.Unmarshal(wTicket.BodyBytes, &ticketResp)
				require.NoError(t, err)
				ticket = ticketResp.Data.Ticket
			},
		},
		{
			name:     "Positive_ConnectAndSubscribe",
			category: "positive",
			run: func(t *testing.T) {
				wsURL := strings.Replace(server.BaseURL, "http", "ws", 1) + "/api/v1/ws?ticket=" + ticket
				u, _ := url.Parse(wsURL)
				var err error
				conn, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
				require.NoError(t, err)

				subscribeMsg := map[string]string{
					"type":    "subscribe",
					"channel": "org_" + orgID + "_notifications",
				}
				err = conn.WriteJSON(subscribeMsg)
				require.NoError(t, err)

				var infoMsg struct {
					Type    string `json:"type"`
					Channel string `json:"channel"`
					Data    string `json:"data"`
				}
				err = conn.ReadJSON(&infoMsg)
				require.NoError(t, err)
				assert.Equal(t, "info", infoMsg.Type)
				assert.Equal(t, "org_"+orgID+"_notifications", infoMsg.Channel)
			},
		},
		{
			name:     "Positive_TriggerAndVerifyNotification",
			category: "positive",
			run: func(t *testing.T) {
				time.Sleep(1 * time.Second) // wait for sub

				loginPayload := map[string]any{
					"username": registerPayload["username"],
					"password": registerPayload["password"],
				}

				wLogin := server.Client.POST("/api/v1/auth/login", loginPayload)
				require.Equal(t, 200, wLogin.StatusCode)

				conn.SetReadDeadline(time.Now().Add(10 * time.Second))
				var wsWrapper struct {
					Type    string `json:"type"`
					Channel string `json:"channel"`
					Data    struct {
						Type    string `json:"type"`
						UserID  string `json:"user_id"`
						Message string `json:"message"`
					} `json:"data"`
				}

				_, message, err := conn.ReadMessage()
				require.NoError(t, err)

				err = json.Unmarshal(message, &wsWrapper)
				require.NoError(t, err)

				assert.Equal(t, "message", wsWrapper.Type)
				assert.Equal(t, "user_login", wsWrapper.Data.Type)
				assert.Equal(t, userID, wsWrapper.Data.UserID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
	
	if conn != nil {
		conn.Close()
	}
}

func TestPresenceE2E_IsolationAndEvents(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()

	createUser := func(namePrefix string) (string, string, string) { // returns accessToken, userID, orgID
		registerPayload := map[string]any{
			"name":     namePrefix,
			"username": namePrefix + "_" + timestamp(),
			"email":    namePrefix + "_" + timestamp() + "@example.com",
			"password": "password123",
		}
		w := server.Client.POST("/api/v1/auth/register", registerPayload)
		require.Equal(t, 201, w.StatusCode)

		var resp struct {
			Data struct {
				AccessToken string `json:"access_token"`
				User        struct {
					ID string `json:"id"`
				} `json:"user"`
			} `json:"data"`
		}
		json.Unmarshal(w.BodyBytes, &resp)
		token := resp.Data.AccessToken
		uid := resp.Data.User.ID

		orgPayload := map[string]any{
			"name": namePrefix + " Org " + timestamp(),
			"slug": strings.ToLower(namePrefix) + "-org-" + timestamp(),
		}
		wCreateOrg := server.Client.POST("/api/v1/organizations", orgPayload, setup.WithAuth(token))
		require.Equal(t, 201, wCreateOrg.StatusCode)

		var orgCreateResp struct {
			Data struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		json.Unmarshal(wCreateOrg.BodyBytes, &orgCreateResp)
		orgID := orgCreateResp.Data.ID

		return token, uid, orgID
	}

	connectWS := func(token, orgID string) *websocket.Conn {
		urlPath := "/api/v1/auth/ticket"
		if orgID != "" {
			urlPath += "?org_id=" + orgID
		}
		wTicket := server.Client.POST(urlPath, nil, setup.WithAuth(token))
		require.Equal(t, 200, wTicket.StatusCode)

		var ticketResp struct {
			Data struct {
				Ticket string `json:"ticket"`
			} `json:"data"`
		}
		json.Unmarshal(wTicket.BodyBytes, &ticketResp)

		wsURL := strings.Replace(server.BaseURL, "http", "ws", 1) + "/api/v1/ws?ticket=" + ticketResp.Data.Ticket
		u, _ := url.Parse(wsURL)
		conn, _, _ := websocket.DefaultDialer.Dial(u.String(), nil)
		return conn
	}

	var tokenA, uidA, org1ID string
	var connA *websocket.Conn
	var tokenC, org2ID string
	var connC *websocket.Conn

	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Positive_UserAConnectsAndJoinsPresence",
			category: "positive",
			run: func(t *testing.T) {
				tokenA, uidA, org1ID = createUser("UserA")
				connA = connectWS(tokenA, org1ID)
				
				channelOrg1 := "presence:org:" + org1ID
				connA.WriteJSON(map[string]string{"type": "subscribe", "channel": channelOrg1})
				_, _, _ = connA.ReadMessage() // read sub info

				time.Sleep(500 * time.Millisecond) // wait for async reg
				wPresence := server.Client.GET("/api/v1/organizations/"+org1ID+"/presence", setup.WithAuth(tokenA))
				require.Equal(t, 200, wPresence.StatusCode)
				assert.Contains(t, string(wPresence.BodyBytes), uidA)
			},
		},
		{
			name:     "Positive_SecondConnectionForUserADoesNotTriggerLeave",
			category: "positive",
			run: func(t *testing.T) {
				connA2 := connectWS(tokenA, org1ID)
				defer connA2.Close()
				connA2.WriteJSON(map[string]string{"type": "subscribe", "channel": "presence:org:" + org1ID})

				// Drain A2 join event from A
				connA.SetReadDeadline(time.Now().Add(5 * time.Second))
				_, _, err := connA.ReadMessage()
				require.NoError(t, err)
			},
		},
		{
			name:     "Vulnerability_UserCIsIsolatedFromUserA",
			category: "vulnerability",
			run: func(t *testing.T) {
				tokenC, _, org2ID = createUser("UserC")
				connC = connectWS(tokenC, org2ID)
				connC.WriteJSON(map[string]string{"type": "subscribe", "channel": "presence:org:" + org2ID})

				// A3 connects
				connA3 := connectWS(tokenA, org1ID)
				
				connA3.WriteJSON(map[string]string{"type": "subscribe", "channel": "presence:org:" + org1ID})
				
				connA.SetReadDeadline(time.Now().Add(5 * time.Second))
				foundJoin := false
				for {
					_, msg, err := connA.ReadMessage()
					if err != nil { break }
					if strings.Contains(string(msg), "\"event\":\"join\"") {
						foundJoin = true
						break
					}
				}
				require.True(t, foundJoin, "Did not receive join event")
				
				connA3.Close()
				connA.SetReadDeadline(time.Now().Add(1 * time.Second))
				foundLeave := false
				for {
					_, msg, err := connA.ReadMessage()
					if err != nil { break }
					if strings.Contains(string(msg), "\"event\":\"leave\"") {
						foundLeave = true
						break
					}
				}
				require.False(t, foundLeave, "Received leave event while same user still has active org connections")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}

	if connA != nil { connA.Close() }
	if connC != nil { connC.Close() }
}
