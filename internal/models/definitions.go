package models

import "github.com/charmbracelet/lipgloss"

const maxWidth = 80

var (
	red     = lipgloss.AdaptiveColor{Light: "#FE5F86", Dark: "#FE5F86"}
	blue    = lipgloss.AdaptiveColor{Light: "#4DA6FF", Dark: "#4DA6FF"}
	indigo  = lipgloss.AdaptiveColor{Light: "#5A56E0", Dark: "#7571F9"}
	green   = lipgloss.AdaptiveColor{Light: "#02BA84", Dark: "#02BF87"}
	orange  = lipgloss.AdaptiveColor{Light: "#FFB733", Dark: "#FFA336"}
	neutral = lipgloss.AdaptiveColor{Light: "#000000", Dark: "#FFFFFF"}
)

var (
	appStyle = lipgloss.NewStyle().Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000")).
			Background(green).
			Padding(0, 1)

	detailBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.AdaptiveColor{Light: "#333333", Dark: "#CCCCCC"}).
			Padding(1, 2).
			Margin(1, 1).
			Width(50)

	promptBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(1, 2).
			Margin(1, 1).
			BorderForeground(lipgloss.Color("9")).
			Align(lipgloss.Center).
			Width(50)

	statusMessageStyleGreen = lipgloss.NewStyle().
				Foreground(green).
				Render

	statusMessageStyleRed = lipgloss.NewStyle().
				Foreground(red).
				Render
)

type mode int

const (
	modeNormal mode = iota
	modeConfirmDelete
	modeGitError
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
		Foreground(indigo).
		Bold(true).
		Padding(0, 1, 0, 2)
	s.Status = lg.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(indigo).
		PaddingLeft(1).
		MarginTop(1)
	s.StatusHeader = lg.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#000000", Dark: "#FFFFFF"}).
		Bold(true)
	s.Title = lg.NewStyle().
		Foreground(lipgloss.Color("#000000")).
		Background(green).
		Padding(0, 1).
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

func min(x, y int) int {
	if x > y {
		return y
	}
	return x
}
