package config

import (
	"bytes"
	"io"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

type answerReader struct {
	answers []string
	idx     int
}

func (r *answerReader) Read(p []byte) (n int, err error) {
	if r.idx >= len(r.answers) {
		return 0, io.EOF
	}
	s := r.answers[r.idx] + "\n"
	r.idx++
	copy(p, s)
	return len(s), nil
}

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
	viper.Reset()
	homeDir := "/fake/home"
	configPath := ""

	InitConfig(homeDir, &configPath)

	assert.Equal(t, filepath.Join(homeDir, ".yatto"), viper.GetString("storage.path"))
	assert.Equal(t, "git", viper.GetString("vcs.backend"))
	assert.Equal(t, "main", viper.GetString("git.default_branch"))
	assert.Equal(t, "Base16", viper.GetString("colors.form.theme"))
	assert.Equal(t, filepath.Join(homeDir, ".config", "yatto", "config.toml"), configPath)

	// Test with explicit config path
	viper.Reset()
	explicitPath := "/my/config.toml"
	InitConfig(homeDir, &explicitPath)
	assert.Equal(t, explicitPath, viper.ConfigFileUsed())
}

func TestCreateConfigFile(t *testing.T) {
	t.Run("aborts when user declines", func(t *testing.T) {
		viper.Reset()
		tempDir := t.TempDir()
		settings := Settings{
			ConfigPath: filepath.Join(tempDir, "config.toml"),
			Home:       tempDir,
			Input:      bytes.NewBufferString("n\n"),
			Output:     &bytes.Buffer{},
		}

		err := CreateConfigFile(settings)
		assert.ErrorIs(t, err, ErrUserAborted)
	})

	t.Run("creates config with git and remote", func(t *testing.T) {
		viper.Reset()
		tempDir := t.TempDir()
		answers := []string{"y", "git", "git@github.com/user/repo.git", "n"}

		settings := Settings{
			ConfigPath: "",
			Home:       tempDir,
			Input:      &answerReader{answers: answers},
			Output:     &bytes.Buffer{},
		}

		var empty string
		InitConfig(settings.Home, &empty)

		err := CreateConfigFile(settings)
		assert.NoError(t, err)

		// Re-read the config to check values
		err = LoadAndValidateConfig()
		assert.NoError(t, err)

		assert.Equal(t, "git", viper.GetString("vcs.backend"))
		assert.True(t, viper.GetBool("git.remote.enable"))
		assert.Equal(t, "git@github.com/user/repo.git", viper.GetString("git.remote.url"))
	})
}
