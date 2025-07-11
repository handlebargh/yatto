package models

import (
	"fmt"
	"io"
	"sort"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
	"github.com/handlebargh/yatto/internal/git"
	"github.com/handlebargh/yatto/internal/items"
	"github.com/spf13/viper"
)

type taskListKeyMap struct {
	toggleHelpMenu key.Binding
	addItem        key.Binding
	chooseItem     key.Binding
	editItem       key.Binding
	deleteItem     key.Binding
	sortByPriority key.Binding
	toggleComplete key.Binding
	showBranchView key.Binding
}

func newTaskListKeyMap() *taskListKeyMap {
	return &taskListKeyMap{
		showBranchView: key.NewBinding(
			key.WithKeys("b"),
			key.WithHelp("b", "branch view"),
		),
		toggleComplete: key.NewBinding(
			key.WithKeys("D"),
			key.WithHelp("D", "toggle done"),
		),
		sortByPriority: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "sort by priority"),
		),
		deleteItem: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete"),
		),
		editItem: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit"),
		),
		chooseItem: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "choose"),
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
		Foreground(neutral).
		BorderForeground(orange).
		Width(64)

	completedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#000000")).
		BorderForeground(orange).
		Padding(0, 1)

	priorityStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#000000")).
		Padding(0, 1)

	// Priority-based coloring.
	switch taskItem.Priority() {
	case "low":
		priorityStyle = priorityStyle.Background(indigo)
	case "medium":
		priorityStyle = priorityStyle.Background(orange)
	case "high":
		priorityStyle = priorityStyle.Background(red)
	}

	// Completed coloring.
	if taskItem.Completed() {
		completedStyle = completedStyle.Background(green)
		titleStyle = titleStyle.Strikethrough(true)
	} else {
		completedStyle = completedStyle.Background(blue)
	}

	// Selection border on the left only.
	if index == m.GlobalIndex() {
		titleStyle = titleStyle.
			Border(lipgloss.ThickBorder(), false, false, false, true).
			MarginLeft(0)
		completedStyle = completedStyle.
			Border(lipgloss.ThickBorder(), false, false, false, true).
			MarginLeft(0)
	} else {
		titleStyle = titleStyle.MarginLeft(1)
		completedStyle = completedStyle.MarginLeft(1)
	}

	line := titleStyle.Render(taskItem.Title()) + "\n" +
		completedStyle.Render(items.CompletedString(taskItem.Completed())) + " " +
		priorityStyle.Render(taskItem.Priority())

	_, err := fmt.Fprint(w, line)
	if err != nil {
		panic(err)
	}
}

type taskListModel struct {
	list      list.Model
	selected  bool
	selection *items.Task
	keys      *taskListKeyMap
	mode      mode
	err       error
	spinner   spinner.Model
	loading   bool
	width     int
	height    int
}

func InitialTaskListModel() taskListModel {
	listKeys := newTaskListKeyMap()

	tasks := items.ReadTasksFromFS()
	items := []list.Item{}

	for _, task := range tasks {
		items = append(items, &task)
	}

	itemList := list.New(items, customTaskDelegate{DefaultDelegate: list.NewDefaultDelegate()}, 0, 0)
	itemList.SetShowPagination(true)
	itemList.SetShowTitle(true)
	itemList.SetShowStatusBar(true)
	itemList.Title = "YATTO"
	itemList.Styles.Title = titleStyle
	itemList.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			listKeys.toggleHelpMenu,
			listKeys.addItem,
			listKeys.chooseItem,
			listKeys.editItem,
			listKeys.deleteItem,
			listKeys.sortByPriority,
			listKeys.toggleComplete,
		}
	}

	return taskListModel{
		list:     itemList,
		selected: false,
		keys:     listKeys,
		loading:  false,
		spinner: spinner.New(
			spinner.WithSpinner(spinner.Pulse),
			spinner.WithStyle(lipgloss.NewStyle().Foreground(orange)),
		),
	}
}

