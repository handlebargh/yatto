package cmd

import (
	"fmt"
	"os"
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
	Run: func(_ *cobra.Command, _ []string) {
		home, err := os.UserHomeDir()
		if err != nil {
			panic(fmt.Errorf("fatal error getting user home directory: %w", err))
		}

		config.InitConfig(home, &configPath)

		printTaskList(printProjects, printRegex)
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
