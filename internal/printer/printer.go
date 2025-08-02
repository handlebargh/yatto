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
	"fmt"
	"slices"

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

// PrintTasks prints tasks to the terminal in a stylized, non-interactive format.
//
// If project IDs are provided, it only prints tasks from those projects.
// If no project IDs are given, it prints tasks from all projects.
//
// If any provided project ID does not exist, an error message is printed to the terminal
// indicating which project IDs were not found.
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

	for _, pt := range pt {

		taskTitle := pt.task.CropTaskTitle(30)
		projectTitle := lipgloss.NewStyle().
			Foreground(helpers.GetColorCode(pt.project.Color())).
			Render(pt.project.Title())

		var taskLabels string
		if len(pt.task.Labels()) > 0 {
			taskLabels = lipgloss.NewStyle().
				Foreground(colors.Blue).
				Render("\n\t" + pt.task.CropTaskLabels(30))
		}

		left := lipgloss.NewStyle().Render(
			"\n\t" +
				taskTitle + "\n\t" +
				projectTitle +
				taskLabels)

		right := lipgloss.NewStyle().Render()

		row := lipgloss.JoinHorizontal(lipgloss.Top, left, right)

		fmt.Println(row)
	}
}
