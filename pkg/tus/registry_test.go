package tus

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// MockHook for testing
type MockHook struct {
	called bool
}

func (m *MockHook) HandleUpload(ctx context.Context, event UploadEvent) error {
	m.called = true
	return nil
}

func TestRegistry_RegisterAndGet(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Register And Get",
			category: "positive",
			run: func(t *testing.T) {
				registry := NewRegistry()
				hook := &MockHook{}
				uploadType := "avatar"

				registry.Register(uploadType, hook)
				retrieved := registry.Get(uploadType)

				assert.NotNil(t, retrieved)
				assert.Equal(t, hook, retrieved)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestRegistry_GetNonExistent(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Get Non Existent",
			category: "negative",
			run: func(t *testing.T) {
				registry := NewRegistry()
				retrieved := registry.Get("non-existent")

				assert.Nil(t, retrieved)
				assert.False(t, registry.Has("non-existent"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestRegistry_HasRegisteredType(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Has Registered Type",
			category: "positive",
			run: func(t *testing.T) {
				registry := NewRegistry()
				hook := &MockHook{}

				registry.Register("avatar", hook)

				assert.True(t, registry.Has("avatar"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestRegistry_OverwriteHook(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Overwrite Hook",
			category: "edge",
			run: func(t *testing.T) {
				registry := NewRegistry()
				hook1 := &MockHook{}
				hook2 := &MockHook{}
				uploadType := "avatar"

				registry.Register(uploadType, hook1)
				registry.Register(uploadType, hook2) // Overwrite

				retrieved := registry.Get(uploadType)
				assert.Equal(t, hook2, retrieved)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestRegistry_EmptyType(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Empty Type",
			category: "edge",
			run: func(t *testing.T) {
				registry := NewRegistry()
				hook := &MockHook{}

				registry.Register("", hook)
				assert.Equal(t, hook, registry.Get(""))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestRegistry_TypeInjection(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Type Injection",
			category: "security",
			run: func(t *testing.T) {
				registry := NewRegistry()
				hook := &MockHook{}
				injectionKey := "../../etc/passwd%20' OR 1=1"

				registry.Register(injectionKey, hook)
				assert.Equal(t, hook, registry.Get(injectionKey))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}
