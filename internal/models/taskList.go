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
	"path/filepath"
	"sort"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
	"github.com/handlebargh/yatto/internal/git"
	"github.com/handlebargh/yatto/internal/items"
)

type taskListKeyMap struct {
	toggleHelpMenu key.Binding
	addItem        key.Binding
	chooseItem     key.Binding
	editItem       key.Binding
	deleteItem     key.Binding
	sortByPriority key.Binding
	sortByDueDate  key.Binding
	sortByState    key.Binding
	toggleComplete key.Binding
}

func newTaskListKeyMap() *taskListKeyMap {
	return &taskListKeyMap{
		toggleComplete: key.NewBinding(
			key.WithKeys("C"),
			key.WithHelp("C", "toggle complete"),
		),
		sortByPriority: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "sort by priority"),
		),
		sortByDueDate: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "sort by due date"),
		),
		sortByState: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "sort by state"),
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

type customTaskDelegate struct {
	list.DefaultDelegate
}

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
		Foreground(blue).
		Padding(0, 1)

	priorityValueStyle := lipgloss.NewStyle().
		Foreground(black).
		Padding(0, 1)

	switch taskItem.Priority() {
	case "low":
		titleStyle = titleStyle.BorderForeground(indigo)
		labelsStyle = labelsStyle.BorderForeground(indigo)
		priorityValueStyle = priorityValueStyle.
			BorderForeground(indigo).Background(indigo)
	case "medium":
		titleStyle = titleStyle.BorderForeground(orange)
		labelsStyle = labelsStyle.BorderForeground(orange)
		priorityValueStyle = priorityValueStyle.
			BorderForeground(orange).Background(orange)
	case "high":
		titleStyle = titleStyle.BorderForeground(red)
		labelsStyle = labelsStyle.BorderForeground(red)
		priorityValueStyle = priorityValueStyle.
			BorderForeground(red).Background(red)
	}

	if index == m.GlobalIndex() {
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

	left := titleStyle.Render(taskItem.Title()) + "\n" +
		labelsStyle.Render(taskItem.Labels())

	right := priorityValueStyle.Render(taskItem.Priority())

	if items.IsToday(taskItem.DueDate()) {
		right = lipgloss.NewStyle().
			Padding(0, 1).
			Background(red).
			Foreground(black).
			Render("Due today")
	}

	if taskItem.Completed() {
		right = lipgloss.NewStyle().
			Padding(0, 1).
			Background(green).
			Foreground(black).
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

type taskListModel struct {
	list             list.Model
	project          *items.Project
	projectModel     *projectListModel
	selected         bool
	selection        *items.Task
	keys             *taskListKeyMap
	mode             mode
	err              error
	progress         progress.Model
	progressDone     bool
	waitingAfterDone bool
	status           string
	width, height    int

	// Glamour renderer
	markdown string
	rendered string
}

func newTaskListModel(project *items.Project, projectModel *projectListModel) taskListModel {
	listKeys := newTaskListKeyMap()

	tasks := project.ReadTasksFromFS()
	items := []list.Item{}

	for _, task := range tasks {
		items = append(items, &task)
	}

	color := getColorCode(project.Color())

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
			listKeys.sortByState,
			listKeys.toggleComplete,
		}
	}

	return taskListModel{
		list:         itemList,
		project:      project,
		projectModel: projectModel,
		selected:     false,
		keys:         listKeys,
		progress:     progress.New(progress.WithGradient("#FFA336", "#02BF87")),
	}
}

func (m taskListModel) Init() tea.Cmd {
	return tickCmd()
}

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
			return m, m.progress.SetPercent(0.5)

		case "update":
			m.status = "🗸  Task updated"
			return m, m.progress.SetPercent(0.5)

		case "complete":
			m.status = "🗸  Task completed"
			return m, m.progress.SetPercent(0.5)

		case "reopen":
			m.status = "🗸  Task reopened"
			return m, m.progress.SetPercent(0.5)

		default:
			return m, nil
		}

	case items.WriteTaskJSONErrorMsg:
		m.mode = 2
		m.err = msg.Err
		return m, m.progress.SetPercent(0.0)

	case items.TaskDeleteDoneMsg:
		m.list.RemoveItem(m.list.GlobalIndex())
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
				if m.selected {
					m.selected = !m.selected
					return m, nil
				}

				return m.projectModel, nil
			}

			switch {
			case key.Matches(msg, m.keys.toggleHelpMenu):
				m.list.SetShowHelp(!m.list.ShowHelp())
				return m, nil

			case key.Matches(msg, m.keys.sortByPriority):
				sortTasksByKey(&m.list, "priority")
				return m, nil

			case key.Matches(msg, m.keys.sortByDueDate):
				sortTasksByKey(&m.list, "dueDate")
				return m, nil

			case key.Matches(msg, m.keys.sortByState):
				sortTasksByKey(&m.list, "state")
				return m, nil

			case key.Matches(msg, m.keys.chooseItem):
				if m.list.SelectedItem() != nil {
					var err error
					m.selected = true
					m.selection = m.list.SelectedItem().(*items.Task)
					m.markdown = m.selection.TaskToMarkdown()
					m.rendered, err = m.projectModel.renderer.Render(m.markdown)
					if err != nil {
						m.mode = 2
						m.err = err
						return m, nil
					}
				}
				return m, nil

			case key.Matches(msg, m.keys.toggleComplete):
				if m.list.SelectedItem() != nil {
					t := m.list.SelectedItem().(*items.Task)
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
	if !m.selected {
		return appStyle.Render(m.list.View())
	}

	return m.rendered
}

// Sorts the tasks list by key.
// Key may be either priority, dueDate or state.
func sortTasksByKey(m *list.Model, key string) {
	// Preserve selected item
	selected := m.SelectedItem()

	// Extract all tasks
	listItems := m.Items()
	tasks := make([]*items.Task, len(listItems))
	for i, item := range listItems {
		task, ok := item.(*items.Task)
		if !ok {
			continue
		}
		tasks[i] = task
	}

	switch key {
	case "priority":
		sort.Slice(tasks, func(i, j int) bool {
			return tasks[i].PriorityValue() > tasks[j].PriorityValue()
		})

	case "dueDate":
		sort.Slice(tasks, func(i, j int) bool {
			if tasks[i].DueDate() == nil {
				return false
			}

			if tasks[j].DueDate() == nil {
				return true
			}

			return tasks[i].DueDate().Before(*tasks[j].DueDate())
		})

	case "state":
		sort.Slice(tasks, func(i, j int) bool {
			return taskSortValue(tasks[i]) < taskSortValue(tasks[j])
		})

	default:
		// Do not sort at all.
	}

	// Convert back to []list.Item
	sortedItems := make([]list.Item, len(tasks))
	for i, t := range tasks {
		sortedItems[i] = t
	}
	m.SetItems(sortedItems)

	// Re-select the same item
	selectedTask, ok := selected.(*items.Task)
	if ok {
		for i, item := range sortedItems {
			task, ok := item.(*items.Task)
			if ok && task.Id() == selectedTask.Id() {
				m.Select(i)
				break
			}
		}
	}
}
