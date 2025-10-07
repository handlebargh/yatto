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
	"cmp"
	"fmt"
	"io"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
	"github.com/handlebargh/yatto/internal/colors"
	"github.com/handlebargh/yatto/internal/helpers"
	"github.com/handlebargh/yatto/internal/items"
	"github.com/handlebargh/yatto/internal/vcs"
	"github.com/spf13/viper"
)

const taskEntryLength = 53

// taskListKeyMap defines the key bindings used in the task list view.
type taskListKeyMap struct {
	quit             key.Binding
	toggleHelpMenu   key.Binding
	addItem          key.Binding
	chooseItem       key.Binding
	editItem         key.Binding
	deleteItem       key.Binding
	sortByPriority   key.Binding
	sortByDueDate    key.Binding
	sortByAuthor     key.Binding
	sortByAssignee   key.Binding
	toggleInProgress key.Binding
	toggleComplete   key.Binding
	goBackVim        key.Binding
	prevPage         key.Binding
	nextPage         key.Binding
	toggleSelect     key.Binding
}

// newTaskListKeyMap initializes and returns a new key map for task list actions.
func newTaskListKeyMap() *taskListKeyMap {
	return &taskListKeyMap{
		quit: key.NewBinding(
			key.WithKeys("q", "esc"),
			key.WithHelp("q/esc", "go back"),
		),
		toggleComplete: key.NewBinding(
			key.WithKeys("C"),
			key.WithHelp("C", "toggle complete on selection"),
		),
		toggleInProgress: key.NewBinding(
			key.WithKeys("P"),
			key.WithHelp("P", "toggle in progress on selection"),
		),
		sortByPriority: key.NewBinding(
			key.WithKeys("alt+p"),
			key.WithHelp("alt+p", "sort by priority"),
		),
		sortByDueDate: key.NewBinding(
			key.WithKeys("alt+d"),
			key.WithHelp("alt+d", "sort by due date"),
		),
		sortByAuthor: key.NewBinding(
			key.WithKeys("alt+a"),
			key.WithHelp("alt+a", "sort by author"),
		),
		sortByAssignee: key.NewBinding(
			key.WithKeys("alt+A"),
			key.WithHelp("alt+A", "sort by assignee"),
		),
		deleteItem: key.NewBinding(
			key.WithKeys("D"),
			key.WithHelp("D", "delete selected tasks"),
		),
		editItem: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit task"),
		),
		chooseItem: key.NewBinding(
			key.WithKeys("enter", "l"),
			key.WithHelp("enter/l", "show task"),
		),
		addItem: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "add task"),
		),
		toggleHelpMenu: key.NewBinding(
			key.WithKeys("H"),
			key.WithHelp("H", "toggle help"),
		),
		goBackVim: key.NewBinding(
			key.WithKeys("h"),
			key.WithHelp("h", "go back"),
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

// customTaskDelegate is a custom list delegate for rendering task items.
type customTaskDelegate struct {
	list.DefaultDelegate
	parent *taskListModel
}

// Height returns the delegate's preferred height.
func (d customTaskDelegate) Height() int {
	showAuthor := viper.GetBool("author.show")
	showAssignee := viper.GetBool("assignee.show")

	if showAuthor && showAssignee {
		return 4
	}

	if showAuthor || showAssignee {
		return 3
	}

	return 2
}

// Render draws a single task item within the task list.
func (d customTaskDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	taskItem, ok := item.(*items.Task)
	if !ok {
		_, err := fmt.Fprint(w, "Invalid item\n")
		if err != nil {
			panic(err)
		}

		return
	}

	// Base styles.
	titleStyle := lipgloss.NewStyle().
		Padding(0, 1).
		Width(60)

	labelsStyle := lipgloss.NewStyle().
		Foreground(colors.Blue()).
		Padding(0, 1)

	authorStyle := lipgloss.NewStyle().
		Foreground(colors.Green()).
		Padding(0, 1)

	assigneeStyle := lipgloss.NewStyle().
		Foreground(colors.Orange()).
		Padding(0, 1)

	priorityValueStyle := lipgloss.NewStyle().
		Foreground(colors.BadgeText()).
		Padding(0, 1)

	switch taskItem.Priority {
	case "low":
		titleStyle = titleStyle.BorderForeground(colors.Indigo())
		labelsStyle = labelsStyle.BorderForeground(colors.Indigo())
		authorStyle = authorStyle.BorderForeground(colors.Indigo())
		assigneeStyle = assigneeStyle.BorderForeground(colors.Indigo())
		priorityValueStyle = priorityValueStyle.
			BorderForeground(colors.Indigo()).Background(colors.Indigo())
	case "medium":
		titleStyle = titleStyle.BorderForeground(colors.Orange())
		labelsStyle = labelsStyle.BorderForeground(colors.Orange())
		authorStyle = authorStyle.BorderForeground(colors.Orange())
		assigneeStyle = assigneeStyle.BorderForeground(colors.Orange())
		priorityValueStyle = priorityValueStyle.
			BorderForeground(colors.Orange()).Background(colors.Orange())
	case "high":
		titleStyle = titleStyle.BorderForeground(colors.Red())
		labelsStyle = labelsStyle.BorderForeground(colors.Red())
		authorStyle = authorStyle.BorderForeground(colors.Red())
		assigneeStyle = assigneeStyle.BorderForeground(colors.Red())
		priorityValueStyle = priorityValueStyle.
			BorderForeground(colors.Red()).Background(colors.Red())
	}

	if index == m.Index() {
		titleStyle = titleStyle.
			Border(lipgloss.NormalBorder(), false, false, false, true).
			MarginLeft(0)
		labelsStyle = labelsStyle.
			Border(lipgloss.NormalBorder(), false, false, false, true).
			MarginLeft(0)
		authorStyle = authorStyle.
			Border(lipgloss.NormalBorder(), false, false, false, true).
			MarginLeft(0)
		assigneeStyle = assigneeStyle.
			Border(lipgloss.NormalBorder(), false, false, false, true).
			MarginLeft(0)
	} else {
		titleStyle = titleStyle.MarginLeft(1)
		labelsStyle = labelsStyle.MarginLeft(1)
		authorStyle = authorStyle.MarginLeft(1)
		assigneeStyle = assigneeStyle.MarginLeft(1)
	}

	// Check if item is selected
	_, selected := d.parent.selectedItems[index]

	marker := ""
	if selected {
		marker = lipgloss.NewStyle().
			Foreground(colors.Red()).
			Render("⟹  ")
	}

	var left strings.Builder

	left.WriteString(marker)
	left.WriteString(titleStyle.Render(taskItem.CropTaskTitle(taskEntryLength)))
	left.WriteString("\n")
	left.WriteString(labelsStyle.Render(taskItem.CropTaskLabels(taskEntryLength)))

	if viper.GetBool("author.show") {
		left.WriteString("\n")
		left.WriteString(authorStyle.Render("Author: "))
		left.WriteString(taskItem.Author)
	}

	me, _ := vcs.UserEmail()
	if viper.GetBool("assignee.show") {
		left.WriteString("\n")
		left.WriteString(assigneeStyle.Render("Assignee: "))
		if taskItem.Assignee == me {
			left.WriteString(lipgloss.NewStyle().Foreground(colors.Red()).Render(taskItem.Assignee))
		} else {
			left.WriteString(taskItem.Assignee)
		}
	}

	var right strings.Builder

	right.WriteString(priorityValueStyle.Render(taskItem.Priority))

	now := time.Now()
	dueDate := taskItem.DueDate

	if dueDate != nil &&
		items.IsToday(dueDate) &&
		dueDate.After(now) {
		right.WriteString(lipgloss.NewStyle().
			Padding(0, 1).
			Background(colors.VividRed()).
			Foreground(colors.BadgeText()).
			Render("due today"))
	}

	if dueDate != nil && dueDate.Before(now) {
		right.WriteString(lipgloss.NewStyle().
			Padding(0, 1).
			Background(colors.VividRed()).
			Foreground(colors.BadgeText()).
			Render("overdue"))
	}

	if taskItem.InProgress {
		right.WriteString(lipgloss.NewStyle().
			Padding(0, 1).
			Background(colors.Blue()).
			Foreground(colors.BadgeText()).
			Render("in progress"))
	}

	if dueDate != nil &&
		!dueDate.Before(now) &&
		!items.IsToday(dueDate) {
		right.WriteString(lipgloss.NewStyle().
			Padding(0, 1).
			Background(colors.Yellow()).
			Foreground(colors.BadgeText()).
			Render("due in " + taskItem.DaysUntilToString() + " day(s)"))
	}

	if taskItem.Completed {
		right.Reset()
		right.WriteString(lipgloss.NewStyle().
			Padding(0, 1).
			Background(colors.Green()).
			Foreground(colors.BadgeText()).
			Render("completed"))
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Render(left.String()),
		right.String(),
	)

	_, err := fmt.Fprint(w, row)
	if err != nil {
		panic(err)
	}
}

// taskListModel represents the Bubble Tea model for the task list view.
type taskListModel struct {
	list             list.Model
	project          *items.Project
	projectModel     *ProjectListModel
	keys             *taskListKeyMap
	mode             mode
	cmdOutput        string
	err              error
	progress         progress.Model
	progressDone     bool
	waitingAfterDone bool
	status           string
	width, height    int
	selectedItems    map[int]*items.Task
}

// newTaskListModel creates a new taskListModel for the given project.
func newTaskListModel(project *items.Project, projectModel *ProjectListModel) taskListModel {
	listKeys := newTaskListKeyMap()

	tasks := project.ReadTasksFromFS()
	var listItems []list.Item

	for _, task := range tasks {
		listItems = append(listItems, &task)
	}

	color := helpers.GetColorCode(project.Color)

	titleStyleTasks := lipgloss.NewStyle().
		Foreground(colors.BadgeText()).
		Background(color).
		Padding(0, 1)

	m := taskListModel{
		project:       project,
		projectModel:  projectModel,
		keys:          listKeys,
		selectedItems: make(map[int]*items.Task),
		progress:      progress.New(progress.WithGradient("#FFA336", "#02BF87")),
	}

	itemList := list.New(
		listItems,
		customTaskDelegate{DefaultDelegate: list.NewDefaultDelegate(), parent: &m},
		0,
		0,
	)
	itemList.SetShowPagination(true)
	itemList.SetShowTitle(true)
	itemList.SetShowStatusBar(true)
	itemList.SetStatusBarItemName("task", "tasks")
	itemList.StatusMessageLifetime = 3 * time.Second
	itemList.Title = project.Title
	itemList.Styles.Title = titleStyleTasks
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
			listKeys.chooseItem,
			listKeys.goBackVim,
			listKeys.addItem,
			listKeys.editItem,
			listKeys.deleteItem,
			listKeys.sortByPriority,
			listKeys.sortByDueDate,
			listKeys.sortByAuthor,
			listKeys.sortByAssignee,
			listKeys.toggleInProgress,
			listKeys.toggleComplete,
			listKeys.toggleSelect,
		}
	}

	m.list = itemList

	return m
}

