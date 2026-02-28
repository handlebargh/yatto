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

package config

import (
	"path"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestValidateConfig(t *testing.T) {
	tempDir := t.TempDir()
	validStoragePath := tempDir

	baseValidConfig := func() *config {
		return &config{
			storagePath:      validStoragePath,
			vcsBackend:       "git",
			gitDefaultBranch: "main",
			gitRemoteName:    "origin",
			jjDefaultBranch:  "main",
			jjRemoteName:     "origin",
			colorsFormTheme:  "Base16",
			colorValues: map[string]string{
				"colors.red_light": "#ff0000",
			},
		}
	}

	t.Run("valid git config", func(t *testing.T) {
		cfg := baseValidConfig()
		err := cfg.Validate()
		assert.NoError(t, err)
	})

	t.Run("valid jj config", func(t *testing.T) {
		cfg := baseValidConfig()
		cfg.vcsBackend = "jj"
		cfg.jjDefaultBranch = "main"
		cfg.jjRemoteName = "origin"
		err := cfg.Validate()
		assert.NoError(t, err)
	})

	t.Run("invalid storage path - empty", func(t *testing.T) {
		cfg := baseValidConfig()
		cfg.storagePath = ""
		err := cfg.Validate()
		assert.ErrorContains(t, err, "storage path cannot be empty")
	})

	t.Run("invalid storage path - relative", func(t *testing.T) {
		cfg := baseValidConfig()
		cfg.storagePath = "relative/path"
		err := cfg.Validate()
		assert.ErrorContains(t, err, "storage path must be absolute")
	})

	t.Run("unknown vcs backend", func(t *testing.T) {
		cfg := baseValidConfig()
		cfg.vcsBackend = "svn"
		err := cfg.Validate()
		assert.ErrorContains(t, err, "unknown vcs backend: svn")
	})

	t.Run("invalid git branch name", func(t *testing.T) {
		cfg := baseValidConfig()
		cfg.gitDefaultBranch = "invalid branch"
		err := cfg.Validate()
		assert.ErrorContains(t, err, "invalid branch name")
	})

	t.Run("invalid jj remote name", func(t *testing.T) {
		cfg := baseValidConfig()
		cfg.vcsBackend = "jj"
		cfg.jjRemoteName = "invalid remote;"
		err := cfg.Validate()
		assert.ErrorContains(t, err, "invalid remote name")
	})

	t.Run("unknown form theme", func(t *testing.T) {
		cfg := baseValidConfig()
		cfg.colorsFormTheme = "MyAwesomeTheme"
		err := cfg.Validate()
		assert.ErrorContains(t, err, "unknown colors.form.theme")
	})

	t.Run("invalid color value", func(t *testing.T) {
		cfg := baseValidConfig()
		cfg.colorValues["colors.red_light"] = "not-a-color"
		err := cfg.Validate()
		assert.ErrorContains(t, err, "invalid color value")
	})
}

func TestInitConfig(t *testing.T) {
	v := viper.New()
	homeDir := "/fake/home"
	configPath := ""

	InitConfig(v, homeDir, &configPath)

	assert.Equal(t, path.Join(homeDir, ".yatto"), v.GetString("storage.path"))
	assert.Equal(t, "git", v.GetString("vcs.backend"))
	assert.Equal(t, "main", v.GetString("git.default_branch"))
	assert.Equal(t, "Base16", v.GetString("colors.form.theme"))
	assert.Equal(t, path.Join(homeDir, ".config", "yatto", "config.toml"), configPath)

	// Test with explicit config path
	v = viper.New()
	explicitPath := "/my/config.toml"
	InitConfig(v, homeDir, &explicitPath)
	assert.Equal(t, explicitPath, v.ConfigFileUsed())
}
