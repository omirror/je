package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"git.mills.io/prologic/je/client"
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
		quiet := viper.GetBool("quiet")
		interactive := viper.GetBool("interactive")

		client := client.NewClient(uri, nil)

		if interactive {
			start(client, args[0], args[1:], os.Stdin, quiet)
		} else {
			start(client, args[0], args[1:], nil, quiet)
		}
	},
}

func init() {
	RootCmd.AddCommand(startCmd)

	startCmd.Flags().BoolP(
		"interactive", "i", false,
		"Pass stdin as input to job",
	)
	viper.BindPFlag("interactive", startCmd.Flags().Lookup("interactive"))
	viper.SetDefault("interactive", false)

	startCmd.Flags().BoolP(
		"quiet", "q", false,
		"Only display numeric IDs",
	)
	viper.BindPFlag("quiet", startCmd.Flags().Lookup("quiet"))
	viper.SetDefault("quiet", false)
}

func start(client *client.Client, name string, args []string, input io.Reader, quiet bool) {
	res, err := client.Start(name, args, input)
	if err != nil {
		log.Errorf("error running job %s: %s", name, err)
		return
	}

	if quiet {
		fmt.Print(res[0].ID)
	} else {
		out, err := json.Marshal(res)
		if err != nil {
			log.Errorf("error encoding job result: %s", err)
			return
		}
		fmt.Printf(string(out))
	}
}
