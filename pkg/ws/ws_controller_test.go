package ws

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// NoOpLogWriter suppresses log output for tests
type NoOpLogWriter struct{}

func (w *NoOpLogWriter) Write([]byte) (int, error) {
	return 0, nil
}

// TestWebSocketOrigin tests origin checking
func TestWebSocketOrigin(t *testing.T) {
	log := logrus.New()
	log.SetOutput(&NoOpLogWriter{})

	tests := []struct {
		name           string
		allowedOrigins []string
		requestOrigin  string
		shouldConnect  bool
	}{
		{
			name:           "Allowed Origin",
			allowedOrigins: []string{"http://example.com"},
			requestOrigin:  "http://example.com",
			shouldConnect:  true,
		},
		{
			name:           "Disallowed Origin",
			allowedOrigins: []string{"http://example.com"},
			requestOrigin:  "http://malicious.com",
			shouldConnect:  false,
		},
		{
			name:           "Wildcard Origin",
			allowedOrigins: []string{"*"},
			requestOrigin:  "http://anywhere.com",
			shouldConnect:  true,
		},
		{
			name:           "Empty Allowed Origins",
			allowedOrigins: []string{},
			requestOrigin:  "http://anywhere.com",
			shouldConnect:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh mock for each subtest
			manager := new(MockManager)
			manager.On("RegisterClient", mock.Anything).Return().Maybe()
			manager.On("UnregisterClient", mock.Anything).Return().Maybe()

			ctrl := NewWebSocketController(log, manager, tt.allowedOrigins, nil, nil)

			// Start a test server
			r := http.NewServeMux()
			r.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
				gin.SetMode(gin.ReleaseMode)
				engine := gin.New()
				engine.GET("/ws", ctrl.HandleWebSocket)
				engine.ServeHTTP(w, r)
			})

			server, err := newPermissiveWSServer(t, http.HandlerFunc(r.ServeHTTP))
			if err != nil {
				return
			}
			defer server.Close()

			// Build WebSocket URL
			wsURL := "ws" + server.URL[4:] + "/ws"

			header := http.Header{}
			if tt.requestOrigin != "" {
				header.Set("Origin", tt.requestOrigin)
			}

			conn, resp, err := websocket.DefaultDialer.Dial(wsURL, header)

			if tt.shouldConnect {
				assert.NoError(t, err, "Should connect")
				if err == nil {
					assert.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode)
					// Close connection to prevent dangling goroutines
					_ = conn.Close()
				}
			} else {
				assert.Error(t, err, "Should fail to connect")
				if resp != nil {
					assert.Equal(t, http.StatusForbidden, resp.StatusCode)
				}
			}
		})
	}
}
