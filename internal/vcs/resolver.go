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

// Package vcs provides internal helpers for managing vcs operations
// such as initialization, committing, and pulling in the configured
// storage directory.
package vcs

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/viper"
)

// InitCmd returns the backend specific init command according
// to configuration.
func InitCmd(v *viper.Viper) tea.Cmd {
	switch v.GetString("vcs.backend") {
	case "git":
		return gitInitCmd(v)
	case "jj":
		return jjInitCmd(v)
	default:
		return nil
	}
}

// CommitCmd returns the backend specific commit command according
// to configuration.
func CommitCmd(v *viper.Viper, message string, files ...string) tea.Cmd {
	switch v.GetString("vcs.backend") {
	case "git":
		return gitCommitCmd(v, message, files...)
	case "jj":
		return jjCommitCmd(v, message)
	default:
		return nil
	}
}

// PullCmd returns the backend specific pull/fetch command according
// to configuration.
func PullCmd(v *viper.Viper) tea.Cmd {
	switch v.GetString("vcs.backend") {
	case "git":
		return gitPullCmd(v)
	case "jj":
		return jjPullCmd(v)
	default:
		return nil
	}
}

// User returns the backend specific userEmail command according
// to configuration.
func User(v *viper.Viper) (string, error) {
	switch v.GetString("vcs.backend") {
	case "git":
		return gitUser(v)
	case "jj":
		return jjUser(v)
	default:
		return "", nil
	}
}

// AllContributors returns the backend specific
// contributors command according to configuration.
func AllContributors(v *viper.Viper) ([]string, error) {
	switch v.GetString("vcs.backend") {
	case "git":
		return gitContributors(v)
	case "jj":
		return jjContributors(v)
	default:
		return nil, nil
	}
}
