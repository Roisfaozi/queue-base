package worker

import (
	"bytes"
	"io"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestAsynqLogger(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Success_DebugLog",
			category: "positive",
			run: func(t *testing.T) {
				var buf bytes.Buffer
				logger := logrus.New()
				logger.SetOutput(&buf)
				logger.SetLevel(logrus.DebugLevel)

				asynqLogger := NewAsynqLogger(logger)
				asynqLogger.Debug("test debug message")
				assert.Contains(t, buf.String(), "test debug message")
			},
		},
		{
			name:     "Success_InfoLog",
			category: "positive",
			run: func(t *testing.T) {
				var buf bytes.Buffer
				logger := logrus.New()
				logger.SetOutput(&buf)
				logger.SetLevel(logrus.DebugLevel)

				asynqLogger := NewAsynqLogger(logger)
				asynqLogger.Info("test info message")
				assert.Contains(t, buf.String(), "test info message")
			},
		},
		{
			name:     "Success_WarnLog",
			category: "positive",
			run: func(t *testing.T) {
				var buf bytes.Buffer
				logger := logrus.New()
				logger.SetOutput(&buf)
				logger.SetLevel(logrus.DebugLevel)

				asynqLogger := NewAsynqLogger(logger)
				asynqLogger.Warn("test warn message")
				assert.Contains(t, buf.String(), "test warn message")
			},
		},
		{
			name:     "Success_ErrorLog",
			category: "positive",
			run: func(t *testing.T) {
				var buf bytes.Buffer
				logger := logrus.New()
				logger.SetOutput(&buf)
				logger.SetLevel(logrus.DebugLevel)

				asynqLogger := NewAsynqLogger(logger)
				asynqLogger.Error("test error message")
				assert.Contains(t, buf.String(), "test error message")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestWorkerConfig(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Success_WorkerConfigValues",
			category: "positive",
			run: func(t *testing.T) {
				cfg := WorkerConfig{
					SMTP: SMTPConfig{
						Host:       "smtp.example.com",
						Port:       587,
						Username:   "user@example.com",
						Password:   "secret",
						FromSender: "Test Sender",
						FromEmail:  "test@example.com",
					},
				}

				assert.Equal(t, "smtp.example.com", cfg.SMTP.Host)
				assert.Equal(t, 587, cfg.SMTP.Port)
				assert.Equal(t, "user@example.com", cfg.SMTP.Username)
				assert.Equal(t, "secret", cfg.SMTP.Password)
				assert.Equal(t, "Test Sender", cfg.SMTP.FromSender)
				assert.Equal(t, "test@example.com", cfg.SMTP.FromEmail)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestSMTPConfig_Defaults(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Success_SMTPConfigDefaults",
			category: "positive",
			run: func(t *testing.T) {
				cfg := SMTPConfig{}

				assert.Empty(t, cfg.Host)
				assert.Zero(t, cfg.Port)
				assert.Empty(t, cfg.Username)
				assert.Empty(t, cfg.Password)
				assert.Empty(t, cfg.FromSender)
				assert.Empty(t, cfg.FromEmail)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestNewAsynqLogger_NotNil(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Success_NewAsynqLoggerNotNil",
			category: "positive",
			run: func(t *testing.T) {
				logger := logrus.New()
				logger.SetOutput(io.Discard)

				asynqLogger := NewAsynqLogger(logger)

				assert.NotNil(t, asynqLogger)
				assert.NotNil(t, asynqLogger.logger)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}
