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
	"fmt"
	"io"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
	"github.com/handlebargh/yatto/internal/git"
	"github.com/handlebargh/yatto/internal/helpers"
	"github.com/handlebargh/yatto/internal/items"
)

// projectListKeyMap defines the key bindings
// used in the project list UI model.
type projectListKeyMap struct {
	quit           key.Binding
	toggleHelpMenu key.Binding
	addProject     key.Binding
	editProject    key.Binding
	chooseProject  key.Binding
	deleteProject  key.Binding
	toggleHelpMenu key.Binding
	addProject     key.Binding
	editProject    key.Binding
	chooseProject  key.Binding
	deleteProject  key.Binding
}

// newProjectListKeyMap returns a new set of key
// bindings for project list operations.
func newProjectListKeyMap() *projectListKeyMap {
	return &projectListKeyMap{
		quit: key.NewBinding(
			key.WithKeys("q", "esc"),
			key.WithHelp("q/esc", "quit"),
		),
		deleteProject: key.NewBinding(
			key.WithKeys("D"),
			key.WithHelp("D", "delete project"),
		),
		chooseProject: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "choose project"),
		),
		addProject: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "add project"),
		),
		editProject: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit project"),
		),
		toggleHelpMenu: key.NewBinding(
			key.WithKeys("H"),
			key.WithHelp("H", "toggle help"),
		),
		chooseProjectVim: key.NewBinding(
			key.WithKeys("l"),
			key.WithHelp("l", "choose project"),
		),
	}
}

// customProjectDelegate implements a custom
// renderer for items in the project list.
type customProjectDelegate struct {
	list.DefaultDelegate
}

// Render renders a custom project item in the list,
// including its task summary and status indicators.
func (d customProjectDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	projectItem, ok := item.(*items.Project)
	if !ok {
		_, err := fmt.Fprint(w, "Invalid item\n")
		if err != nil {
			panic(err)
		}

		return
	}

	color := helpers.GetColorCode(projectItem.Color())

	// Base styles.
	listItemStyle := lipgloss.NewStyle().
		Foreground(color).
		Padding(0, 1).
		Width(60)

	listItemInfoStyle := lipgloss.NewStyle().
		Align(lipgloss.Right)

	if index == m.GlobalIndex() {
		listItemStyle = listItemStyle.
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(color).
			MarginLeft(0)
	} else {
		listItemStyle = listItemStyle.MarginLeft(1)
	}

	left := listItemStyle.Render(projectItem.Title() + "\n" +
		projectItem.Description())

	numTasks, numCompletedTasks, numDueTasks, err := projectItem.NumOfTasks()
	if err != nil {
		m.NewStatusMessage(
			textStyleRed(fmt.Sprintf("Error gathering task info for project %s", projectItem.Title())),
		)
	}

	var taskDueMessage string
	if numDueTasks > 0 {
		if numDueTasks == 1 {
			taskDueMessage = textStyleRed("1 task due today")
		} else {
			taskDueMessage = textStyleRed(fmt.Sprintf("%d tasks due today", numDueTasks))
		}
	}

	taskTotalCompleteMessage := fmt.Sprintf("%d/%d tasks completed", numCompletedTasks, numTasks)
	if numCompletedTasks == numTasks {
		taskTotalCompleteMessage = textStyleGreen(taskTotalCompleteMessage)
	}

	right := listItemInfoStyle.Render(
		fmt.Sprintf("%s\n%s", taskTotalCompleteMessage, taskDueMessage),
	)

	row := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Render(left),
		right,
	)

	_, err = fmt.Fprint(w, row)
	if err != nil {
		panic(err)
	}
}

// projectListModel defines the TUI model used to
// manage and interact with projects.
type projectListModel struct {
	list             list.Model
	selected         bool
	keys             *projectListKeyMap
	mode             mode
	err              error
	progress         progress.Model
	progressDone     bool
	waitingAfterDone bool
	status           string
	width, height    int

	renderer *glamour.TermRenderer
}

// InitialProjectListModel returns an initialized projectListModel
// with all necessary state and UI settings.
func InitialProjectListModel() projectListModel {
	listKeys := newProjectListKeyMap()

	projects := helpers.ReadProjectsFromFS()
	items := []list.Item{}

	for _, project := range projects {
		items = append(items, &project)
	}

	itemList := list.New(items, customProjectDelegate{DefaultDelegate: list.NewDefaultDelegate()}, 0, 0)
	itemList.SetShowPagination(true)
	itemList.SetShowTitle(true)
	itemList.SetShowStatusBar(true)
	itemList.SetStatusBarItemName("project", "projects")
	itemList.Title = "Projects"
	itemList.Styles.Title = titleStyleProjects
	// Disable the quit keybindings, so we can implement our own.
	itemList.DisableQuitKeybindings()
	itemList.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			listKeys.quit,
		}
	}
	itemList.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			listKeys.toggleHelpMenu,
			listKeys.addProject,
			listKeys.chooseProject,
			listKeys.deleteProject,
			listKeys.editProject,
			listKeys.chooseProjectVim,
		}
	}

	renderer, err := glamour.NewTermRenderer(glamour.WithAutoStyle())
	if err != nil {
		panic(err)
	}

	return projectListModel{
		list:     itemList,
		keys:     listKeys,
		renderer: renderer,
		progress: progress.New(progress.WithGradient("#FFA336", "#02BF87")),
	}
}

// Init initializes the Bubble Tea program
// for the project list model.
func (m projectListModel) Init() tea.Cmd {
	return tea.Batch(
		tickCmd(),
		git.InitCmd(),
	)
}

