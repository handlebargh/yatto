package models

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/handlebargh/yatto/internal/git"
	"github.com/handlebargh/yatto/internal/items"

	tea "github.com/charmbracelet/bubbletea"
)

type branchListKeyMap struct {
	toggleHelpMenu key.Binding
	addBranch      key.Binding
	chooseBranch   key.Binding
	deleteBranch   key.Binding
}

func newBranchListKeyMap() *branchListKeyMap {
	return &branchListKeyMap{
		deleteBranch: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete branch"),
		),
		chooseBranch: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "choose branch"),
		),
		addBranch: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "add branch"),
		),
		toggleHelpMenu: key.NewBinding(
			key.WithKeys("H"),
			key.WithHelp("H", "toggle help"),
		),
	}
}

type branchListModel struct {
	list      list.Model
	listModel *taskListModel
	keys      *branchListKeyMap
	selected  bool
	selection *items.Branch
	mode      mode
	err       error
	spinner   spinner.Model
	loading   bool
}

func InitialBranchListModel(listModel *taskListModel) branchListModel {
	m := branchListModel{}

	listKeys := newBranchListKeyMap()

	branches, _, err := git.GetBranches()
	if err != nil {
		m.mode = 2
		m.err = err
	}

	listItems := []list.Item{}

	for _, branch := range branches {
		listItems = append(listItems, &branch)
	}

	itemList := list.New(listItems, list.NewDefaultDelegate(), 40, 40)
	itemList.SetShowPagination(true)
	itemList.SetShowTitle(true)
	itemList.SetShowStatusBar(true)
	itemList.Title = "Branches"
	itemList.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			listKeys.toggleHelpMenu,
			listKeys.addBranch,
			listKeys.chooseBranch,
			listKeys.deleteBranch,
		}
	}

	m.list = itemList
	m.listModel = listModel
	m.selected = false
	m.keys = listKeys
	m.spinner = spinner.New(
		spinner.WithSpinner(spinner.Dot),
		spinner.WithStyle(lipgloss.NewStyle().Foreground(orange)))

	return m
}

func (m branchListModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m branchListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case git.AddBranchErrorMsg:
		m.mode = 2
		m.err = msg.Err
		return m, nil

	case git.DeleteBranchErrorMsg:
		m.mode = 2
		m.err = msg.Err
		return m, nil

	case git.CheckoutBranchErrorMsg:
		m.mode = 2
		m.err = msg.Err
		return m, nil

	case git.AddBranchDoneMsg:
		m.loading = false
		m.list.InsertItem(0, msg.Branch)

	case git.DeleteBranchDoneMsg:
		m.loading = false
		m.list.RemoveItem(m.list.GlobalIndex())
		return m, m.list.NewStatusMessage(statusMessageStyleGreen("🗑  branch deleted"))

	case git.CheckoutBranchDoneMsg:

	case tea.WindowSizeMsg:
		h, v := appStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)

	case tea.KeyMsg:
		switch m.mode {
		case modeConfirmDelete:
			switch msg.String() {
			case "y", "Y":
				m.mode = modeNormal
				if m.list.SelectedItem() != nil {
					return m, git.DeleteBranchCmd(m.list.SelectedItem().(*items.Branch).Title())
				}
				return m, nil

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

				return m.listModel, nil
			}

			switch {
			case key.Matches(msg, m.keys.toggleHelpMenu):
				m.list.SetShowHelp(!m.list.ShowHelp())
				return m, nil

			case key.Matches(msg, m.keys.addBranch):
				branch := &items.Branch{}
				branchFormModel := newBranchFormModel(branch, &m)
				return branchFormModel, nil

			case key.Matches(msg, m.keys.deleteBranch):
				if m.list.SelectedItem() != nil {
					if m.list.SelectedItem().(*items.Branch).Title() == "main" ||
						m.list.SelectedItem().(*items.Branch).Title() == "done" ||
						m.list.SelectedItem().(*items.Branch).Title() == "inProgress" ||
						m.list.SelectedItem().(*items.Branch).Title() == "onHold" {
						return m, m.list.NewStatusMessage(statusMessageStyleRed("This branch is protected"))
					}

					m.mode = modeConfirmDelete
				}
				return m, nil

			case key.Matches(msg, m.keys.chooseBranch):
				return m, nil
			}
		}
	}

	newListModel, cmd := m.list.Update(msg)
	m.list = newListModel
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m branchListModel) View() string {
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
		selected := m.list.SelectedItem().(*items.Branch)

		boxContent := fmt.Sprintf("Delete \"%s\"?\n\n[y] Yes   [n] No", selected.Title())

		leftColumn := appStyle.Render(m.list.View())
		rightColumn := promptBoxStyle.Render(boxContent)

		return lipgloss.JoinHorizontal(lipgloss.Top, leftColumn, rightColumn)
	}

	return appStyle.Render(m.list.View())
}