func (m taskListModel) Init() tea.Cmd {
	if viper.GetBool("git.enable") {
		return tea.Batch(
			m.spinner.Tick,
			git.InitCmd(),
		)
	}

	return nil
}

func (m taskListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case git.GitInitDoneMsg:
		return m, m.list.NewStatusMessage(statusMessageStyleGreen("🕹  Initialization completed"))

	case git.GitInitErrorMsg:
		m.loading = false
		m.mode = 2
		m.err = msg.Err
		return m, nil

	case git.GitCommitDoneMsg:
		m.loading = false
		return m, m.list.NewStatusMessage(statusMessageStyleGreen("🗘  Changes committed"))

	case git.GitCommitErrorMsg:
		m.loading = false
		m.mode = 2
		m.err = msg.Err
		return m, nil

	case items.WriteJSONDoneMsg:
		m.loading = false
		switch msg.Kind {
		case "create":
			m.list.InsertItem(0, &msg.Task)
			return m, m.list.NewStatusMessage(statusMessageStyleGreen("🗸  Task created"))

		case "update":
			return m, m.list.NewStatusMessage(statusMessageStyleGreen("🗸  Task updated"))

		case "complete":
			if msg.Task.Completed() {
				return m, m.list.NewStatusMessage(statusMessageStyleGreen("🗸  Task completed"))
			}
			return m, m.list.NewStatusMessage(statusMessageStyleGreen("🗸  Task reopened"))

		default:
			return m, nil
		}

	case items.WriteJSONErrorMsg:
		m.loading = false
		m.mode = 2
		m.err = msg.Err
		return m, nil

	case items.TaskDeleteDoneMsg:
		m.loading = false
		m.list.RemoveItem(m.list.GlobalIndex())
		return m, m.list.NewStatusMessage(statusMessageStyleGreen("🗑  Task deleted"))

	case items.TaskDeleteErrorMsg:
		m.loading = false
		m.mode = 2
		m.err = msg.Err
		return m, nil

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
					m.loading = true
					cmds = append(cmds, items.DeleteTaskFromFS(m.list.SelectedItem().(*items.Task)),
						git.CommitCmd(m.list.SelectedItem().(*items.Task).Id(),
							"delete: "+m.list.SelectedItem().(*items.Task).Title()),
					)
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

				return m, tea.Quit
			}

			switch {
			case key.Matches(msg, m.keys.toggleHelpMenu):
				m.list.SetShowHelp(!m.list.ShowHelp())
				return m, nil

			case key.Matches(msg, m.keys.showBranchView):
				branchListModel := InitialBranchListModel(&m)
				return branchListModel, nil

			case key.Matches(msg, m.keys.sortByPriority):
				sortTasksByPriority(&m.list)
				return m, nil

			case key.Matches(msg, m.keys.chooseItem):
				if m.list.SelectedItem() != nil {
					m.selected = true
					m.selection = m.list.SelectedItem().(*items.Task)
				}
				return m, nil

			case key.Matches(msg, m.keys.toggleComplete):
				if m.list.SelectedItem() != nil {
					t := m.list.SelectedItem().(*items.Task)
					json := items.MarshalTask(
						t.Id(),
						t.Title(),
						t.Description(),
						t.Priority(),
						!t.Completed())

					m.list.SelectedItem().(*items.Task).TaskCompleted = !m.list.SelectedItem().(*items.Task).TaskCompleted
					cmds = append(cmds, items.WriteJson(json, *t, "complete"),
						git.CommitCmd(t.Id(),
							"complete: "+t.Title()),
					)
					m.loading = true
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
					// Switch to InputModel for editing.
					formModel := newTaskFormModel(m.list.SelectedItem().(*items.Task), &m, true)
					return formModel, nil
				}

			case key.Matches(msg, m.keys.addItem):
				task := &items.Task{
					TaskId:          uuid.NewString(),
					TaskTitle:       "",
					TaskDescription: "",
				}
				formModel := newTaskFormModel(task, &m, false)
				return formModel, nil
			}
		}
	}

	newListModel, cmd := m.list.Update(msg)
	m.list = newListModel
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m taskListModel) View() string {
	// Display spinner while git operation is running.
	if m.loading {
		spinnerStyle := lipgloss.NewStyle().
			Width(m.width).
			Height(m.height).
			Align(lipgloss.Center).
			AlignVertical(lipgloss.Center)

		return spinnerStyle.Render(fmt.Sprintf("\n%s %s\n   %s", m.spinner.View(),
			lipgloss.NewStyle().Foreground(orange).Render("Synchronization in progress"),
			lipgloss.NewStyle().Foreground(red).Render("Do not exit application!")))
	}

	// Display deletion confirm view.
	if m.mode == modeConfirmDelete {
		selected := m.list.SelectedItem().(*items.Task)

		boxContent := fmt.Sprintf("Delete \"%s\"?\n\n[y] Yes   [n] No", selected.Title())

		leftColumn := appStyle.Render(m.list.View())
		rightColumn := promptBoxStyle.Render(boxContent)

		return lipgloss.JoinHorizontal(lipgloss.Top, leftColumn, rightColumn)
	}

	// Display git error view
	if m.mode == modeGitError {
		boxContent := "An error occured while executing git:\n\n" +
			m.err.Error() + "\n\n" +
			"Please commit manually!"

		leftColumn := appStyle.Render(m.list.View())
		rightColumn := promptBoxStyle.Render(boxContent)

		return lipgloss.JoinHorizontal(lipgloss.Top, leftColumn, rightColumn)
	}

	// Display list view.
	if !m.selected {
		return appStyle.Render(m.list.View())
	}

	// Display task view.
	completed := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#000000")).
		Padding(0, 1)

	switch m.selection.Completed() {
	case true:
		completed = completed.Background(green)
	case false:
		completed = completed.Background(blue)
	}

	headline := lipgloss.NewStyle().
		Foreground(green).
		Bold(true)

	priority := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#000000")).
		Padding(0, 1)

	switch m.selection.Priority() {
	case "high":
		priority = priority.Background(red)
	case "medium":
		priority = priority.Background(orange)
	case "low":
		priority = priority.Background(indigo)
	default:
		priority = priority.Background(indigo)
	}

	leftColumn := appStyle.Render(m.list.View())
	rightColumn := detailBoxStyle.Render(
		completed.Render(
			items.CompletedString(m.selection.Completed()),
		) + " " + priority.Render(
			m.selection.Priority(),
		) + "\n\n" +
			headline.Render(
				"Title",
			) + "\n\n" + m.selection.Title() + "\n\n" +
			headline.Render(
				"Description:",
			) + "\n\n" + m.selection.TaskDescription,
	)
	return lipgloss.JoinHorizontal(lipgloss.Top, leftColumn, rightColumn)
}

func sortTasksByPriority(m *list.Model) {
	// Preserve selected item
	selected := m.SelectedItem()

	// Extract all tasks
	listItems := m.Items()
	tasks := make([]*items.Task, len(listItems))
	for i, item := range listItems {
		tasks[i] = item.(*items.Task)
	}

	// Sort tasks by priority
	sort.Slice(tasks, func(i, j int) bool {
		return items.PriorityValue(tasks[i].Priority()) >
			items.PriorityValue(tasks[j].Priority())
	})

	// Convert back to []list.Item
	sortedItems := make([]list.Item, len(tasks))
	for i, t := range tasks {
		sortedItems[i] = t
	}
	m.SetItems(sortedItems)

	// Re-select the same item
	if selected != nil {
		for i, item := range sortedItems {
			if item == selected {
				m.Select(i)
				break
			}
		}
	}
}
