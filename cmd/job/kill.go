package main

import (
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/prologic/je"
)

// killCmd represents the run command
var killCmd = &cobra.Command{
	Use:     "kill [flags] <id>",
	Aliases: []string{"stop"},
	Short:   "Stop the given job",
	Long: `This stops the given job gracefully by sending the job SIGINT.
You can also forcibly kill the job with -f/--force which will use SIGKILL
instead forcing the job to terminate uncleanly.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		uri := viper.GetString("uri")
		client := je.NewClient(uri, nil)

		id := args[0]

		force, err := cmd.Flags().GetBool("force")
		if err != nil {
			log.Errorf("error getting -f/--force flag: %s", err)
			os.Exit(1)
		}

		os.Exit(kill(client, id, force))
	},
}

func init() {
	RootCmd.AddCommand(killCmd)

	killCmd.Flags().BoolP(
		"force", "f", false,
		"Force kill job by sending SIGKILL",
	)
}

func kill(c *je.Client, id string, force bool) int {
	err := c.Kill(id, force)
	if err != nil {
		log.Errorf("error retrieving information for job #%s: %s", id, err)
		return 1
	}
	return 0
}
