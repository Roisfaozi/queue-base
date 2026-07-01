package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Success",
			category: "positive",
			run: func(t *testing.T) {
				password := "mysecretpassword"
				hash, err := HashPassword(password)

				assert.NoError(t, err, "HashPassword should not return an error")
				assert.NotEmpty(t, hash, "Hash should not be empty")
				assert.NotEqual(t, password, hash, "Hash should be different from the original password")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestCheckPasswordHash_Success(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Success",
			category: "positive",
			run: func(t *testing.T) {
				password := "mysecretpassword"
				hash, err := HashPassword(password)
				assert.NoError(t, err)

				match := CheckPasswordHash(password, hash)
				assert.True(t, match, "Password and hash should match")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestCheckPasswordHash_Failure(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Failure",
			category: "negative",
			run: func(t *testing.T) {
				password := "mysecretpassword"
				wrongPassword := "wrongpassword"
				hash, err := HashPassword(password)
				assert.NoError(t, err)

				match := CheckPasswordHash(wrongPassword, hash)
				assert.False(t, match, "Wrong password should not match the hash")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestCheckPasswordHash_InvalidHash(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "InvalidHash",
			category: "negative",
			run: func(t *testing.T) {
				password := "mysecretpassword"
				invalidHash := "this-is-not-a-valid-bcrypt-hash"

				match := CheckPasswordHash(password, invalidHash)
				assert.False(t, match, "Password should not match an invalid hash")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}
