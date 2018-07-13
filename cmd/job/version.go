package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/prologic/je"
)

// versionCmd represents the run command
var versionCmd = &cobra.Command{
	Use:     "version",
	Aliases: []string{},
	Short:   "Display the version",
	Long:    `This display the version of je`,
	Run: func(cmd *cobra.Command, args []string) {
		os.Exit(version())
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}

func version() int {
	fmt.Printf("je v%s", je.FullVersion())
	return 0
}
