package main

import (
	"io"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"git.mills.io/prologic/je/client"
)

// writeCmd represents the run command
var writeCmd = &cobra.Command{
	Use:     "write [flags] <id>",
	Aliases: []string{"send"},
	Short:   "Writes data to an interactive job",
	Long: `This writes data from standard input to a running interactive job.

Normally jobs are not interactive meaning that they're stdin pipes are closed
once written to from their input. You can keep the stdin of a job open
indefiniately by using the -i/--interactive options of the run and start
commands. With this option you can then write further input to the job(s)
with the write command.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		uri := viper.GetString("uri")
		client := client.NewClient(uri, nil)

		id := args[0]

		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			os.Exit(write(client, id, os.Stdin))
		} else {
			log.Errorf("stdin is not a pipe")
			os.Exit(1)
		}
	},
}

func init() {
	RootCmd.AddCommand(writeCmd)
}

func write(client *client.Client, id string, input io.Reader) int {
	err := client.Write(id, input)
	if err != nil {
		log.Errorf("error writing to job #%s: %s", id, err)
		return 1
	}

	return 0
}
