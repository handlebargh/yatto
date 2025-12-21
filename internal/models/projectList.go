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

package models

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
	"github.com/handlebargh/yatto/internal/colors"
	"github.com/handlebargh/yatto/internal/helpers"
	"github.com/handlebargh/yatto/internal/items"
	"github.com/handlebargh/yatto/internal/vcs"
	"github.com/spf13/viper"
)

const projectDescLength = 200

// projectListKeyMap defines the key bindings
// used in the project list UI model.
type projectListKeyMap struct {
	quit           key.Binding
	toggleHelpMenu key.Binding
	addProject     key.Binding
	editProject    key.Binding
	chooseProject  key.Binding
	deleteProject  key.Binding
	prevPage       key.Binding
	nextPage       key.Binding
	toggleSelect   key.Binding
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
			key.WithHelp("D", "delete selected projects"),
		),
		chooseProject: key.NewBinding(
			key.WithKeys("enter", "l"),
			key.WithHelp("enter/l", "choose project"),
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
		prevPage: key.NewBinding(
			key.WithKeys("left", "pgup", "b", "u"),
			key.WithHelp("←/pgup/b/u", "prev page"),
		),
		nextPage: key.NewBinding(
			key.WithKeys("right", "pgdown", "f", "d"),
			key.WithHelp("→/pgdn/f/d", "next page"),
		),
		toggleSelect: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("space", "select/deselect"),
		),
	}
}

// customProjectDelegate implements a custom
// renderer for items in the project list.
type customProjectDelegate struct {
	list.DefaultDelegate
	parent *ProjectListModel
}

func (d customProjectDelegate) Height() int {
	return 3
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

	color := helpers.GetColorCode(projectItem.Color)

	availableWidth := max(m.Width(), 40)
	leftWidth := max(availableWidth-40, 20)

	// Check if item is selected
	_, selected := d.parent.selectedItems[projectItem.ID]

	marker := ""
	indent := 0
	if selected {
		marker = lipgloss.NewStyle().
			Foreground(colors.Red()).
			Render("⟹  ")
		indent = 3
	}

	// Base styles.
	listTitleStyle := lipgloss.NewStyle().
		Foreground(color).
		Padding(0, 1).
		Width(leftWidth - indent)

	listDescStyle := lipgloss.NewStyle().
		Padding(0, 1).
		MarginLeft(indent).
		Width(leftWidth - indent).
		Height(2)

	listItemInfoStyle := lipgloss.NewStyle().
		Width(40)

	if index == m.GlobalIndex() {
		listTitleStyle = listTitleStyle.
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(color)
		listDescStyle = listDescStyle.
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(color)

	} else if !selected {
		listTitleStyle = listTitleStyle.MarginLeft(1)
		listDescStyle = listDescStyle.MarginLeft(1)
	}

	var left strings.Builder

	left.WriteString(marker)
	left.WriteString(listTitleStyle.Render(projectItem.Title))
	left.WriteString("\n")
	left.WriteString(listDescStyle.Render(projectItem.CropDescription(projectDescLength)))

	numTasks, numCompletedTasks, numDueTasks, err := projectItem.NumOfTasks(d.parent.config)
	if err != nil {
		m.NewStatusMessage(
			lipgloss.NewStyle().Foreground(colors.Red()).Render(
				fmt.Sprintf("Error gathering task info for project %s", projectItem.Title),
			),
		)
	}

	var taskDueMessage string
	if numDueTasks > 0 {
		if numDueTasks == 1 {
			taskDueMessage = lipgloss.NewStyle().
				Foreground(colors.Red()).
				Render("1 task due today")
		} else {
			taskDueMessage = lipgloss.NewStyle().
				Foreground(colors.Red()).
				Render(fmt.Sprintf("%d tasks due today", numDueTasks))
		}
	}

	taskTotalCompleteMessage := fmt.Sprintf("%d/%d tasks completed", numCompletedTasks, numTasks)
	if numCompletedTasks == numTasks {
		taskTotalCompleteMessage = lipgloss.NewStyle().Foreground(colors.Green()).Render(taskTotalCompleteMessage)
	}

	var right strings.Builder

	right.WriteString(listItemInfoStyle.Render(taskTotalCompleteMessage))
	right.WriteString("\n")
	right.WriteString(taskDueMessage)

	row := lipgloss.NewStyle().
		Width(availableWidth).
		Render(
			lipgloss.JoinHorizontal(
				lipgloss.Top,
				left.String(),
				listItemInfoStyle.Render(right.String()),
			),
		)

	_, err = fmt.Fprint(w, row)
	if err != nil {
		panic(err)
	}
}

