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

// Package models defines the Bubble Tea-based
// TUI models for managing and interacting with
// task and project lists.
package models

import (
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/handlebargh/yatto/internal/colors"
)

type (
	// tickMsg is a message type used to trigger time-based updates, such as animations.
	tickMsg time.Time

	// mode defines the state of the TUI, used for contextual behavior (e.g., normal, confirm delete, error).
	mode int

	// doneWaitingMsg signals that the progress bar has finished its post-completion delay.
	doneWaitingMsg struct{}
)

const (
	// modeNormal indicates the default UI mode.
	modeNormal mode = iota

	// modeConfirmDelete indicates the UI is prompting for delete confirmation.
	modeConfirmDelete

	// modeBackendError indicates a backend-related error has occurred and should be displayed.
	modeBackendError
)

var (
	// appStyle defines the base padding for the entire application.
	appStyle = lipgloss.NewStyle().Padding(1, 2)

	// titleStyleProjects styles the title header for the project list.
	titleStyleProjects = lipgloss.NewStyle().
				Foreground(colors.BadgeText()).
				Background(colors.Green()).
				Padding(0, 1)

	// textStyleGreen renders strings using the green foreground color.
	textStyleGreen = lipgloss.NewStyle().
			Foreground(colors.Green()).
			Render

	// textStyleRed renders strings using the red foreground color.
	textStyleRed = lipgloss.NewStyle().
			Foreground(colors.Red()).
			Render
)

// Styles defines a reusable collection of lipgloss styles used in task and project forms.
type Styles struct {
	Base,
	HeaderText,
	Status,
	StatusHeader,
	Title,
	Priority,
	Completed,
	Highlight,
	ErrorHeaderText,
	Help lipgloss.Style
}

// NewStyles returns a new instance of Styles configured using the provided lipgloss.Renderer.
// It defines base padding, bold headers, status boxes, error highlights, and more UI styling presets.
func NewStyles(lg *lipgloss.Renderer) *Styles {
	s := Styles{}
	s.Base = lg.NewStyle().
		Padding(1, 4, 0, 1)
	s.HeaderText = lg.NewStyle().
		Bold(true).
		Padding(0, 1, 0, 2)
	s.Status = lg.NewStyle().
		Border(lipgloss.RoundedBorder()).
		PaddingLeft(1).
		MarginTop(1)
	s.StatusHeader = lg.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#000000", Dark: "#FFFFFF"}).
		Bold(true)
	s.Title = lg.NewStyle().
		Bold(true)
	s.Priority = lg.NewStyle().
		Foreground(lipgloss.Color("#000000")).
		Padding(0, 1)
	s.Completed = lg.NewStyle().
		Foreground(lipgloss.Color("#000000")).
		Padding(0, 1)
	s.Highlight = lg.NewStyle().
		Foreground(lipgloss.Color("212"))
	s.ErrorHeaderText = s.HeaderText.
		Foreground(colors.Red())
	s.Help = lg.NewStyle().
		Foreground(lipgloss.Color("240"))
	return &s
}
