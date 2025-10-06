package cmd

import (
	"fmt"

	"github.com/handlebargh/yatto/internal/version"
	"github.com/spf13/cobra"
)

var versionFlag bool

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print application version",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Println(version.Header())
		fmt.Println(version.Info())
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&versionFlag, "version", "v", false, "Print application version")
	rootCmd.AddCommand(versionCmd)
}
