package config

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedactionHook(t *testing.T) {
	entry := logrus.NewEntry(logrus.New()).WithFields(logrus.Fields{
		"api_key": "secret-key",
		"safe":    "value",
		"nested":  logrus.Fields{"access_token": "secret-token", "name": "visible"},
	})

	require.NoError(t, (&RedactionHook{}).Fire(entry))

	assert.Equal(t, "[REDACTED]", entry.Data["api_key"])
	assert.Equal(t, "value", entry.Data["safe"])
	nested := entry.Data["nested"].(logrus.Fields)
	assert.Equal(t, "[REDACTED]", nested["access_token"])
	assert.Equal(t, "visible", nested["name"])
}
