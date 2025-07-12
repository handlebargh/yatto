package models

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/lipgloss"
	"github.com/handlebargh/yatto/internal/git"
	"github.com/handlebargh/yatto/internal/items"
	"github.com/spf13/viper"

	tea "github.com/charmbracelet/bubbletea"
)

type branchListKeyMap struct {
	toggleHelpMenu key.Binding
	addBranch      key.Binding
	checkoutBranch key.Binding
	deleteBranch   key.Binding
	chooseBranch   key.Binding
}

func newBranchListKeyMap() *branchListKeyMap {
	return &branchListKeyMap{
		chooseBranch: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "branch info"),
		),
		deleteBranch: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete branch"),
		),
		checkoutBranch: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "checkout branch"),
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
	list             list.Model
	listModel        *taskListModel
	keys             *branchListKeyMap
	selected         bool
	selection        *items.Branch
	mode             mode
	err              error
	progress         progress.Model
	progressDone     bool
	waitingAfterDone bool
	status           string
	width            int
	height           int
}

func InitialBranchListModel(listModel *taskListModel) branchListModel {
	listKeys := newBranchListKeyMap()

	branches, _, err := git.GetBranches()
	if err != nil {
		panic(err)
	}

	listItems := []list.Item{}

	for _, branch := range branches {
		listItems = append(listItems, &branch)
	}

	itemList := list.New(listItems, list.NewDefaultDelegate(), 0, 0)
	itemList.SetShowPagination(true)
	itemList.SetShowTitle(true)
	itemList.SetShowStatusBar(true)
	itemList.Title = "Branches"
	itemList.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			listKeys.toggleHelpMenu,
			listKeys.addBranch,
			listKeys.checkoutBranch,
			listKeys.deleteBranch,
			listKeys.chooseBranch,
		}
	}

	return branchListModel{
		list:      itemList,
		listModel: listModel,
		selected:  false,
		keys:      listKeys,
		progress:  progress.New(progress.WithDefaultGradient()),
	}
}

func (m branchListModel) Init() tea.Cmd {
	return tickCmd()
}

func (m branchListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tickMsg:
		if m.progress.Percent() >= 1.0 && !m.waitingAfterDone {
			m.progressDone = true
			m.waitingAfterDone = true

			// Return a timer command to keep displaying 100% progress
			// for one second.
			return m, tea.Tick(time.Second, func(t time.Time) tea.Msg {
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
		m.status = fmt.Sprintf("🗸  New branch %s added", msg.Branch.Title())
		m.progressDone = true
		return m, tea.Batch(m.progress.SetPercent(1.0), m.list.InsertItem(0, &msg.Branch))

	case git.DeleteBranchDoneMsg:
		m.status = fmt.Sprintf("🗑  branch %s deleted", msg.Branch.Title())
		m.progressDone = true
		m.list.RemoveItem(m.list.GlobalIndex())
		return m, m.progress.SetPercent(1.0)

	case git.CheckoutBranchDoneMsg:
		m.status = fmt.Sprintf("🗸  branch %s checked out", msg.Branch.Title())
		m.progressDone = true
		return m, m.progress.SetPercent(1.0)

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
				m.mode = modeNormal
				if m.list.SelectedItem() != nil {
					return m, tea.Batch(
						m.progress.SetPercent(0.1),
						tickCmd(),
						git.DeleteBranchCmd(*m.list.SelectedItem().(*items.Branch)))
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

			case key.Matches(msg, m.keys.checkoutBranch):
				if m.list.SelectedItem() != nil {
					return m, tea.Batch(
						m.progress.SetPercent(0.1),
						tickCmd(),
						git.CheckoutBranchCmd(*m.list.SelectedItem().(*items.Branch)))
				}
				return m, nil

			case key.Matches(msg, m.keys.addBranch):
				branch := &items.Branch{}
				branchFormModel := newBranchFormModel(branch, &m)
				return branchFormModel, nil

			case key.Matches(msg, m.keys.deleteBranch):
				if m.list.SelectedItem() != nil {
					if m.list.SelectedItem().(*items.Branch).Title() == viper.GetString("git.default_branch") ||
						m.list.SelectedItem().(*items.Branch).Title() == "done" ||
						m.list.SelectedItem().(*items.Branch).Title() == "inProgress" ||
						m.list.SelectedItem().(*items.Branch).Title() == "onHold" {
						return m, m.list.NewStatusMessage(statusMessageStyleRed("This branch is protected"))
					}

					m.mode = modeConfirmDelete
				}
				return m, nil

			case key.Matches(msg, m.keys.chooseBranch):
				if m.list.SelectedItem() != nil {
					m.selected = true
					m.selection = m.list.SelectedItem().(*items.Branch)
				}
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
	// Progress bar styling
	progressStyle := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center)

	// Display progress bar at 100%
	if m.progressDone && m.waitingAfterDone {
		return progressStyle.Render(m.status + "\n\n" + m.progress.ViewAs(1.0))
	}

	// Display progress bar if not at 0%
	if m.progress.Percent() != 0.0 {
		return progressStyle.Render(m.status + "\n\n" + m.progress.View())
	}

	// Display deletion confirm view.
	if m.mode == modeConfirmDelete {
		selected := m.list.SelectedItem().(*items.Branch)

		boxContent := fmt.Sprintf("Delete \"%s\"?\n\n[y] Yes   [n] No", selected.Title())

		leftColumn := appStyle.Render(m.list.View())
		rightColumn := promptBoxStyle.Render(boxContent)

		return lipgloss.JoinHorizontal(lipgloss.Top, leftColumn, rightColumn)
	}

	// Display list view.
	if !m.selected {
		return appStyle.Render(m.list.View())
	}

	// Display branch view
	leftColumn := appStyle.Render(m.list.View())
	rightColumn := detailBoxStyle.Render("hello")

	return lipgloss.JoinHorizontal(lipgloss.Top, leftColumn, rightColumn)
}
