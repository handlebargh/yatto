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

// Package staticprinter provides the logic to print task lists
// in a non-interactive way to stdout.
package staticprinter

import (
	"cmp"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/handlebargh/yatto/internal/colors"
	"github.com/handlebargh/yatto/internal/helpers"
	"github.com/handlebargh/yatto/internal/items"
	"github.com/handlebargh/yatto/internal/vcs"
	"github.com/spf13/viper"
)

// projectTask represents a single task along with the project it belongs to.
// It is used to keep project context when working with individual tasks.
type projectTask struct {
	project items.Project
	task    items.Task
}

// getProjectTasks retrieves tasks from the filesystem for the given project IDs.
//
// If no project IDs are provided, it returns tasks from all available projects.
// For each task, the associated project is also returned via the projectTask type.
//
// It returns two values:
//   - A slice of projectTask, each containing a task and its corresponding project.
//   - A slice of strings representing project IDs that were requested but not found.
func getProjectTasks(v *viper.Viper, projectsIDs ...string) ([]projectTask, []string) {
	projects := helpers.ReadProjectsFromFS(v)

	foundIDs := make(map[string]bool)
	var result []projectTask

	for _, project := range projects {
		id := project.ID
		if len(projectsIDs) == 0 || slices.Contains(projectsIDs, id) {
			foundIDs[id] = true
			for _, task := range project.ReadTasksFromFS(v) {
				result = append(result, projectTask{
					project: project,
					task:    task,
				})
			}
		}
	}

	var missing []string
	if len(projectsIDs) > 0 {
		for _, id := range projectsIDs {
			if !foundIDs[id] {
				missing = append(missing, id)
			}
		}
	}

	return result, missing
}

// sortTasks sorts a slice of projectTask items in-place using a stable sort,
// applying a multi-level comparison based on task state, due date, and priority.
//
// The sorting precedence is as follows:
//  1. State: Tasks that are in progress are ordered before those that are not.
//  2. Due Date: Tasks with earlier due dates come before later ones. Tasks with a due date
//     are prioritized over tasks without one.
//  3. Priority: Tasks with higher numeric priority values are ranked higher.
//
// The sort is stable, preserving the relative order of equal elements across criteria.
func sortTasks(v *viper.Viper, tasks []projectTask) {
	me, _ := vcs.User(v)

	slices.SortStableFunc(tasks, func(x, y projectTask) int {
		for _, key := range []string{"state", "assignee", "dueDate", "priority"} {
			switch key {
			case "state":
				// In-progress before others
				if x.task.InProgress && !y.task.InProgress {
					return -1
				}
				if !x.task.InProgress && y.task.InProgress {
					return 1
				}

			case "assignee":
				switch {
				case x.task.Assignee == "" && y.task.Assignee != "":
					return 1
				case x.task.Assignee != "" && y.task.Assignee == "":
					return -1
				case x.task.Assignee == me && y.task.Assignee != me:
					return -1
				case x.task.Assignee != me && y.task.Assignee == me:
					return 1
				default:
					return strings.Compare(strings.ToLower(x.task.Assignee), strings.ToLower(y.task.Assignee))
				}

			case "priority":
				// Higher number = higher priority
				if compare := cmp.Compare(y.task.PriorityValue(), x.task.PriorityValue()); compare != 0 {
					return compare
				}

			case "dueDate":
				dx, dy := x.task.DueDate, y.task.DueDate
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
			}
		}
		return 0
	})
}

// PrintTasks displays a styled list of all non-completed tasks for the given project IDs.
//
// For each provided project ID, it attempts to retrieve associated tasks. If any project IDs
// are not found, an error message is printed for each.
//
// The remaining tasks are filtered to exclude completed ones, then sorted by in-progress state,
// due date, and priority using sortTasks. Each task is printed with:
//   - A cropped task title
//   - The project title, color-coded
//   - Optional labels, color-coded
//   - Priority, styled by level (low, medium, high)
//   - Badges indicating task state, including:
//   - "due today", "overdue", "in progress", or "due in N day(s)"
func PrintTasks(v *viper.Viper, labelRegex string, author, assignee bool, projectsIDs ...string) {
	projTask, missing := getProjectTasks(v, projectsIDs...)

	if len(missing) > 0 {
		for _, projectID := range missing {
			fmt.Println(
				lipgloss.NewStyle().
					Foreground(colors.Red()).
					Render(fmt.Sprintf("\nerror: project ID %s not found\n", projectID)),
			)
		}
	}

	me, _ := vcs.User(v)
	regex := regexp.MustCompile(labelRegex)

	var pendingTasks []projectTask
	for _, pt := range projTask {
		if !pt.task.Completed && regex.MatchString(pt.task.Labels) {
			switch {
			case author && pt.task.Author == me:
				pendingTasks = append(pendingTasks, pt)
			case assignee && pt.task.Assignee == me:
				pendingTasks = append(pendingTasks, pt)
			case !author && !assignee:
				pendingTasks = append(pendingTasks, pt)
			default:
				break
			}
		}
	}

	sortTasks(v, pendingTasks)

	if len(pendingTasks) == 0 {
		fmt.Println(
			lipgloss.NewStyle().
				Foreground(colors.Green()).
				Render("yatto: No open tasks found"),
		)
	}

	for _, pt := range pendingTasks {
		taskTitle := pt.task.CropTaskTitle(40)
		projectTitle := lipgloss.NewStyle().
			Foreground(helpers.GetColorCode(pt.project.Color)).
			Render(pt.project.Title)
		taskPriority := pt.task.Priority

		var left strings.Builder

		left.WriteString("\n")
		left.WriteString(lipgloss.NewStyle().Width(50).Render(taskTitle))
		left.WriteString("\n")
		left.WriteString(lipgloss.NewStyle().Width(50).Render(projectTitle))
		left.WriteString("\n")
		left.WriteString(lipgloss.NewStyle().Width(50).Foreground(colors.Blue()).Render(pt.task.CropTaskLabels(40)))

		if v.GetBool("author.show_printer") {
			left.WriteString("\n")
			left.WriteString(lipgloss.NewStyle().Foreground(colors.Green()).Render("Author: "))
			left.WriteString(pt.task.Author)
		}

		me, _ := vcs.User(v)
		if v.GetBool("assignee.show_printer") {
			left.WriteString("\n")
			left.WriteString(lipgloss.NewStyle().Foreground(colors.Orange()).Render("Assignee: "))
			if pt.task.Assignee == me {
				left.WriteString(lipgloss.NewStyle().Foreground(colors.Red()).Render(pt.task.Assignee))
			} else {
				left.WriteString(pt.task.Assignee)
			}
		}

		priorityValueStyle := lipgloss.NewStyle().
			Foreground(colors.BadgeText()).
			Padding(0, 1)

		switch pt.task.Priority {
		case "low":
			priorityValueStyle = priorityValueStyle.Background(colors.Indigo())
		case "medium":
			priorityValueStyle = priorityValueStyle.Background(colors.Orange())
		case "high":
			priorityValueStyle = priorityValueStyle.Background(colors.Red())
		}

		var right strings.Builder

		right.WriteString("\n")
		right.WriteString(priorityValueStyle.Render(taskPriority))

		now := time.Now()
		dueDate := pt.task.DueDate

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

		if pt.task.InProgress {
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
				Render("due in " + pt.task.DaysUntilToString() + " day(s)"))
		}

		row := lipgloss.JoinHorizontal(lipgloss.Top, left.String(), right.String())

		fmt.Println(row)
	}
}