// ProjectListModel defines the TUI model used to
// manage and interact with projects.
type ProjectListModel struct {
	config        *viper.Viper
	list          list.Model
	selected      bool
	keys          *projectListKeyMap
	mode          mode
	cmdOutput     string
	err           error
	spinner       spinner.Model
	spinning      bool
	status        string
	width, height int
	selectedItems map[string]*items.Project

	renderer *glamour.TermRenderer
}

// InitialProjectListModel returns an initialized projectListModel
// with all necessary state and UI settings.
func InitialProjectListModel(v *viper.Viper) ProjectListModel {
	listKeys := newProjectListKeyMap()

	// Read all projects from FS to populate project list.
	projects := helpers.ReadProjectsFromFS(v)
	var listItems []list.Item

	for _, project := range projects {
		listItems = append(listItems, &project)
	}

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(colors.Orange())

	m := ProjectListModel{
		config:        v,
		keys:          listKeys,
		spinner:       sp,
		spinning:      false,
		selectedItems: make(map[string]*items.Project),
	}

	itemList := list.New(
		listItems,
		customProjectDelegate{DefaultDelegate: list.NewDefaultDelegate(), parent: &m},
		0,
		0,
	)
	itemList.SetShowPagination(true)
	itemList.SetShowTitle(true)
	itemList.SetShowStatusBar(true)
	itemList.SetStatusBarItemName("project", "projects")
	itemList.StatusMessageLifetime = 3 * time.Second
	itemList.Title = "Projects"
	itemList.Styles.Title = lipgloss.NewStyle().
		Foreground(colors.BadgeText()).
		Background(colors.Green()).
		Padding(0, 1)
	// Disable the quit keybindings, so we can implement our own.
	itemList.DisableQuitKeybindings()
	// Set our own prev/next page keys.
	itemList.KeyMap.NextPage = listKeys.nextPage
	itemList.KeyMap.PrevPage = listKeys.prevPage
	itemList.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			listKeys.quit,
		}
	}
	itemList.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			listKeys.toggleHelpMenu,
			listKeys.chooseProject,
			listKeys.addProject,
			listKeys.editProject,
			listKeys.deleteProject,
			listKeys.toggleSelect,
		}
	}

	m.list = itemList

	renderer, err := glamour.NewTermRenderer(glamour.WithAutoStyle())
	if err != nil {
		panic(err)
	}

	m.renderer = renderer

	return m
}

// Init initializes the Bubble Tea program
// for the project list model.
func (m ProjectListModel) Init() tea.Cmd {
	return tea.Batch(
		vcs.InitCmd(m.config),
	)
}

