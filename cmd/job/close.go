package main

import (
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/prologic/je"
)

// closeCmd represents the run command
var closeCmd = &cobra.Command{
	Use:     "close [flags] <id>",
	Aliases: []string{"done"},
	Short:   "Close the input of an interactive job",
	Long: `This close the input stream of an interactive job.

Normally jobs are not interactive meaning that they're stdin pipes are closed
once written to from their input. You can keep the stdin of a job open
indefiniately by using the -i/--interactive options of the run and start
commands. With this option you can then write further input to the job(s)
with the write command. When finished you can explicitly close the stream
by using the close command.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		uri := viper.GetString("uri")
		client := je.NewClient(uri, nil)

		id := args[0]

		os.Exit(close(client, id))
	},
}

func init() {
	RootCmd.AddCommand(closeCmd)
}

func close(client *je.Client, id string) int {
	err := client.Close(id)
	if err != nil {
		log.Errorf("error writing to job #%s: %s", id, err)
		return 1
	}

	return 0
}
