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
			key.WithHelp("C", "toggle complete"),
		),
		toggleInProgress: key.NewBinding(
			key.WithKeys("P"),
			key.WithHelp("P", "toggle in progress"),
		),
		sortByPriority: key.NewBinding(
			key.WithKeys("alt+p"),
			key.WithHelp("alt+p", "sort by priority"),
		),
		sortByDueDate: key.NewBinding(
			key.WithKeys("alt+d"),
			key.WithHelp("alt+d", "sort by due date"),
		),
		deleteItem: key.NewBinding(
			key.WithKeys("D"),
			key.WithHelp("D", "delete"),
		),
		editItem: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit"),
		),
		chooseItem: key.NewBinding(
			key.WithKeys("enter", "l"),
			key.WithHelp("enter/l", "show"),
		),
		addItem: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "add item"),
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
			key.WithHelp("space", "toggle select"),
		),
	}
}

// customTaskDelegate is a custom list delegate for rendering task items.
type customTaskDelegate struct {
	list.DefaultDelegate
	parent *taskListModel
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

	priorityValueStyle := lipgloss.NewStyle().
		Foreground(colors.BadgeText()).
		Padding(0, 1)

	switch taskItem.Priority {
	case "low":
		titleStyle = titleStyle.BorderForeground(colors.Indigo())
		labelsStyle = labelsStyle.BorderForeground(colors.Indigo())
		priorityValueStyle = priorityValueStyle.
			BorderForeground(colors.Indigo()).Background(colors.Indigo())
	case "medium":
		titleStyle = titleStyle.BorderForeground(colors.Orange())
		labelsStyle = labelsStyle.BorderForeground(colors.Orange())
		priorityValueStyle = priorityValueStyle.
			BorderForeground(colors.Orange()).Background(colors.Orange())
	case "high":
		titleStyle = titleStyle.BorderForeground(colors.Red())
		labelsStyle = labelsStyle.BorderForeground(colors.Red())
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
	} else {
		titleStyle = titleStyle.MarginLeft(1)
		labelsStyle = labelsStyle.MarginLeft(1)
	}

	// Check if item is selected
	_, selected := d.parent.selected[index]

	marker := ""
	if selected {
		marker = lipgloss.NewStyle().
			Foreground(colors.Red()).
			Render("➜ ")
	}

	left := titleStyle.Render(marker+taskItem.CropTaskTitle(taskEntryLength)) + "\n" +
		labelsStyle.Render(taskItem.CropTaskLabels(taskEntryLength))

	right := priorityValueStyle.Render(taskItem.Priority)

	now := time.Now()
	dueDate := taskItem.DueDate

	if dueDate != nil &&
		items.IsToday(dueDate) &&
		dueDate.After(now) {
		right += lipgloss.NewStyle().
			Padding(0, 1).
			Background(colors.VividRed()).
			Foreground(colors.BadgeText()).
			Render("due today")
	}

	if dueDate != nil && dueDate.Before(now) {
		right += lipgloss.NewStyle().
			Padding(0, 1).
			Background(colors.VividRed()).
			Foreground(colors.BadgeText()).
			Render("overdue")
	}

	if taskItem.InProgress {
		right += lipgloss.NewStyle().
			Padding(0, 1).
			Background(colors.Blue()).
			Foreground(colors.BadgeText()).
			Render("in progress")
	}

	if dueDate != nil &&
		!dueDate.Before(now) &&
		!items.IsToday(dueDate) {
		right += lipgloss.NewStyle().
			Padding(0, 1).
			Background(colors.Yellow()).
			Foreground(colors.BadgeText()).
			Render("due in " + taskItem.DaysUntilToString() + " day(s)")
	}

	if taskItem.Completed {
		right = lipgloss.NewStyle().
			Padding(0, 1).
			Background(colors.Green()).
			Foreground(colors.BadgeText()).
			Render("done")
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Render(left),
		right,
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
	selected         map[int]struct{}
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
		project:      project,
		projectModel: projectModel,
		keys:         listKeys,
		selected:     make(map[int]struct{}),
		progress:     progress.New(progress.WithGradient("#FFA336", "#02BF87")),
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
			m.status = "🗸  Task started"

		case "stop":
			m.status = "🗸  Task stopped"

		case "complete":
			m.status = "🗸  Task completed"

		case "reopen":
			m.status = "🗸  Task reopened"

		default:
			return m, nil
		}
		return m, m.progress.SetPercent(0.5)

	case items.WriteTaskJSONErrorMsg:
		m.mode = 2
		m.err = msg.Err
		return m, m.progress.SetPercent(0.0)

	case items.TaskDeleteDoneMsg:
		selected := m.list.SelectedItem()
		if selected != nil {
			m.list.RemoveItem(m.list.Index())
		}
		m.status = "🗑  Task deleted"
		return m, m.progress.SetPercent(0.5)

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
				if m.list.SelectedItem() != nil {
					cmds = append(cmds,
						m.progress.SetPercent(0.10),
						tickCmd(),
						m.list.SelectedItem().(*items.Task).DeleteTaskFromFS(*m.project),
						vcs.CommitCmd(filepath.Join(m.project.ID, m.list.SelectedItem().(*items.Task).ID+".json"),
							"delete: "+m.list.SelectedItem().(*items.Task).Title),
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

			case key.Matches(msg, m.keys.chooseItem):
				if m.list.SelectedItem() != nil {
					markdown := m.list.SelectedItem().(*items.Task).TaskToMarkdown()
					pagerModel := newTaskPagerModel(markdown, &m)

					return pagerModel, tea.WindowSize()
				}
				return m, nil

			case key.Matches(msg, m.keys.toggleInProgress):
				if m.list.SelectedItem() != nil {
					t := m.list.SelectedItem().(*items.Task)

					if t.Completed {
						return m, m.list.NewStatusMessage(textStyleRed("Cannot set done task in progress"))
					}

					t.InProgress = !t.InProgress
					json := t.MarshalTask()

					cmds = append(cmds, tickCmd(), m.progress.SetPercent(0.10))
					if t.InProgress {
						cmds = append(cmds,
							t.WriteTaskJSON(json, *m.project, "start"),
							vcs.CommitCmd(filepath.Join(m.project.ID, t.ID+".json"), "starting progress: "+t.Title),
						)
						m.status = ""
						return m, tea.Batch(cmds...)
					}

					cmds = append(cmds,
						t.WriteTaskJSON(json, *m.project, "stop"),
						vcs.CommitCmd(filepath.Join(m.project.ID, t.ID+".json"), "stopping progress: "+t.Title),
					)
					m.status = ""
					return m, tea.Batch(cmds...)
				}
				return m, nil

			case key.Matches(msg, m.keys.toggleComplete):
				if m.list.SelectedItem() != nil {
					t := m.list.SelectedItem().(*items.Task)
					t.InProgress = false
					t.Completed = !t.Completed

					json := t.MarshalTask()

					cmds = append(cmds, tickCmd(), m.progress.SetPercent(0.10))
					if t.Completed {
						cmds = append(cmds,
							t.WriteTaskJSON(json, *m.project, "complete"),
							vcs.CommitCmd(filepath.Join(m.project.ID, t.ID+".json"), "complete: "+t.Title),
						)
						m.status = ""
						return m, tea.Batch(cmds...)
					}

					cmds = append(cmds,
						t.WriteTaskJSON(json, *m.project, "reopen"),
						vcs.CommitCmd(filepath.Join(m.project.ID, t.ID+".json"), "reopen: "+t.Title),
					)
					m.status = ""
					return m, tea.Batch(cmds...)
				}
				return m, nil

			case key.Matches(msg, m.keys.deleteItem):
				if m.list.SelectedItem() != nil {
					m.mode = modeConfirmDelete
				}
				return m, nil

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
					i := m.list.GlobalIndex()

					if _, ok := m.selected[i]; ok {
						delete(m.selected, i)
					} else {
						m.selected[i] = struct{}{}
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
			Render(textStyleGreen(m.status) + "\n\n" + m.progress.ViewAs(1.0))
	}

	// Display progress bar if not at 0%
	if m.progress.Percent() != 0.0 {
		return centeredStyle.Bold(true).
			Render(textStyleGreen(m.status) + "\n\n" + m.progress.View())
	}

	// Display deletion confirm view.
	if m.mode == modeConfirmDelete {
		selected := m.list.SelectedItem().(*items.Task)

		return centeredStyle.Render(
			fmt.Sprintf("Delete \"%s\"?\n\n", selected.Title) +
				textStyleRed("[y] Yes") + "    " + textStyleGreen("[n] No"),
		)
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

	slices.SortStableFunc(tasks, func(x, y *items.Task) int {
		for _, k := range keys {
			switch k {
			case "priority":
				// Completed tasks to bottom
				if x.Completed && !y.Completed {
					return 1
				}
				if !x.Completed && y.Completed {
					return -1
				}
				// Higher number = higher priority
				if compare := cmp.Compare(y.PriorityValue(), x.PriorityValue()); compare != 0 {
					return compare
				}

			case "dueDate":
				dx, dy := x.DueDate, y.DueDate
				switch {
				case dx == nil && dy != nil:
					return 1
				case dx != nil && dy == nil:
					return -1
				case dx != nil && dy != nil:
					if dx.Before(*dy) {
						return -1
					}
					if dx.After(*dy) {
						return 1
					}
				}

			case "state":
				// Completed tasks go to bottom
				if x.Completed && !y.Completed {
					return 1
				}
				if !x.Completed && y.Completed {
					return -1
				}
				// In-progress before others
				if x.InProgress && !y.InProgress {
					return -1
				}
				if !x.InProgress && y.InProgress {
					return 1
				}
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