// Update handles incoming messages and updates
// the project list model state accordingly.
func (m ProjectListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case spinner.TickMsg:
		if m.spinning {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case doneWaitingMsg:
		m.spinning = false
		return m, nil

	case vcs.InitDoneMsg:
		return m, nil

	case vcs.InitErrorMsg:
		m.mode = 2
		m.err = msg.Err
		return m, nil

	case vcs.CommitDoneMsg:
		// Remove all map entries after successful commit.
		for k := range m.selectedItems {
			delete(m.selectedItems, k)
		}
		m.status = "🗘  Changes committed"

		// Wait 1 second before fully stopping spinner
		return m, tea.Tick(time.Second, func(time.Time) tea.Msg {
			return doneWaitingMsg{}
		})

	case vcs.CommitErrorMsg:
		m.mode = modeBackendError
		m.cmdOutput = msg.CmdOutput
		m.err = msg.Err
		m.spinning = false
		return m, nil

	case vcs.PullErrorMsg:
		m.mode = modeBackendError
		m.cmdOutput = msg.CmdOutput
		m.err = msg.Err
		m.spinning = false
		return m, nil

	case vcs.PushErrorMsg:
		m.mode = modeBackendError
		m.cmdOutput = msg.CmdOutput
		m.err = msg.Err
		m.spinning = false
		return m, nil

	case items.WriteProjectJSONDoneMsg:
		switch msg.Kind {
		case "create":
			m.list.InsertItem(0, &msg.Project)
			m.status = "🗸  Project created ― committing changes"
			return m, nil

		case "update":
			m.status = "🗸  Project updated ― committing changes"
			return m, nil
		}
		return m, nil

	case items.WriteProjectJSONErrorMsg:
		m.mode = 2
		m.err = msg.Err
		return m, nil

	case items.ProjectDeleteDoneMsg:
		for i, project := range m.selectedItems {
			if idx := project.FindListIndexByID(m.list.Items()); idx >= 0 {
				m.list.RemoveItem(idx)
				delete(m.selectedItems, i)
			}
		}
		m.status = "✘ Project(s) deleted ― committing changes"
		return m, nil

	case items.ProjectDeleteErrorMsg:
		m.mode = 2
		m.err = msg.Err
		m.spinning = false
		return m, nil

	case tea.WindowSizeMsg:
		h, v := appStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}

		if m.spinning {
			break
		}

		switch m.mode {
		case modeBackendError:
			switch msg.String() {
			case "esc", "q":
				m.mode = modeNormal
				return m, nil
			}

		case modeConfirmDelete:
			switch msg.String() {
			case "y", "Y":
				if len(m.selectedItems) == 0 {

					m.mode = modeNormal
					return m, nil
				}

				var projectNames, projectPaths []string
				var deleteCmds []tea.Cmd
				for _, item := range m.selectedItems {
					projectNames = append(projectNames, item.Title)
					projectPaths = append(projectPaths, item.ID)
					deleteCmds = append(deleteCmds, item.DeleteProjectFromFS(m.config))
				}

				message := fmt.Sprintf("delete: %d project(s)\n\n- %s", len(projectNames), strings.Join(projectNames, "\n- "))

				m.spinning = true

				cmds = append(cmds, m.spinner.Tick)
				cmds = append(cmds, deleteCmds...)
				cmds = append(cmds, vcs.CommitCmd(m.config, message, projectPaths...))

				m.status = ""

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

			case key.Matches(msg, m.keys.chooseProject):
				if m.list.SelectedItem() != nil {
					listModel := newTaskListModel(m.list.SelectedItem().(*items.Project), &m)
					return listModel, tea.WindowSize()
				}
				return m, nil

			case key.Matches(msg, m.keys.deleteProject):
				if len(m.selectedItems) > 0 {
					m.mode = modeConfirmDelete
				} else {
					cmds = append(cmds, m.list.NewStatusMessage(lipgloss.NewStyle().
						Foreground(colors.Red()).
						Render("No project selected")))
				}

				return m, tea.Batch(cmds...)

			case key.Matches(msg, m.keys.editProject):
				if m.list.SelectedItem() != nil {
					// Switch to formModel for editing.
					formModel := newProjectFormModel(m.list.SelectedItem().(*items.Project), &m, true)
					return formModel, tea.WindowSize()
				}

			case key.Matches(msg, m.keys.addProject):
				project := &items.Project{
					ID:          uuid.NewString(),
					Title:       "",
					Description: "",
				}
				formModel := newProjectFormModel(project, &m, false)
				return formModel, tea.WindowSize()

			case key.Matches(msg, m.keys.toggleSelect):
				if m.list.SelectedItem() != nil {
					p := m.list.SelectedItem().(*items.Project)

					if _, ok := m.selectedItems[p.ID]; ok {
						delete(m.selectedItems, p.ID)
					} else {
						m.selectedItems[p.ID] = p
					}
					return m, nil
				}
			}
		default:
			panic("unhandled default case in project list")
		}
	}

	newListModel, cmd := m.list.Update(msg)
	m.list = newListModel
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// View renders the current UI state of the project list,
// including list view, progress bar, and any status messages.
func (m ProjectListModel) View() string {
	centeredStyle := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center)

	// Spinner active view
	if m.spinning {
		return centeredStyle.
			Render(fmt.Sprintf("%s  %s", m.spinner.View(), m.status))
	}

	// Display deletion confirm view.
	if m.mode == modeConfirmDelete {
		if len(m.selectedItems) > 0 {
			return centeredStyle.Render(
				fmt.Sprintf("Delete %d project(s)?\n\n%s%s%s", len(m.selectedItems),
					"[y] Yes",
					"    ",
					"[n] No",
				))
		}
	}

	// Display VCS error view
	if m.mode == modeBackendError {
		var e strings.Builder

		e.WriteString("An error occurred during a backend operation:")
		e.WriteString("\n\n")
		e.WriteString(m.cmdOutput)
		e.WriteString("\n\n")
		e.WriteString("Please commit manually!")

		return centeredStyle.Render(e.String())
	}

	// Display list view.
	return appStyle.Render(m.list.View())
}
