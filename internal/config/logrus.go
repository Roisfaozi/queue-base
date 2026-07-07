package config

import (
	"fmt"
	"path"
	"runtime"
	"strings"

	"github.com/Roisfaozi/queue-base/pkg/constants"
	"github.com/sirupsen/logrus"
)

// TraceContextHook attaches RequestID from context to the log entry
type TraceContextHook struct{}

type RedactionHook struct{}

func (h *TraceContextHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *TraceContextHook) Fire(entry *logrus.Entry) error {
	if entry.Context != nil {
		if reqID, ok := entry.Context.Value(constants.RequestIDKey).(string); ok {
			entry.Data["request_id"] = reqID
		}
		// Also trace UserID if available (e.g. from authenticated context)
		if userID, ok := entry.Context.Value(constants.UserIDKey).(string); ok {
			entry.Data["user_id"] = userID
		}
	}
	return nil
}

func (h *RedactionHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *RedactionHook) Fire(entry *logrus.Entry) error {
	for key, value := range entry.Data {
		entry.Data[key] = redactLogValue(key, value)
	}
	return nil
}

func redactLogValue(key string, value any) any {
	if isSensitiveLogKey(key) {
		return "[REDACTED]"
	}
	switch v := value.(type) {
	case logrus.Fields:
		return redactLogFields(v)
	case map[string]any:
		return redactLogFields(logrus.Fields(v))
	case map[string]string:
		fields := logrus.Fields{}
		for k, val := range v {
			fields[k] = val
		}
		return redactLogFields(fields)
	default:
		return value
	}
}

func redactLogFields(fields logrus.Fields) logrus.Fields {
	redacted := logrus.Fields{}
	for key, value := range fields {
		redacted[key] = redactLogValue(key, value)
	}
	return redacted
}

func isSensitiveLogKey(key string) bool {
	normalized := strings.ToLower(strings.ReplaceAll(key, "-", "_"))
	for _, token := range []string{"password", "secret", "token", "api_key", "apikey", "authorization", "credential", "hash"} {
		if strings.Contains(normalized, token) {
			return true
		}
	}
	return false
}

func NewLogrus(config *AppConfig) *logrus.Logger {
	logger := logrus.New()

	// Add Trace Hook
	logger.AddHook(&TraceContextHook{})
	logger.AddHook(&RedactionHook{})

	level, err := logrus.ParseLevel(config.Log.Level)
	if err != nil {
		logger.SetLevel(logrus.InfoLevel)
		logger.Warnf("Invalid log level '%s'. Defaulting to 'info'.", config.Log.Level)
	} else {
		logger.SetLevel(level)
	}

	logger.SetReportCaller(true)

	if config.Server.AppEnv == "development" {
		logger.SetFormatter(&logrus.TextFormatter{
			ForceColors:     true,
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05.000",
			CallerPrettyfier: func(f *runtime.Frame) (string, string) {
				filename := path.Base(f.File)
				return fmt.Sprintf("%s()", f.Function), fmt.Sprintf("%s:%d", filename, f.Line)
			},
		})
	} else {
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05.000",
			CallerPrettyfier: func(f *runtime.Frame) (string, string) {
				filename := path.Base(f.File)
				return fmt.Sprintf("%s()", f.Function), fmt.Sprintf("%s:%d", filename, f.Line)
			},
		})
	}

	return logger
}
