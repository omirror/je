package main

import (
	"encoding/json"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"git.mills.io/prologic/je/client"
)

// infoCmd represents the run command
var infoCmd = &cobra.Command{
	Use:     "info [flags] <id>",
	Aliases: []string{"view"},
	Short:   "View information about a job",
	Long:    `This retrieves and view information about a job`,
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		uri := viper.GetString("uri")
		client := client.NewClient(uri, nil)

		id := args[0]

		os.Exit(info(client, id))
	},
}

func init() {
	RootCmd.AddCommand(infoCmd)
}

func info(c *client.Client, id string) int {
	res, err := c.GetJobByID(id)
	if err != nil {
		log.Errorf("error retrieving information for job #%s: %s", id, err)
		return 1
	}

	out, err := json.Marshal(res)
	if err != nil {
		log.Errorf("error encoding job results: %s", err)
		return 1
	}

	fmt.Print(string(out))

	return 0
}
