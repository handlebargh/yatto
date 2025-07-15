package models

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
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
	red     = lipgloss.AdaptiveColor{Light: "#FE5F86", Dark: "#FE5F86"}
	indigo  = lipgloss.AdaptiveColor{Light: "#5A56E0", Dark: "#7571F9"}
	green   = lipgloss.AdaptiveColor{Light: "#02BA84", Dark: "#02BF87"}
	orange  = lipgloss.AdaptiveColor{Light: "#FFB733", Dark: "#FFA336"}
	neutral = lipgloss.AdaptiveColor{Light: "#333333", Dark: "#DDDDDD"}
)

var (
	appStyle = lipgloss.NewStyle().Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000")).
			Background(green).
			Padding(0, 1)

	statusMessageStyleGreen = lipgloss.NewStyle().
				Foreground(green).
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

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second*1, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
