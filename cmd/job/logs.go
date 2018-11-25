package main

import (
	"io"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/prologic/je"
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
		client := je.NewClient(uri, nil)

		id := args[0]

		follow, err := cmd.Flags().GetBool("follow")
		if err != nil {
			log.Errorf("error getting -f/--follow flag: %s", err)
			os.Exit(1)
		}

		os.Exit(logs(client, id, follow))
	},
}

func init() {
	RootCmd.AddCommand(logsCmd)

	logsCmd.Flags().BoolP(
		"follow", "f", false,
		"Follow logs as it is written to",
	)

}

func logs(client *je.Client, id string, follow bool) int {
	r, err := client.Logs(id, follow)
	if err != nil {
		log.Errorf("error retrieving logs for job %s: %s", id, err)
		return 1
	}

	io.Copy(os.Stdout, r)

	return 0
}