// Init initializes the taskListModel and returns an initial command.
func (m taskListModel) Init() tea.Cmd {
	return tickCmd()
}

// Update handles incoming messages and updates the taskListModel accordingly.
func (m taskListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tickMsg:
		if m.progress.Percent() >= 1.0 && !m.waitingAfterDone {
			m.progressDone = true
			m.waitingAfterDone = true

			// Return a timer command to keep displaying 100% progress
			// for half a second.
			return m, tea.Tick(time.Millisecond*500, func(_ time.Time) tea.Msg {
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

	case vcs.CommitDoneMsg:
		// Remove all map entries after successful commit.
		for k := range m.selectedItems {
			delete(m.selectedItems, k)
		}
		m.status = "🗘  Changes committed"
		m.progressDone = true
		return m, m.progress.SetPercent(1.0)

	case vcs.CommitErrorMsg:
		m.mode = 2
		m.cmdOutput = msg.CmdOutput
		m.err = msg.Err
		return m, m.progress.SetPercent(0.0)

	case vcs.PullErrorMsg:
		m.mode = 2
		m.cmdOutput = msg.CmdOutput
		m.err = msg.Err
		return m, m.progress.SetPercent(0.0)

	case vcs.PushErrorMsg:
		m.mode = 2
		m.cmdOutput = msg.CmdOutput
		m.err = msg.Err
		return m, m.progress.SetPercent(0.0)

	case items.WriteTaskJSONDoneMsg:
		switch msg.Kind {
		case "create":
			m.list.InsertItem(0, &msg.Task)
			m.status = "🗸  Task created"

		case "update":
			m.status = "🗸  Task updated"

		case "start":
			m.status = "🗸  Task(s) started"

		case "stop":
			m.status = "🗸  Task(s) stopped"

		case "complete":
			m.status = "🗸  Task(s) completed"

		case "reopen":
			m.status = "🗸  Task(s) reopened"

		default:
			return m, nil
		}
		return m, m.progress.SetPercent(0.5)

	case items.WriteTaskJSONErrorMsg:
		m.mode = 2
		m.err = msg.Err
		return m, m.progress.SetPercent(0.0)

	case items.TaskDeleteDoneMsg:
		for i, task := range m.selectedItems {
			if idx := task.FindListIndexByID(m.list.Items()); idx >= 0 {
				m.list.RemoveItem(idx)
				delete(m.selectedItems, i)
				m.status = "🗑  Task(s) deleted"

				return m, m.progress.SetPercent(0.5)
			}
		}

	case items.TaskDeleteErrorMsg:
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
				if len(m.selectedItems) == 0 {

					m.mode = modeNormal
					return m, nil
				}

				var taskNames, taskPaths []string
				var deleteCmds []tea.Cmd
				for _, item := range m.selectedItems {
					taskNames = append(taskNames, item.Title)
					taskPaths = append(taskPaths, filepath.Join(m.project.ID, item.ID+".json"))
					deleteCmds = append(deleteCmds, item.DeleteTaskFromFS(*m.project))
				}

				message := fmt.Sprintf("delete: %d task(s)\n\n- %s", len(taskNames), strings.Join(taskNames, "\n- "))
				cmds = append(cmds, m.progress.SetPercent(0.10), tickCmd())
				cmds = append(cmds, deleteCmds...)
				cmds = append(cmds, vcs.CommitCmd(message, taskPaths...))

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

			if msg.Type == tea.KeyCtrlC {
				return m, tea.Quit
			}

			switch {
			case key.Matches(msg, m.keys.quit):
				return m.projectModel, nil

			case key.Matches(msg, m.keys.goBackVim):
				return m.projectModel, nil

			case key.Matches(msg, m.keys.toggleHelpMenu):
				m.list.SetShowHelp(!m.list.ShowHelp())
				return m, nil

			case key.Matches(msg, m.keys.sortByPriority):
				sortTasksByKeys(&m.list, []string{"state", "priority"})
				return m, nil

			case key.Matches(msg, m.keys.sortByDueDate):
				sortTasksByKeys(&m.list, []string{"state", "dueDate"})
				return m, nil

			case key.Matches(msg, m.keys.sortByAuthor):
				sortTasksByKeys(&m.list, []string{"state", "author", "dueDate", "priority"})
				return m, nil

			case key.Matches(msg, m.keys.sortByAssignee):
				sortTasksByKeys(&m.list, []string{"state", "assignee", "dueDate", "priority"})
				return m, nil

			case key.Matches(msg, m.keys.chooseItem):
				if m.list.SelectedItem() != nil {
					markdown := m.list.SelectedItem().(*items.Task).TaskToMarkdown()
					pagerModel := newTaskPagerModel(markdown, &m)

					return pagerModel, tea.WindowSize()
				}
				return m, nil

			case key.Matches(msg, m.keys.toggleInProgress):
				m, cmds = m.toggleTasks(
					func(t *items.Task) { t.InProgress = !t.InProgress },
					func(t *items.Task) (bool, string) {
						if t.Completed {
							return false, "Cannot set completed task as in progress"
						}
						return true, ""
					},
					func(t *items.Task) string {
						if t.InProgress {
							return "start"
						}
						return "stop"
					},
					"progress",
				)

				return m, tea.Batch(cmds...)

			case key.Matches(msg, m.keys.toggleComplete):
				m, cmds = m.toggleTasks(
					func(t *items.Task) { t.Completed = !t.Completed; t.InProgress = false },
					func(_ *items.Task) (bool, string) { return true, "" },
					func(t *items.Task) string {
						if t.Completed {
							return "complete"
						}
						return "reopen"
					},
					"completion",
				)

				return m, tea.Batch(cmds...)

			case key.Matches(msg, m.keys.deleteItem):
				if len(m.selectedItems) > 0 {
					m.mode = modeConfirmDelete
				} else {
					cmds = append(cmds, m.list.NewStatusMessage(lipgloss.NewStyle().
						Foreground(colors.Red()).
						Render("No task selected")))
				}

				return m, tea.Batch(cmds...)

			case key.Matches(msg, m.keys.editItem):
				if m.list.SelectedItem() != nil {
					// Switch to formModel for editing.
					formModel := newTaskFormModel(m.list.SelectedItem().(*items.Task), &m, true)
					return formModel, tea.WindowSize()
				}

				return m, nil

			case key.Matches(msg, m.keys.addItem):
				task := &items.Task{
					ID:          uuid.NewString(),
					Title:       "",
					Description: "",
				}
				formModel := newTaskFormModel(task, &m, false)
				return formModel, tea.WindowSize()

			case key.Matches(msg, m.keys.toggleSelect):
				if m.list.SelectedItem() != nil {
					t := m.list.SelectedItem().(*items.Task)
					i := m.list.GlobalIndex()

					if _, ok := m.selectedItems[i]; ok {
						delete(m.selectedItems, i)
					} else {
						m.selectedItems[i] = t
					}
					return m, nil
				}
			}
		default:
			panic("unhandled default case in task list")
		}
	}

	newListModel, cmd := m.list.Update(msg)
	m.list = newListModel
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// View returns the string representation of the task list view.
func (m taskListModel) View() string {
	centeredStyle := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center)

	// Display progress bar at 100%
	if m.progressDone && m.waitingAfterDone {
		return centeredStyle.Bold(true).
			Render(lipgloss.NewStyle().Foreground(colors.Green()).Render(m.status) +
				"\n\n" + m.progress.ViewAs(1.0))
	}

	// Display progress bar if not at 0%
	if m.progress.Percent() != 0.0 {
		return centeredStyle.Bold(true).
			Render(lipgloss.NewStyle().Foreground(colors.Green()).Render(m.status) +
				"\n\n" + m.progress.View())
	}

	// Display deletion confirm view.
	if m.mode == modeConfirmDelete {
		// Check bulk selection
		if len(m.selectedItems) > 0 {
			return centeredStyle.Render(
				fmt.Sprintf("Delete %d task(s)?\n\n%s%s%s", len(m.selectedItems),
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

// sortTasksByKey sorts the tasks in the list model by a specified keys.
// Valid keys include "priority", "dueDate", and "state".
func sortTasksByKeys(m *list.Model, keys []string) {
	selected := m.SelectedItem()
	listItems := m.Items()

	var tasks []*items.Task
	for _, item := range listItems {
		if task, ok := item.(*items.Task); ok {
			tasks = append(tasks, task)
		}
	}

	me, _ := vcs.UserEmail()

	slices.SortStableFunc(tasks, func(x, y *items.Task) int {
		for _, k := range keys {
			var cmpResult int
			switch k {
			case "state":
				switch {
				case x.Completed && !y.Completed:
					cmpResult = 1
				case !x.Completed && y.Completed:
					cmpResult = -1
				case x.InProgress && !y.InProgress:
					cmpResult = -1
				case !x.InProgress && y.InProgress:
					cmpResult = 1
				default:
					cmpResult = 0
				}
			case "assignee":
				switch {
				case x.Assignee == "" && y.Assignee != "":
					cmpResult = 1
				case x.Assignee != "" && y.Assignee == "":
					cmpResult = -1
				case x.Assignee == me && y.Assignee != me:
					cmpResult = -1
				case x.Assignee != me && y.Assignee == me:
					cmpResult = 1
				default:
					cmpResult = strings.Compare(strings.ToLower(x.Assignee), strings.ToLower(y.Assignee))
				}
			case "dueDate":
				dx, dy := x.DueDate, y.DueDate
				switch {
				case dx == nil && dy != nil:
					cmpResult = 1
				case dx != nil && dy == nil:
					cmpResult = -1
				case dx != nil && dy != nil:
					switch {
					case dx.Before(*dy):
						cmpResult = -1
					case dx.After(*dy):
						cmpResult = 1
					default:
						cmpResult = 0
					}
				default:
					cmpResult = 0
				}
			case "priority":
				if x.Completed != y.Completed {
					if x.Completed {
						cmpResult = 1
					} else {
						cmpResult = -1
					}
				} else {
					cmpResult = cmp.Compare(y.PriorityValue(), x.PriorityValue())
				}
			case "author":
				switch {
				case x.Author == "" && y.Author != "":
					cmpResult = 1
				case x.Author != "" && y.Author == "":
					cmpResult = -1
				case x.Author == me && y.Author != me:
					cmpResult = -1
				case x.Author != me && y.Author == me:
					cmpResult = 1
				default:
					cmpResult = strings.Compare(strings.ToLower(x.Author), strings.ToLower(y.Author))
				}
			}
			if cmpResult != 0 {
				return cmpResult
			}
		}
		return 0
	})

	// Convert back to list.Item and re-select
	sortedItems := make([]list.Item, len(tasks))
	for i, t := range tasks {
		sortedItems[i] = t
	}
	m.SetItems(sortedItems)

	// Reselect the previously selected task
	if selectedTask, ok := selected.(*items.Task); ok {
		for i, item := range sortedItems {
			if task, ok := item.(*items.Task); ok && task.ID == selectedTask.ID {
				m.Select(i)
				break
			}
		}
	}
}

// toggleTasks applies a toggle operation to all selected tasks in the task list.
//
// Parameters:
//   - toggleFunc: a function that modifies a task (e.g., toggling InProgress or Completed).
//   - precondition: a function that checks whether the task can be toggled; returns
//     a bool indicating if the task passes the check, and a string message if not.
//   - commitKind: a function that determines the kind of action for the task (used
//     in writing JSON and commit messages).
//   - actionName: a string describing the type of action (e.g., "progress" or "completion")
//     used in the commit message.
//
// The function returns an updated taskListModel and a slice of tea.Cmds that perform
// the necessary operations, including writing JSON, updating progress, and creating
// a VCS commit. If no tasks are selected, it returns a status message and no other
// operations.
func (m taskListModel) toggleTasks(
	toggleFunc func(*items.Task),
	precondition func(*items.Task) (bool, string),
	commitKind func(*items.Task) string,
	actionName string,
) (taskListModel, []tea.Cmd) {
	if len(m.selectedItems) == 0 {
		return m, []tea.Cmd{
			m.list.NewStatusMessage(lipgloss.NewStyle().
				Foreground(colors.Red()).
				Render("No task selected")),
		}
	}

	var cmds, writeCmds []tea.Cmd
	var taskPaths, taskNames []string

	for _, t := range m.selectedItems {
		ok, msg := precondition(t)
		if !ok {
			cmds = append(cmds, m.list.NewStatusMessage(lipgloss.NewStyle().
				Foreground(colors.Red()).
				Render(msg)))

			return m, cmds
		}

		toggleFunc(t)
		json := t.MarshalTask()
		writeCmds = append(writeCmds, t.WriteTaskJSON(json, *m.project, commitKind(t)))
		taskPaths = append(taskPaths, filepath.Join(m.project.ID, t.ID+".json"))
		taskNames = append(taskNames, t.Title)
	}

	commitMsg := fmt.Sprintf("Change %s state of %d task(s)\n\n- %s",
		actionName, len(taskNames), strings.Join(taskNames, "\n- "))

	cmds = append(cmds, tickCmd(), m.progress.SetPercent(0.10))
	cmds = append(cmds, writeCmds...)
	cmds = append(cmds, vcs.CommitCmd(commitMsg, taskPaths...))

	return m, cmds
}
