package fetchmodel

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/handlebargh/yatto/internal/vcs"
)

// FetchModel defines the model used for displaying a spinner while syncing with a remote Git repository.
type FetchModel struct {
	Spinner   spinner.Model
	CmdOutput string
	Err       error
	Width     int
	Height    int
}

// Init initializes the spinner model and starts the init command.
func (m FetchModel) Init() tea.Cmd {
	return tea.Batch(
		m.Spinner.Tick,
		vcs.InitCmd(),
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
		return m, vcs.PullCmd()

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
