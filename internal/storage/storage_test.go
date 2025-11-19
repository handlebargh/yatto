// Copyright 2025 handlebargh and contributors
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

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
