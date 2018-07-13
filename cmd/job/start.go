package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/prologic/je/client"
)

// startCmd represents the run command
var startCmd = &cobra.Command{
	Use:     "start [flags] <name> [--] [args]",
	Aliases: []string{"exec"},
	Short:   "Starts the given job type by name",
	Long: `This starts the job given by the provided name argument
asyncrhonsly. This does not wait for the job to complete and in contrast to
run returns immediately. Arguments to the job can be provided but if those
arguments are themselves command-line options to an execute use -- [args].

Input can also be provided to the job by using the -i/--interactive option
to pass stadard input to the job.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		uri := viper.GetString("uri")
		client := client.NewClient(uri, nil)

		interactive, err := cmd.Flags().GetBool("interactive")
		if err != nil {
			log.Errorf("error getting -i/--interactive flag: %s", err)
			os.Exit(1)
		}

		quiet, err := cmd.Flags().GetBool("quiet")
		if err != nil {
			log.Errorf("error getting -q/--quiet flag: %s", err)
			os.Exit(1)
		}

		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			os.Exit(start(client, args[0], args[1:], os.Stdin, interactive, quiet))
		} else {
			os.Exit(start(client, args[0], args[1:], nil, interactive, quiet))
		}
	},
}

func init() {
	RootCmd.AddCommand(startCmd)

	startCmd.Flags().BoolP(
		"interactive", "i", false,
		"Keep stdin open",
	)

	startCmd.Flags().BoolP(
		"quiet", "q", false,
		"Only display numeric IDs",
	)
}

func start(client *client.Client, name string, args []string, input io.Reader, interactive, quiet bool) int {
	res, err := client.Create(name, args, input, interactive, false)
	if err != nil {
		log.Errorf("error running job %s: %s", name, err)
		return 1
	}

	if quiet {
		fmt.Print(res[0].ID)
	} else {
		out, err := json.Marshal(res)
		if err != nil {
			log.Errorf("error encoding job result: %s", err)
			return 1
		}
		fmt.Printf(string(out))
	}

	return 0
}
