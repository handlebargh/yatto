// Copyright 2025-2026 handlebargh and contributors
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

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/progress"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/glamour/v2"
	"charm.land/lipgloss/v2"
	"github.com/google/uuid"
	"github.com/handlebargh/yatto/internal/colors"
	"github.com/handlebargh/yatto/internal/helpers"
	"github.com/handlebargh/yatto/internal/items"
	"github.com/handlebargh/yatto/internal/vcs"
	"github.com/spf13/viper"
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
			key.WithKeys(" ", "space"),
			key.WithHelp("space", "select/deselect"),
		),
	}
}

// rendererReadyMsg is sent when the glamour terminal renderer has been
// successfully initialized and is ready for use.
type rendererReadyMsg struct {
	renderer *glamour.TermRenderer
}

// projectListState holds shared mutable state that must remain consistent
// between the ProjectListModel and its customProjectDelegate across value
// copies. Fields are accessed via pointer to avoid stale reads after updates.
type projectListState struct {
	taskStats     map[string]items.TaskStats
	selectedItems map[string]*items.Project
	renderer      *glamour.TermRenderer
}

// customProjectDelegate implements a custom
// renderer for items in the project list.
type customProjectDelegate struct {
	list.DefaultDelegate
	parent *ProjectListModel
}

func (d customProjectDelegate) Height() int {
	return 5
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

	availableWidth := max(m.Width(), 50)

	// Check if item is selected
	_, selected := d.parent.state.selectedItems[projectItem.ID]

	// === Creative Modern Design ===
	// Each project as a clean card with rounded corners
	// Layout: [icon] Title
	//         Description
	//         [───────>   ] x/y (zz%) • N due

	// Border styling - dynamic based on state
	// Selected items get double-line borders
	var borderStyle lipgloss.Style
	var cornerTL, cornerTR, cornerBL, cornerBR, borderH, borderV string

	switch {
	case selected && index == m.GlobalIndex():
		// Selected and on cursor: double border with green color
		borderStyle = lipgloss.NewStyle().Foreground(colors.Green())
		cornerTL = borderStyle.Render("╔")
		cornerTR = borderStyle.Render("╗")
		cornerBL = borderStyle.Render("╚")
		cornerBR = borderStyle.Render("╝")
		borderH = borderStyle.Render("═")
		borderV = borderStyle.Render("║")
	case selected:
		// Selected: double border with red color
		borderStyle = lipgloss.NewStyle().Foreground(colors.Red())
		cornerTL = borderStyle.Render("╔")
		cornerTR = borderStyle.Render("╗")
		cornerBL = borderStyle.Render("╚")
		cornerBR = borderStyle.Render("╝")
		borderH = borderStyle.Render("═")
		borderV = borderStyle.Render("║")
	case index == m.GlobalIndex():
		// Current item: single border with project color
		borderStyle = lipgloss.NewStyle().Foreground(colors.Green())
		cornerTL = borderStyle.Render("╭")
		cornerTR = borderStyle.Render("╮")
		cornerBL = borderStyle.Render("╰")
		cornerBR = borderStyle.Render("╯")
		borderH = borderStyle.Render("─")
		borderV = borderStyle.Render("│")
	default:
		// Normal: subtle gray single border
		borderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#333333"))
		cornerTL = borderStyle.Render("╭")
		cornerTR = borderStyle.Render("╮")
		cornerBL = borderStyle.Render("╰")
		cornerBR = borderStyle.Render("╯")
		borderH = borderStyle.Render("─")
		borderV = borderStyle.Render("│")
	}

	// All lines must be exactly availableWidth characters
	// Content between borders: availableWidth - 2 (for the two border characters)
	contentWidth := availableWidth - 2

	stats := d.parent.state.taskStats[projectItem.ID]
	numTasks := stats.Total
	numCompletedTasks := stats.Completed
	numDueTasks := stats.Due

	var progressPercent float64
	if numTasks > 0 {
		progressPercent = float64(numCompletedTasks) / float64(numTasks)
	}

	// Progress bar - colored based on completion
	var progressBar progress.Model
	switch {
	case numCompletedTasks == numTasks && numTasks > 0:
		progressBar = d.parent.progressGreen
	case progressPercent < 0.33:
		progressBar = d.parent.progressRed
	case progressPercent < 0.60:
		progressBar = d.parent.progressOrange
	case progressPercent < 1:
		progressBar = d.parent.progressYellow
	default:
		progressBar = d.parent.progressGreen
	}
	progressBar.ShowPercentage = false
	progressBarView := progressBar.ViewAs(progressPercent)

	// Selection indicator - use colored circle for all states
	var indicator string
	switch {
	case index == m.GlobalIndex():
		// Current: colored circle matching project
		indicator = lipgloss.NewStyle().Foreground(color).Render("●")
	case selected:
		// Selected: red circle
		indicator = lipgloss.NewStyle().Foreground(colors.Red()).Render("●")
	default:
		// Normal: subtle gray circle
		indicator = lipgloss.NewStyle().Foreground(lipgloss.Color("#555555")).Render("○")
	}

	// Title
	titleStyle := lipgloss.NewStyle().
		Foreground(color).
		Bold(true)
	title := titleStyle.Render(projectItem.Title)

	// Description
	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#777777"))
	desc := descStyle.Render(projectItem.CropDescription(60))

	// Stats line
	var statsStr string
	if numTasks > 0 {
		percentInt := int(progressPercent * 100)
		if numCompletedTasks == numTasks {
			statsStr = lipgloss.NewStyle().
				Foreground(colors.Green()).
				Render(fmt.Sprintf("✓ %d/%d (%d%%)", numCompletedTasks, numTasks, percentInt))
		} else {
			statsStr = fmt.Sprintf("%d/%d (%d%%)", numCompletedTasks, numTasks, percentInt)
		}
	} else {
		statsStr = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			Render("empty")
	}

	// Due indicator
	var dueStr string
	if numDueTasks > 0 {
		dueStr = lipgloss.NewStyle().
			Foreground(colors.Red()).
			Render(fmt.Sprintf(" • %d due", numDueTasks))
	}

	// Build the card
	var sb strings.Builder

	// Top border
	sb.WriteString(cornerTL)
	sb.WriteString(strings.Repeat(borderH, availableWidth-2))
	sb.WriteString(cornerTR)
	sb.WriteString("\n")

	// Title line: │ indicator title padding │
	// Total: 1 + 1 + len(ind) + 1 + len(title) + pad + 1 = availableWidth
	// So: pad = availableWidth - 5 - len(ind) - len(title)
	sb.WriteString(borderV)
	sb.WriteString(" ")
	sb.WriteString(indicator)
	sb.WriteString(" ")
	sb.WriteString(title)
	pad := contentWidth - (1 + lipgloss.Width(indicator) + 1 + lipgloss.Width(title))
	if pad > 0 {
		sb.WriteString(strings.Repeat(" ", pad))
	}
	sb.WriteString(borderV)
	sb.WriteString("\n")

	// Description line: │  desc padding │
	sb.WriteString(borderV)
	sb.WriteString(" ")
	sb.WriteString(strings.Repeat(" ", 2))
	sb.WriteString(desc)
	pad = contentWidth - (1 + 2 + lipgloss.Width(desc))
	if pad > 0 {
		sb.WriteString(strings.Repeat(" ", pad))
	}
	sb.WriteString(borderV)
	sb.WriteString("\n")

	// Padding line (blank) for visual breathing room
	sb.WriteString(borderV)
	sb.WriteString(strings.Repeat(" ", contentWidth))
	sb.WriteString(borderV)
	sb.WriteString("\n")

	// Progress line: │  [====      ] stats │
	// All progress bars are 30 chars. Use fixed-width stats for alignment.
	// Layout: │  [padding][30-char bar][space][20-char stats]│
	// Total: 1 + 1 + 2 + padding + 30 + 1 + 20 + 1 = availableWidth
	// So: padding = availableWidth - 56
	statsFixed := lipgloss.NewStyle().Width(20).Render(statsStr + dueStr)
	sb.WriteString(borderV)
	sb.WriteString(" ")
	sb.WriteString(strings.Repeat(" ", 2))
	padBefore := availableWidth - 56 // 5 for borders/spaces + 30 bar + 1 space + 20 stats
	if padBefore > 0 {
		sb.WriteString(strings.Repeat(" ", padBefore))
	}
	sb.WriteString(progressBarView)
	sb.WriteString(" ")
	sb.WriteString(statsFixed)
	sb.WriteString(borderV)
	sb.WriteString("\n")

	// Bottom border
	sb.WriteString(cornerBL)
	sb.WriteString(strings.Repeat(borderH, availableWidth-2))
	sb.WriteString(cornerBR)

	_, err := fmt.Fprint(w, sb.String())
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
	state         *projectListState

	progressRed    progress.Model
	progressOrange progress.Model
	progressYellow progress.Model
	progressGreen  progress.Model
	// Fields for tracking pending commit after project write (from form)
	commitMessage string
	commitPath    string
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
		config:   v,
		keys:     listKeys,
		spinner:  sp,
		spinning: false,
		state: &projectListState{
			taskStats:     make(map[string]items.TaskStats),
			selectedItems: make(map[string]*items.Project),
		},
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

	m.progressRed = progress.New(progress.WithColors(colors.Red()), progress.WithWidth(30))
	m.progressOrange = progress.New(progress.WithColors(colors.Orange()), progress.WithWidth(30))
	m.progressYellow = progress.New(progress.WithColors(colors.Yellow()), progress.WithWidth(30))
	m.progressGreen = progress.New(progress.WithColors(colors.Green()), progress.WithWidth(30))

	return m
}

