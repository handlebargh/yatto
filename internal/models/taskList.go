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
	"cmp"
	"fmt"
	"io"
	"path/filepath"
	"slices"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
	"github.com/handlebargh/yatto/internal/colors"
	"github.com/handlebargh/yatto/internal/git"
	"github.com/handlebargh/yatto/internal/helpers"
	"github.com/handlebargh/yatto/internal/items"
)

const taskEntryLength = 53

// taskListKeyMap defines the key bindings used in the task list view.
type taskListKeyMap struct {
	toggleHelpMenu   key.Binding
	addItem          key.Binding
	chooseItem       key.Binding
	editItem         key.Binding
	deleteItem       key.Binding
	sortByPriority   key.Binding
	sortByDueDate    key.Binding
	toggleInProgress key.Binding
	toggleComplete   key.Binding
}

// newTaskListKeyMap initializes and returns a new key map for task list actions.
func newTaskListKeyMap() *taskListKeyMap {
	return &taskListKeyMap{
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
			key.WithKeys("enter"),
			key.WithHelp("enter", "show"),
		),
		addItem: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "add item"),
		),
		toggleHelpMenu: key.NewBinding(
			key.WithKeys("H"),
			key.WithHelp("H", "toggle help"),
		),
	}
}

// customTaskDelegate is a custom list delegate for rendering task items.
type customTaskDelegate struct {
	list.DefaultDelegate
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
		Foreground(colors.Blue).
		Padding(0, 1)

	priorityValueStyle := lipgloss.NewStyle().
		Foreground(colors.Black).
		Padding(0, 1)

	switch taskItem.Priority() {
	case "low":
		titleStyle = titleStyle.BorderForeground(colors.Indigo)
		labelsStyle = labelsStyle.BorderForeground(colors.Indigo)
		priorityValueStyle = priorityValueStyle.
			BorderForeground(colors.Indigo).Background(colors.Indigo)
	case "medium":
		titleStyle = titleStyle.BorderForeground(colors.Orange)
		labelsStyle = labelsStyle.BorderForeground(colors.Orange)
		priorityValueStyle = priorityValueStyle.
			BorderForeground(colors.Orange).Background(colors.Orange)
	case "high":
		titleStyle = titleStyle.BorderForeground(colors.Red)
		labelsStyle = labelsStyle.BorderForeground(colors.Red)
		priorityValueStyle = priorityValueStyle.
			BorderForeground(colors.Red).Background(colors.Red)
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

	left := titleStyle.Render(taskItem.CropTaskTitle(taskEntryLength)) + "\n" +
		labelsStyle.Render(taskItem.CropTaskLabels(taskEntryLength))

	right := priorityValueStyle.Render(taskItem.Priority())

	now := time.Now()
	dueDate := taskItem.DueDate()

	if dueDate != nil &&
		items.IsToday(dueDate) &&
		dueDate.After(now) {
		right = right + lipgloss.NewStyle().
			Padding(0, 1).
			Background(colors.VividRed).
			Foreground(colors.Black).
			Render("due today")
	}

	if dueDate != nil && dueDate.Before(now) {
		right = right + lipgloss.NewStyle().
			Padding(0, 1).
			Background(colors.VividRed).
			Foreground(colors.Black).
			Render("overdue")
	}

	if taskItem.InProgress() {
		right = right + lipgloss.NewStyle().
			Padding(0, 1).
			Background(colors.Blue).
			Foreground(colors.Black).
			Render("in progress")
	}

	if dueDate != nil &&
		!dueDate.Before(now) &&
		!items.IsToday(dueDate) {
		right = right + lipgloss.NewStyle().
			Padding(0, 1).
			Background(colors.Yellow).
			Foreground(colors.Black).
			Render("due in "+taskItem.DaysUntilToString()+" days")
	}

	if taskItem.Completed() {
		right = lipgloss.NewStyle().
			Padding(0, 1).
			Background(colors.Green).
			Foreground(colors.Black).
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
	projectModel     *projectListModel
	keys             *taskListKeyMap
	mode             mode
	err              error
	progress         progress.Model
	progressDone     bool
	waitingAfterDone bool
	status           string
	width, height    int
}

// newTaskListModel creates a new taskListModel for the given project.
func newTaskListModel(project *items.Project, projectModel *projectListModel) taskListModel {
	listKeys := newTaskListKeyMap()

	tasks := project.ReadTasksFromFS()
	items := []list.Item{}

	for _, task := range tasks {
		items = append(items, &task)
	}

	color := helpers.GetColorCode(project.Color())

	titleStyleTasks := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#000000")).
		Background(color).
		Padding(0, 1)

	itemList := list.New(items, customTaskDelegate{DefaultDelegate: list.NewDefaultDelegate()}, 0, 0)
	itemList.SetShowPagination(true)
	itemList.SetShowTitle(true)
	itemList.SetShowStatusBar(true)
	itemList.SetStatusBarItemName("task", "tasks")
	itemList.Title = project.Title()
	itemList.Styles.Title = titleStyleTasks
	itemList.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			listKeys.toggleHelpMenu,
			listKeys.addItem,
			listKeys.chooseItem,
			listKeys.editItem,
			listKeys.deleteItem,
			listKeys.sortByPriority,
			listKeys.sortByDueDate,
			listKeys.toggleInProgress,
			listKeys.toggleComplete,
		}
	}

	return taskListModel{
		list:         itemList,
		project:      project,
		projectModel: projectModel,
		keys:         listKeys,
		progress:     progress.New(progress.WithGradient("#FFA336", "#02BF87")),
	}
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

	case git.GitCommitDoneMsg:
		m.status = "🗘  Changes committed"
		m.progressDone = true
		return m, m.progress.SetPercent(1.0)

	case git.GitCommitErrorMsg:
		m.mode = 2
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
						git.CommitCmd(filepath.Join(m.project.Id(), m.list.SelectedItem().(*items.Task).Id()+".json"),
							"delete: "+m.list.SelectedItem().(*items.Task).Title()),
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

			switch msg.String() {
			case "esc", "q":
				return m.projectModel, nil
			}

			switch {
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

					if t.Completed() {
						return m, m.list.NewStatusMessage(textStyleRed("Cannot set done task in progress"))
					}

					t.SetInProgress(!t.InProgress())
					json := t.MarshalTask()

					cmds = append(cmds, tickCmd(), m.progress.SetPercent(0.10))
					if t.InProgress() {
						cmds = append(cmds,
							t.WriteTaskJson(json, *m.project, "start"),
							git.CommitCmd(filepath.Join(m.project.Id(), t.Id()+".json"), "starting progress: "+t.Title()),
						)
						m.status = ""
						return m, tea.Batch(cmds...)
					}

					cmds = append(cmds,
						t.WriteTaskJson(json, *m.project, "stop"),
						git.CommitCmd(filepath.Join(m.project.Id(), t.Id()+".json"), "stopping progress: "+t.Title()),
					)
					m.status = ""
					return m, tea.Batch(cmds...)
				}
				return m, nil

			case key.Matches(msg, m.keys.toggleComplete):
				if m.list.SelectedItem() != nil {
					t := m.list.SelectedItem().(*items.Task)
					t.SetInProgress(false)
					t.SetCompleted(!t.Completed())

					json := t.MarshalTask()

					cmds = append(cmds, tickCmd(), m.progress.SetPercent(0.10))
					if t.Completed() {
						cmds = append(cmds,
							t.WriteTaskJson(json, *m.project, "complete"),
							git.CommitCmd(filepath.Join(m.project.Id(), t.Id()+".json"), "complete: "+t.Title()),
						)
						m.status = ""
						return m, tea.Batch(cmds...)
					}

					cmds = append(cmds,
						t.WriteTaskJson(json, *m.project, "reopen"),
						git.CommitCmd(filepath.Join(m.project.Id(), t.Id()+".json"), "reopen: "+t.Title()),
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

			case key.Matches(msg, m.keys.addItem):
				task := &items.Task{
					TaskId:          uuid.NewString(),
					TaskTitle:       "",
					TaskDescription: "",
				}
				formModel := newTaskFormModel(task, &m, false)
				return formModel, tea.WindowSize()
			}
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
		return centeredStyle.Bold(true).Render(textStyleGreen(m.status) + "\n\n" + m.progress.ViewAs(1.0))
	}

	// Display progress bar if not at 0%
	if m.progress.Percent() != 0.0 {
		return centeredStyle.Bold(true).Render(textStyleGreen(m.status) + "\n\n" + m.progress.View())
	}

	// Display deletion confirm view.
	if m.mode == modeConfirmDelete {
		selected := m.list.SelectedItem().(*items.Task)

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
		for _, key := range keys {
			switch key {
			case "priority":
				// Completed tasks to bottom
				if x.Completed() && !y.Completed() {
					return 1
				}
				if !x.Completed() && y.Completed() {
					return -1
				}
				// Higher number = higher priority
				if cmp := cmp.Compare(y.PriorityValue(), x.PriorityValue()); cmp != 0 {
					return cmp
				}

			case "dueDate":
				dx, dy := x.DueDate(), y.DueDate()
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
				if x.Completed() && !y.Completed() {
					return 1
				}
				if !x.Completed() && y.Completed() {
					return -1
				}
				// In-progress before others
				if x.InProgress() && !y.InProgress() {
					return -1
				}
				if !x.InProgress() && y.InProgress() {
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
			if task, ok := item.(*items.Task); ok && task.Id() == selectedTask.Id() {
				m.Select(i)
				break
			}
		}
	}
}
