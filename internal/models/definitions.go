// Copyright 2025 handlebargh
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

package models

import (
	"time"

	"github.com/charmbracelet/lipgloss"
)

type (
	tickMsg        time.Time
	mode           int
	doneWaitingMsg struct{}
)

const (
	modeNormal mode = iota
	modeConfirmDelete
	modeGitError
)

var (
	red    = lipgloss.AdaptiveColor{Light: "#FE5F86", Dark: "#FE5F86"}
	indigo = lipgloss.AdaptiveColor{Light: "#5A56E0", Dark: "#7571F9"}
	green  = lipgloss.AdaptiveColor{Light: "#02BA84", Dark: "#02BF87"}
	orange = lipgloss.AdaptiveColor{Light: "#FFB733", Dark: "#FFA336"}
	blue   = lipgloss.AdaptiveColor{Light: "#1e90ff", Dark: "#1e90ff"}
	black  = lipgloss.Color("#000000")
)

var (
	appStyle = lipgloss.NewStyle().Padding(1, 2)

	titleStyleProjects = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#000000")).
				Background(green).
				Padding(0, 1)

	textStyleGreen = lipgloss.NewStyle().
			Foreground(green).
			Render

	textStyleRed = lipgloss.NewStyle().
			Foreground(red).
			Render
)

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
		Foreground(red)
	s.Help = lg.NewStyle().
		Foreground(lipgloss.Color("240"))
	return &s
}