// Update handles incoming messages and updates
// the project list model state accordingly.
func (m projectListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tickMsg:
		if m.progress.Percent() >= 1.0 && !m.waitingAfterDone {
			m.progressDone = true
			m.waitingAfterDone = true

			// Return a timer command to keep displaying 100% progress
			// for half a second.
			return m, tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
				return doneWaitingMsg{}
			})
		}

		return m, tickCmd()

	// FrameMsg is sent when the progress bar wants to animate itself
	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd

	case doneWaitingMsg:
		m.progressDone, m.waitingAfterDone = false, false
		// Reset the progress bar.
		return m, m.progress.SetPercent(0.0)

	case git.GitInitDoneMsg:
		return m, nil

	case git.GitInitErrorMsg:
		m.mode = 2
		m.err = msg.Err
		return m, nil

	case git.GitCommitDoneMsg:
		m.status = "🗘  Changes committed"
		m.progressDone = true
		return m, m.progress.SetPercent(1.0)

	case git.GitCommitErrorMsg:
		m.mode = 2
		m.err = msg.Err
		return m, m.progress.SetPercent(0.0)

	case items.WriteProjectJSONDoneMsg:
		switch msg.Kind {
		case "create":
			m.list.InsertItem(0, &msg.Project)
			m.status = "🗸  Project created"
			return m, m.progress.SetPercent(0.5)

		case "update":
			m.status = "🗸  Project updated"
			return m, m.progress.SetPercent(0.5)

		default:
			return m, nil
		}

	case items.WriteProjectJSONErrorMsg:
		m.mode = 2
		m.err = msg.Err
		return m, m.progress.SetPercent(0.0)

	case items.ProjectDeleteDoneMsg:
		m.list.RemoveItem(m.list.GlobalIndex())
		m.status = "🗑  Project deleted"
		return m, m.progress.SetPercent(0.5)

	case items.ProjectDeleteErrorMsg:
		m.mode = 2
		m.err = msg.Err
		return m, m.progress.SetPercent(0.0)

	case tea.WindowSizeMsg:
		h, v := appStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch m.mode {
		case modeConfirmDelete:
			switch msg.String() {
			case "y", "Y":
				if m.list.SelectedItem() != nil {
					cmds = append(cmds,
						m.progress.SetPercent(0.10),
						tickCmd(),
						m.list.SelectedItem().(*items.Project).DeleteProjectFromFS(),
						git.CommitCmd(m.list.SelectedItem().(*items.Project).Id(),
							"delete: "+m.list.SelectedItem().(*items.Project).Title()),
					)
					m.status = ""
				}

				m.mode = modeNormal
				return m, tea.Batch(cmds...)

			case "n", "N", "esc", "q":
				m.mode = modeNormal
				return m, nil
			}

		case modeNormal:
			// Don't match any of the keys below if we're actively filtering.
			if m.list.FilterState() == list.Filtering {
				break
			}

			if msg.Type == tea.KeyCtrlC {
				return m, tea.Quit
			}

			switch {
			case key.Matches(msg, m.keys.quit):
				if m.selected {
					m.selected = !m.selected
					return m, nil
				}

				return m, tea.Quit

			case key.Matches(msg, m.keys.toggleHelpMenu):
				m.list.SetShowHelp(!m.list.ShowHelp())
				return m, nil

			case key.Matches(msg, m.keys.chooseProject) || key.Matches(msg, m.keys.chooseProjectVim):
				if m.list.SelectedItem() != nil {
					listModel := newTaskListModel(m.list.SelectedItem().(*items.Project), &m)
					return listModel, tea.WindowSize()
				}
				return m, nil

			case key.Matches(msg, m.keys.deleteProject):
				if m.list.SelectedItem() != nil {
					m.mode = modeConfirmDelete
				}
				return m, nil

			case key.Matches(msg, m.keys.editProject):
				if m.list.SelectedItem() != nil {
					// Switch to formModel for editing.
					formModel := newProjectFormModel(m.list.SelectedItem().(*items.Project), &m, true)
					return formModel, tea.WindowSize()
				}

			case key.Matches(msg, m.keys.addProject):
				project := &items.Project{
					ProjectId:          uuid.NewString(),
					ProjectTitle:       "",
					ProjectDescription: "",
				}
				formModel := newProjectFormModel(project, &m, false)
				return formModel, tea.WindowSize()
			}
		}
	}

	newListModel, cmd := m.list.Update(msg)
	m.list = newListModel
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// View renders the current UI state of the project list,
// including list view, progress bar, and any status messages.
func (m projectListModel) View() string {
	centeredStyle := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center)

	// Display progress bar at 100%
	if m.progressDone && m.waitingAfterDone {
		return centeredStyle.Bold(true).Render(textStyleGreen(m.status) + "\n\n" + m.progress.ViewAs(1.0))
	}

	// Display progress bar if not at 0%
	if m.progress.Percent() != 0.0 {
		return centeredStyle.Bold(true).Render(textStyleGreen(m.status) + "\n\n" + m.progress.View())
	}

	// Display deletion confirm view.
	if m.mode == modeConfirmDelete {
		selected := m.list.SelectedItem().(*items.Project)

		return centeredStyle.Render(
			fmt.Sprintf("Delete \"%s\"?\n\n", selected.Title()) +
				textStyleRed("[y] Yes") + "    " + textStyleGreen("[n] No"),
		)
	}

	// Display git error view
	if m.mode == modeGitError {
		content := "An error occurred while executing git:\n\n" +
			m.err.Error() + "\n\n" +
			"Please commit manually!"

		return centeredStyle.Render(content)
	}

	// Display list view.
	return appStyle.Render(m.list.View())
}
