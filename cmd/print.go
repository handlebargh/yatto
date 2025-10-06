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

package cmd

import (
	"strings"

	"github.com/handlebargh/yatto/internal/config"
	"github.com/handlebargh/yatto/internal/printer"
	"github.com/spf13/cobra"
)

var (
	pullFlag      bool
	printProjects string
	printRegex    string
)

var printCmd = &cobra.Command{
	Use:   "print",
	Short: "Print tasks to stdout",
	RunE: func(_ *cobra.Command, _ []string) error {
		config.InitConfig(homePath, &configPath)
		printTaskList(printProjects, printRegex)

		return nil
	},
}

// printTaskList prints a list of tasks based on the provided project names
// and a regular expression filter.
//
// The function takes two strings as input:
// - printProjects: a space-separated string of project names.
// - printRegex: a regular expression used to filter tasks.
//
// It splits the printProjects string into individual project names, then calls
// printer.PrintTasks with the regex and the list of projects.
func printTaskList(printProjects, printRegex string) {
	// Get a slice of strings from user input.
	projects := strings.Fields(printProjects)

	printer.PrintTasks(printRegex, projects...)
}

func init() {
	printCmd.Flags().BoolVarP(&pullFlag, "pull", "p", false, "Pull the remote before printing")
	printCmd.Flags().StringVarP(&printProjects, "projects", "P", "", "List of project UUIDs to print from")
	printCmd.Flags().StringVarP(&printRegex, "regex", "r", "", "Regex to be used on task labels")
	rootCmd.AddCommand(printCmd)
}
