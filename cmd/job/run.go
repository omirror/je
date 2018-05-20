package main

import (
	"encoding/json"
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"git.mills.io/prologic/je/client"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:     "run [flags] <name>",
	Aliases: []string{"exec"},
	Short:   "Runs the given job type by name",
	Long: `This runs the job given by the provided name argument and waits
for it to complete before returning and printing the result of the job and
its output log.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		uri := viper.GetString("uri")
		client := client.NewClient(uri, nil)

		name := args[0]

		run(client, name)
	},
}

func init() {
	RootCmd.AddCommand(runCmd)
}

func run(client *client.Client, name string) {
	res, err := client.Run(name)
	if err != nil {
		log.Errorf("error running job %s: %s", name, err)
		return
	}

	out, err := json.Marshal(res)
	if err != nil {
		log.Errorf("error encoding job result: %s", err)
		return
	}

	fmt.Print(string(out))
}
