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

// Package fetchmodel provides the spinner animation that runs during the
// pull command at startup.
package fetchmodel

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/handlebargh/yatto/internal/vcs"
	"github.com/spf13/viper"
)

// FetchModel defines the model used for displaying a spinner while syncing with a remote Git repository.
type FetchModel struct {
	Config    *viper.Viper
	Spinner   spinner.Model
	CmdOutput string
	Err       error
	Width     int
	Height    int
}

// NewFetchModel initializes and returns a new FetchModel instance,
func NewFetchModel(v *viper.Viper) FetchModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = s.Style.
		Foreground(lipgloss.AdaptiveColor{Light: "#FFB733", Dark: "#FFA336"}).
		Bold(true)

	m := FetchModel{
		Config:  v,
		Spinner: s,
	}

	return m
}

// Init initializes the spinner model and starts the init command.
func (m FetchModel) Init() tea.Cmd {
	return tea.Batch(
		m.Spinner.Tick,
		vcs.InitCmd(m.Config),
	)
}

// Update handles messages received during the spinner's lifecycle,
// such as window resize, spinner ticks, Git pull results, and user input.
func (m FetchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.Spinner, cmd = m.Spinner.Update(msg)
		return m, cmd

	case vcs.InitDoneMsg:
		return m, vcs.PullCmd(m.Config)

	case vcs.InitErrorMsg:
		m.CmdOutput = msg.CmdOutput
		m.Err = msg.Err
		return m, nil

	case vcs.PullDoneMsg:
		return m, tea.Quit

	case vcs.PullErrorMsg:
		m.Err = msg.Err
		return m, nil

	case vcs.PullNoInitMsg:
		m.Err = vcs.ErrorNoInit
		return m, nil

	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Interrupt
		}

		switch msg.String() {
		case "esc", "q":
			return m, tea.Interrupt
		}
	}

	return m, nil
}

// View renders the spinner UI, displaying a loading animation or an error message,
// centered in the terminal window.
func (m FetchModel) View() string {
	var content string
	if m.Err != nil {
		content = m.CmdOutput
	} else {
		content = fmt.Sprintf("%s Fetching data from remote…", m.Spinner.View())
	}

	// Center horizontally and vertically
	return lipgloss.Place(
		m.Width,
		m.Height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}
