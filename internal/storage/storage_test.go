package storage

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestFileExists(t *testing.T) {
	tempDir := t.TempDir()
	viper.Set("storage.path", tempDir)

	t.Run("returns true when file exists", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "exists.txt")
		_, err := os.Create(filePath)
		assert.NoError(t, err)

		assert.True(t, FileExists("exists.txt"))
	})

	t.Run("returns false when file does not exist", func(t *testing.T) {
		assert.False(t, FileExists("nonexistent.txt"))
	})
}

func TestCreateStorageDir(t *testing.T) {
	t.Run("does nothing if directory already exists", func(t *testing.T) {
		tempDir := t.TempDir()
		settings := Settings{
			Path:   tempDir,
			Input:  &bytes.Buffer{},
			Output: &bytes.Buffer{},
			Exit:   func(int) {},
		}

		err := CreateStorageDir(settings)
		assert.NoError(t, err)
	})

	t.Run("creates directory when user confirms", func(t *testing.T) {
		tempDir := t.TempDir()
		storagePath := filepath.Join(tempDir, "storage")

		settings := Settings{
			Path:   storagePath,
			Input:  bytes.NewBufferString("y\n"),
			Output: &bytes.Buffer{},
			Exit:   func(int) {},
		}

		err := CreateStorageDir(settings)
		assert.NoError(t, err)

		_, err = os.Stat(storagePath)
		assert.NoError(t, err, "storage directory should be created")
	})

	t.Run("returns ErrUserAborted when user declines", func(t *testing.T) {
		tempDir := t.TempDir()
		storagePath := filepath.Join(tempDir, "storage")

		settings := Settings{
			Path:   storagePath,
			Input:  bytes.NewBufferString("n\n"),
			Output: &bytes.Buffer{},
			Exit:   func(int) {},
		}

		err := CreateStorageDir(settings)
		assert.ErrorIs(t, err, ErrUserAborted)

		_, err = os.Stat(storagePath)
		assert.True(t, os.IsNotExist(err), "storage directory should not be created")
	})

	t.Run("returns ErrUserAborted on unexpected input", func(t *testing.T) {
		tempDir := t.TempDir()
		storagePath := filepath.Join(tempDir, "storage")

		settings := Settings{
			Path:   storagePath,
			Input:  bytes.NewBufferString("maybe\n"),
			Output: &bytes.Buffer{},
			Exit:   func(int) {},
		}

		err := CreateStorageDir(settings)
		assert.ErrorIs(t, err, ErrUserAborted)
	})
}
