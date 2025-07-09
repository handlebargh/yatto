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
	"github.com/handlebargh/yatto/internal/task"
	"github.com/spf13/viper"
)

var (
	red    = lipgloss.AdaptiveColor{Light: "#FE5F86", Dark: "#FE5F86"}
	indigo = lipgloss.AdaptiveColor{Light: "#5A56E0", Dark: "#7571F9"}
	green  = lipgloss.AdaptiveColor{Light: "#02BA84", Dark: "#02BF87"}
	orange = lipgloss.AdaptiveColor{Light: "#FFB733", Dark: "#FFA336"}
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

	statusMessageStyle = lipgloss.NewStyle().
				Foreground(green).
				Render
)

type mode int

const (
	modeNormal mode = iota
	modeConfirmDelete
	modeGitError
)

type listKeyMap struct {
	toggleHelpMenu key.Binding
	insertItem     key.Binding
	chooseItem     key.Binding
	editItem       key.Binding
	deleteItem     key.Binding
	sortByPriority key.Binding
}

func newListKeyMap() *listKeyMap {
	return &listKeyMap{
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
		insertItem: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "add item"),
		),
		toggleHelpMenu: key.NewBinding(
			key.WithKeys("H"),
			key.WithHelp("H", "toggle help"),
		),
	}
}

type customDelegate struct {
	list.DefaultDelegate
}

func (d customDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	taskItem, ok := item.(*task.Task)
	if !ok {
		_, err := fmt.Fprint(w, "Invalid item")
		if err != nil {
			panic(err)
		}

		return
	}

	var style lipgloss.Style
	switch taskItem.Priority() {
	case "low":
		style = lipgloss.NewStyle().Foreground(indigo).BorderForeground(indigo)
	case "medium":
		style = lipgloss.NewStyle().Foreground(orange).BorderForeground(orange)
	case "high":
		style = lipgloss.NewStyle().Foreground(red).BorderForeground(red)
	}

	if taskItem.Completed() {
		style = style.Strikethrough(true).Foreground(green).BorderForeground(green)
	}

	if index == m.GlobalIndex() {
		style = style.Border(lipgloss.NormalBorder(), false, false, false, true)
	} else {
		style = style.PaddingLeft(1)
	}

	_, err := fmt.Fprint(w, style.Render(" "+taskItem.Title()))
	if err != nil {
		panic(err)
	}
}

type listModel struct {
	list      list.Model
	selected  bool
	selection *task.Task
	keys      *listKeyMap
	mode      mode
	err       error
	spinner   spinner.Model
	loading   bool
}

func InitialListModel() listModel {
	listKeys := newListKeyMap()

	tasks := task.ReadTasksFromFS()
	items := []list.Item{}

	for _, task := range tasks {
		items = append(items, &task)
	}

	itemList := list.New(items, customDelegate{DefaultDelegate: list.NewDefaultDelegate()}, 0, 0)
	itemList.SetShowPagination(true)
	itemList.SetShowTitle(true)
	itemList.SetShowStatusBar(true)
	itemList.Title = "YATTO"
	itemList.Styles.Title = titleStyle
	itemList.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			listKeys.toggleHelpMenu,
			listKeys.insertItem,
			listKeys.chooseItem,
			listKeys.editItem,
			listKeys.deleteItem,
			listKeys.sortByPriority,
		}
	}

	return listModel{
		list:     itemList,
		selected: false,
		keys:     listKeys,
		spinner: spinner.New(
			spinner.WithSpinner(spinner.Dot),
			spinner.WithStyle(lipgloss.NewStyle().Foreground(orange)),
		),
	}
}

func (m listModel) Init() tea.Cmd {
	if viper.GetBool("use_git") {
		storageDir := viper.GetString("storage_dir")
		return tea.Batch(
			m.spinner.Tick,

			git.GitInit(storageDir),
		)
	}

	return nil
}

func (m listModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case git.GitErrorMsg:
		m.loading = false
		m.mode = 2
		m.err = msg.Err
		return m, nil

	case git.GitDoneMsg:
		m.loading = false
		return m, m.list.NewStatusMessage(statusMessageStyle("✅ Tasks synchronized"))

	case task.JsonWriteDoneMsg:
		m.loading = false
		return m, m.list.NewStatusMessage(statusMessageStyle("✅ Task created/updated"))

	case task.JsonWriteErrorMsg:
		m.loading = false
		m.mode = 2
		m.err = msg.Err
		return m, nil

	case task.TaskDeleteDoneMsg:
		m.loading = false
		m.list.RemoveItem(m.list.GlobalIndex())
		return m, m.list.NewStatusMessage(statusMessageStyle("✅ Task deleted"))

	case task.TaskDeleteErrorMsg:
		m.loading = false
		m.mode = 2
		m.err = msg.Err
		return m, nil

	case tea.WindowSizeMsg:
		h, v := appStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)

	case tea.KeyMsg:
		switch m.mode {
		case modeConfirmDelete:
			switch msg.String() {
			case "y", "Y":
				var cmd tea.Cmd
				if m.list.SelectedItem() != nil {
					m.loading = true
					cmd = task.DeleteTaskFromFSCmd(m.list.SelectedItem().(*task.Task),
						"delete: "+m.list.SelectedItem().(*task.Task).Title())
				}
				m.mode = modeNormal
				return m, cmd

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

			case key.Matches(msg, m.keys.sortByPriority):
				sortTasksByPriority(&m.list)
				return m, nil

			case key.Matches(msg, m.keys.chooseItem):
				if m.list.SelectedItem() != nil {
					m.selected = true
					m.selection = m.list.SelectedItem().(*task.Task)
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
					formModel := newFormModel(m.list.SelectedItem().(*task.Task), &m, true)
					return formModel, nil
				}

			case key.Matches(msg, m.keys.insertItem):
				task := &task.Task{
					TaskId:          uuid.NewString(),
					TaskTitle:       "",
					TaskDescription: "",
				}
				formModel := newFormModel(task, &m, false)
				return formModel, nil
			}
		}
	}

	newListModel, cmd := m.list.Update(msg)
	m.list = newListModel
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m listModel) View() string {
	// Display spinner while git operation is running.
	if m.loading {
		leftColumn := appStyle.Render(m.list.View())

		rightColumn := fmt.Sprintf("\n%s %s\n   %s", m.spinner.View(),
			lipgloss.NewStyle().Foreground(orange).Render("Synchronization in progress"),
			lipgloss.NewStyle().Foreground(red).Render("Do not exit application!"))

		return lipgloss.JoinHorizontal(lipgloss.Top, leftColumn, rightColumn)
	}

	// Display deletion confirm view.
	if m.mode == modeConfirmDelete {
		selected := m.list.SelectedItem().(*task.Task)

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
		completed = completed.Background(red)
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
			task.CompletedString(m.selection.Completed()),
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
	items := m.Items()
	tasks := make([]*task.Task, len(items))
	for i, item := range items {
		tasks[i] = item.(*task.Task)
	}

	// Sort tasks by priority
	sort.Slice(tasks, func(i, j int) bool {
		return task.PriorityValue(tasks[i].Priority()) >
			task.PriorityValue(tasks[j].Priority())
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
