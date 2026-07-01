package local_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Roisfaozi/queue-base/pkg/storage/local"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLocalStorage(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Create Success",
			category: "positive",
			run: func(t *testing.T) {
				ls, err := local.NewLocalStorage(tempDir, "http://localhost")
				assert.NoError(t, err)
				assert.NotNil(t, ls)
			},
		},
		{
			name:     "Create Default Path",
			category: "positive",
			run: func(t *testing.T) {
				// Caution: this creates ./uploads in the current directory
				// We should clean it up or skip if we want to avoid side effects in source tree
				// But for coverage we might need to test it.
				// Alternatively, we skip this if we can't easily mock os.MkdirAll failure here.
				// Let's pass empty string which defaults to ./uploads
				ls, err := local.NewLocalStorage("", "http://localhost")
				assert.NoError(t, err)
				assert.Equal(t, "./uploads", ls.RootPath)
				_ = os.RemoveAll("./uploads")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestUploadFile(t *testing.T) {
	tempDir := t.TempDir()
	baseURL := "http://localhost/files"
	ls, err := local.NewLocalStorage(tempDir, baseURL)
	require.NoError(t, err)

	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Success",
			category: "positive",
			run: func(t *testing.T) {
				content := "test content"
				reader := strings.NewReader(content)
				filename := "test.txt"

				url, err := ls.UploadFile(context.Background(), reader, filename, "text/plain")
				assert.NoError(t, err)
				assert.Equal(t, baseURL+"/"+filename, url)

				// Verify file exists
				data, err := os.ReadFile(filepath.Join(tempDir, filename))
				assert.NoError(t, err)
				assert.Equal(t, content, string(data))
			},
		},
		{
			name:     "Path Traversal",
			category: "vulnerability",
			run: func(t *testing.T) {
				content := "hacker content"
				reader := strings.NewReader(content)
				// Should be cleaned to just "passwd" inside root path
				filename := "../../../etc/passwd"

				url, err := ls.UploadFile(context.Background(), reader, filename, "text/plain")
				assert.NoError(t, err)
				assert.Equal(t, baseURL+"/passwd", url) // filepath.Clean removes ..

				// Verify it was written to tempDir/passwd
				_, err = os.Stat(filepath.Join(tempDir, "passwd"))
				assert.NoError(t, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestDeleteFile(t *testing.T) {
	tempDir := t.TempDir()
	baseURL := "http://localhost/files"
	ls, err := local.NewLocalStorage(tempDir, baseURL)
	require.NoError(t, err)

	filename := "delete_me.txt"
	fullPath := filepath.Join(tempDir, filename)
	err = os.WriteFile(fullPath, []byte("content"), 0644)
	require.NoError(t, err)

	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Success",
			category: "positive",
			run: func(t *testing.T) {
				err := ls.DeleteFile(context.Background(), filename)
				assert.NoError(t, err)

				_, err = os.Stat(fullPath)
				assert.True(t, os.IsNotExist(err))
			},
		},
		{
			name:     "Not Exist",
			category: "negative",
			run: func(t *testing.T) {
				err := ls.DeleteFile(context.Background(), "nonexistent.txt")
				assert.NoError(t, err) // Should return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestGetFileUrl(t *testing.T) {
	ls := &local.LocalStorage{BaseURL: "http://localhost"}

	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Normal Path",
			category: "positive",
			run: func(t *testing.T) {
				url, err := ls.GetFileUrl(context.Background(), "test.png")
				assert.NoError(t, err)
				assert.Equal(t, "http://localhost/test.png", url)
			},
		},
		{
			name:     "Path Traversal Cleaned",
			category: "vulnerability",
			run: func(t *testing.T) {
				url, err := ls.GetFileUrl(context.Background(), "../test.png")
				assert.NoError(t, err)
				assert.Equal(t, "http://localhost/test.png", url)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestUploadFile_CreateError(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Invalid Nested Path",
			category: "negative",
			run: func(t *testing.T) {
				// Use an invalid/non-existent nested path that cannot be created
				// This works cross-platform (Windows and Unix)
				invalidPath := filepath.Join(t.TempDir(), "nonexistent", "deep", "path")

				// Don't create the parent directories - the upload should fail
				ls := &local.LocalStorage{RootPath: invalidPath, BaseURL: "http://localhost"}

				reader := strings.NewReader("content")
				_, err := ls.UploadFile(context.Background(), reader, "fail.txt", "text/plain")
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "failed to create destination file")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}
