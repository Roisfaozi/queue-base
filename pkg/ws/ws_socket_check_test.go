package ws_test

import (
	"net"
	"os"
	"testing"
)

// canListenTCP checks whether the environment allows binding a TCP socket.
// Many sandbox/CI environments block socket creation.
func canListenTCP(tb testing.TB) bool {
	tb.Helper()
	if os.Getenv("NO_SOCKET_TESTS") != "" {
		return false
	}
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err == nil {
		_ = l.Close()
		return true
	}
	return false
}
