package main

import (
	"io"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"git.mills.io/prologic/je/client"
)

// logsCmd represents the run command
var logsCmd = &cobra.Command{
	Use:     "logs [flags] <id>",
	Aliases: []string{"log"},
	Short:   "Retrieves logs for a job",
	Long:    `This retrives and display the logs for the job given by id`,
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		uri := viper.GetString("uri")
		client := client.NewClient(uri, nil)

		id := args[0]

		os.Exit(logs(client, id))
	},
}

func init() {
	RootCmd.AddCommand(logsCmd)
}

func logs(client *client.Client, id string) int {
	r, err := client.Logs(id)
	if err != nil {
		log.Errorf("error retrieving logs for job %s: %s", id, err)
		return 1
	}

	io.Copy(os.Stdout, r)

	return 0
}
