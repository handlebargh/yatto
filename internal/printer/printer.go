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

// package printer provides the logic to print task lists
// in a non-interactive way to stdout.
package printer

import (
	"cmp"
	"fmt"
	"slices"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/handlebargh/yatto/internal/colors"
	"github.com/handlebargh/yatto/internal/helpers"
	"github.com/handlebargh/yatto/internal/items"
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
func getProjectTasks(projectsIDs ...string) ([]projectTask, []string) {
	projects := helpers.ReadProjectsFromFS()

	foundIDs := make(map[string]bool)
	var result []projectTask

	for _, project := range projects {
		id := project.Id()
		if len(projectsIDs) == 0 || slices.Contains(projectsIDs, id) {
			foundIDs[id] = true
			for _, task := range project.ReadTasksFromFS() {
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
func sortTasks(tasks []projectTask) {
	slices.SortStableFunc(tasks, func(x, y projectTask) int {
		for _, key := range []string{"state", "dueDate", "priority"} {
			switch key {
			case "priority":
				// Higher number = higher priority
				if cmp := cmp.Compare(y.task.PriorityValue(), x.task.PriorityValue()); cmp != 0 {
					return cmp
				}

			case "dueDate":
				dx, dy := x.task.DueDate(), y.task.DueDate()
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
				// In-progress before others
				if x.task.InProgress() && !y.task.InProgress() {
					return -1
				}
				if !x.task.InProgress() && y.task.InProgress() {
					return 1
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
func PrintTasks(projectsIDs ...string) {
	pt, missing := getProjectTasks(projectsIDs...)

	if len(missing) > 0 {
		for _, projectId := range missing {
			fmt.Println(
				lipgloss.NewStyle().
					Foreground(colors.Red).
					Render(fmt.Sprintf("\nerror: project ID %s not found\n", projectId)),
			)
		}
	}

	var pendingTasks []projectTask
	for _, pt := range pt {
		if !pt.task.Completed() {
			pendingTasks = append(pendingTasks, pt)
		}
	}

	sortTasks(pendingTasks)

	for _, pt := range pendingTasks {
		taskTitle := pt.task.CropTaskTitle(40)
		projectTitle := lipgloss.NewStyle().
			Foreground(helpers.GetColorCode(pt.project.Color())).
			Render(pt.project.Title())
		taskPriority := pt.task.Priority()

		var taskLabels string
		if len(pt.task.Labels()) > 0 {
			taskLabels = lipgloss.NewStyle().
				Foreground(colors.Blue).
				Render("\n  " + pt.task.CropTaskLabels(40))
		}

		left := lipgloss.NewStyle().
			Width(50).
			Render(
				"\n  " +
					taskTitle + "\n  " +
					projectTitle +
					taskLabels)

		priorityValueStyle := lipgloss.NewStyle().
			Foreground(colors.Black).
			Padding(0, 1)

		switch pt.task.Priority() {
		case "low":
			priorityValueStyle = priorityValueStyle.Background(colors.Indigo)
		case "medium":
			priorityValueStyle = priorityValueStyle.Background(colors.Orange)
		case "high":
			priorityValueStyle = priorityValueStyle.Background(colors.Red)
		}

		right := lipgloss.NewStyle().Render(
			"\n" + priorityValueStyle.Render(taskPriority),
		)

		now := time.Now()
		dueDate := pt.task.DueDate()

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

		if pt.task.InProgress() {
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
				Render("due in "+pt.task.DaysUntilToString()+" day(s)")
		}

		row := lipgloss.JoinHorizontal(lipgloss.Top, left, right)

		fmt.Println(row)
	}
}