// Init initializes the Bubble Tea program
// for the project list model.
func (m ProjectListModel) Init() tea.Cmd {
	projects := m.allProjects()
	return tea.Batch(
		vcs.InitCmd(m.config),
		items.LoadAllTaskStatsCmd(m.config, projects),
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

	case rendererReadyMsg:
		m.state.renderer = msg.renderer
		return m, nil

	case doneWaitingMsg:
		m.spinning = false
		return m, nil

	case returnedToProjectListMsg:
		return m, items.LoadAllTaskStatsCmd(m.config, m.allProjects())

	case vcs.InitDoneMsg:
		return m, nil

	case vcs.InitErrorMsg:
		m.mode = 2
		m.err = msg.Err
		return m, nil

	case vcs.CommitDoneMsg:
		// Remove all map entries after successful commit.
		for k := range m.state.selectedItems {
			delete(m.state.selectedItems, k)
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

		case "update":
			m.status = "🗸  Project updated ― committing changes"
		}

		// If we have a pending commit (from project create/update via form),
		// trigger it now that the file has been written
		if m.commitPath != "" && (msg.Kind == "create" || msg.Kind == "update") {
			cmds := []tea.Cmd{
				vcs.CommitCmd(
					m.config,
					m.commitMessage,
					m.commitPath,
				),
				items.LoadAllTaskStatsCmd(m.config, m.allProjects()),
			}
			// Clear the pending commit info
			m.commitMessage = ""
			m.commitPath = ""
			return m, tea.Batch(cmds...)
		}

		return m, items.LoadAllTaskStatsCmd(m.config, m.allProjects())

	case items.WriteProjectJSONErrorMsg:
		m.mode = modeBackendError
		m.err = msg.Err
		m.cmdOutput = msg.Err.Error()
		m.spinning = false
		return m, nil

	case items.ProjectDeleteDoneMsg:
		for i, project := range m.state.selectedItems {
			if idx := project.FindListIndexByID(m.list.Items()); idx >= 0 {
				m.list.RemoveItem(idx)
				delete(m.state.selectedItems, i)
			}
		}
		m.status = "✘ Project(s) deleted ― committing changes"
		return m, items.LoadAllTaskStatsCmd(m.config, m.allProjects())

	case items.ProjectDeleteErrorMsg:
		m.mode = 2
		m.err = msg.Err
		m.spinning = false
		return m, nil

	case items.TaskStatsDoneMsg:
		m.state.taskStats = msg.Stats
		return m, nil

	case items.TaskStatsErrorMsg:
		return m, nil

	case tea.WindowSizeMsg:
		h, v := appStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyPressMsg:
		if msg.Code == 'c' && msg.Mod == tea.ModCtrl {
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
				if len(m.state.selectedItems) == 0 {

					m.mode = modeNormal
					return m, nil
				}

				var projectNames, projectPaths []string
				var deleteCmds []tea.Cmd
				for _, item := range m.state.selectedItems {
					projectNames = append(projectNames, item.Title)
					projectPaths = append(projectPaths, item.ID)
					deleteCmds = append(deleteCmds, item.DeleteProjectFromFS(m.config))
				}

				message := fmt.Sprintf(
					"delete: %d project(s)\n\n- %s",
					len(projectNames),
					strings.Join(projectNames, "\n- "),
				)

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
					listModel := newTaskListModel(m.list.SelectedItem().(*items.Project), &m, m.width, m.height)
					return listModel, tea.RequestWindowSize
				}
				return m, nil

			case key.Matches(msg, m.keys.deleteProject):
				if len(m.state.selectedItems) > 0 {
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
					return formModel, tea.RequestWindowSize
				}

			case key.Matches(msg, m.keys.addProject):
				project := &items.Project{
					ID:          uuid.NewString(),
					Title:       "",
					Description: "",
				}
				formModel := newProjectFormModel(project, &m, false)
				return formModel, tea.RequestWindowSize

			case key.Matches(msg, m.keys.toggleSelect):
				if m.list.SelectedItem() != nil {
					p := m.list.SelectedItem().(*items.Project)

					if _, ok := m.state.selectedItems[p.ID]; ok {
						delete(m.state.selectedItems, p.ID)
					} else {
						m.state.selectedItems[p.ID] = p
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
func (m ProjectListModel) View() tea.View {
	centeredStyle := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center)

	// Spinner active view
	if m.spinning {
		content := centeredStyle.
			Render(fmt.Sprintf("%s  %s", m.spinner.View(), m.status))
		v := tea.NewView(content)
		v.AltScreen = true
		return v
	}

	// Display deletion confirm view.
	if m.mode == modeConfirmDelete {
		if len(m.state.selectedItems) > 0 {
			content := centeredStyle.Render(
				fmt.Sprintf("Delete %d project(s)?\n\n%s%s%s", len(m.state.selectedItems),
					"[y] Yes",
					"    ",
					"[n] No",
				))
			v := tea.NewView(content)
			v.AltScreen = true
			return v
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

		content := centeredStyle.Render(e.String())
		v := tea.NewView(content)
		v.AltScreen = true
		return v
	}

	// Display list view.
	content := appStyle.Render(m.list.View())
	v := tea.NewView(content)
	v.AltScreen = true
	return v
}
