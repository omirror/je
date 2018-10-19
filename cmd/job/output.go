package main

import (
	"io"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/prologic/je/client"
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

		follow, err := cmd.Flags().GetBool("follow")
		if err != nil {
			log.Errorf("error getting -f/--follow flag: %s", err)
			os.Exit(1)
		}

		os.Exit(output(client, id, follow))
	},
}

func init() {
	RootCmd.AddCommand(outputCmd)

	outputCmd.Flags().BoolP(
		"follow", "f", false,
		"Follow output as it is written to",
	)
}

func output(client *client.Client, id string, follow bool) int {
	r, err := client.Output(id, follow)
	if err != nil {
		log.Errorf("error retrieving logs for job %s: %s", id, err)
		return 1
	}

	io.Copy(os.Stdout, r)

	return 0
}
