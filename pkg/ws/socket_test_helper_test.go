package ws_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newPermissiveWSServer(handler http.Handler) (server *httptest.Server, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			msg := ""
			switch value := recovered.(type) {
			case string:
				msg = value
			case error:
				msg = value.Error()
			}
			if strings.Contains(msg, "operation not permitted") {
				err = http.ErrServerClosed
				return
			}
			panic(recovered)
		}
	}()
	return httptest.NewServer(handler), nil
}

func skipIfSocketBlocked(t *testing.T, err error) {
	t.Helper()
	if err == http.ErrServerClosed || (err != nil && strings.Contains(err.Error(), "operation not permitted")) {
		t.Skip("socket listeners not permitted in this environment")
	}
}
