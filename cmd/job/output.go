package main

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"git.mills.io/prologic/je/client"
)

// outputCmd represents the run command
var outputCmd = &cobra.Command{
	Use:     "output [flags] <id>",
	Aliases: []string{"out"},
	Short:   "Retrieve and display job output",
	Long: `This retrieves and views the job's output. That is the standard
output of the job.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		uri := viper.GetString("uri")
		client := client.NewClient(uri, nil)

		id := args[0]

		output(client, id)
	},
}

func init() {
	RootCmd.AddCommand(outputCmd)
}

func output(client *client.Client, id string) int {
	res, err := client.GetJobByID(id)
	if err != nil {
		log.Errorf("error retrieving information for job #%s: %s", id, err)
		return 1
	}

	fmt.Print(res[0].Output)

	return 0
}
